package dao

import (
	"fmt"
	"log"
	"time"

	"go-admin/config"
	"go-admin/model"
)

// plainIndexes 三种方言通用（等值/范围查询用的 B-Tree 索引）
// 说明：历史表只有主键与少数唯一索引，列表页按 status / 父节点 / 时间过滤、按时间排序时
// 会退化为顺序扫描，这里补齐高频过滤列。
var plainIndexes = []struct{ Name, Table, Column string }{
	{"idx_sys_oper_log_status", "sys_oper_log", "status"},
	{"idx_sys_job_log_create_time", "sys_job_log", "create_time"},
	{"idx_sys_user_dept_id", "sys_user", "dept_id"},
	{"idx_sys_dept_parent_id", "sys_dept", "parent_id"},
	{"idx_sys_menu_parent_id", "sys_menu", "parent_id"},
	{"idx_sys_logininfor_username", "sys_logininfor", "username"},
}

// pgTrgmIndexes PostgreSQL 三元组索引：让 LIKE '%关键词%' 这类前后模糊查询能用上索引
// 依赖 pg_trgm 扩展；扩展不可用时仅跳过这几项，不影响运行
var pgTrgmIndexes = []struct{ Name, Table, Column string }{
	{"idx_sys_oper_log_title_trgm", "sys_oper_log", "title"},
	{"idx_sys_oper_log_name_trgm", "sys_oper_log", "oper_name"},
	{"idx_sys_logininfor_username_trgm", "sys_logininfor", "username"},
	{"idx_sys_job_log_name_trgm", "sys_job_log", "job_name"},
}

// ensureIndexes 确保性能相关索引存在（幂等），并在 PostgreSQL 上刷新统计信息
func ensureIndexes() error {
	switch Dialect() {
	case "postgres":
		for _, ix := range plainIndexes {
			stmt := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)", ix.Name, ix.Table, ix.Column)
			if err := DB.Exec(stmt).Error; err != nil {
				return fmt.Errorf("创建索引 %s 失败: %w", ix.Name, err)
			}
		}
		hasTrgm, err := ensurePgTrgm()
		if err != nil {
			return err
		}
		if hasTrgm {
			for _, ix := range pgTrgmIndexes {
				stmt := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s USING gin (%s gin_trgm_ops)", ix.Name, ix.Table, ix.Column)
				if err := DB.Exec(stmt).Error; err != nil {
					return fmt.Errorf("创建三元组索引 %s 失败: %w", ix.Name, err)
				}
			}
		}
		// 迁移/批量导入后 pg_stat 的统计行数可能仍为 0（如 sys_user 显示 0 行），
		// 会让优化器误判走全表扫描，这里对热点表重新采样
		if err := DB.Exec("ANALYZE sys_oper_log, sys_logininfor, sys_job_log, sys_user, sys_dept, sys_menu, sys_api, sys_gen_table").Error; err != nil {
			log.Printf("⚠️ ANALYZE 失败（不影响功能）: %v", err)
		}
	case "mysql":
		for _, ix := range plainIndexes {
			if mysqlIndexExists(ix.Table, ix.Name) {
				continue
			}
			stmt := fmt.Sprintf("ALTER TABLE %s ADD INDEX %s (%s)", ix.Table, ix.Name, ix.Column)
			if err := DB.Exec(stmt).Error; err != nil {
				return fmt.Errorf("创建索引 %s 失败: %w", ix.Name, err)
			}
		}
	default: // sqlite：本地库数据量小，仍然补上通用索引
		for _, ix := range plainIndexes {
			stmt := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)", ix.Name, ix.Table, ix.Column)
			if err := DB.Exec(stmt).Error; err != nil {
				return fmt.Errorf("创建索引 %s 失败: %w", ix.Name, err)
			}
		}
	}
	return nil
}

// ensurePgTrgm 确保 pg_trgm 扩展可用；返回是否可用
func ensurePgTrgm() (bool, error) {
	var cnt int64
	if err := DB.Raw("SELECT count(*) FROM pg_extension WHERE extname = 'pg_trgm'").Scan(&cnt).Error; err != nil {
		return false, fmt.Errorf("查询 pg_trgm 扩展失败: %w", err)
	}
	if cnt > 0 {
		return true, nil
	}
	if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error; err != nil {
		// 普通租户无 CREATE EXTENSION 权限时只降级：模糊查询退回顺序扫描
		log.Printf("⚠️ 无法创建 pg_trgm 扩展，模糊搜索索引已跳过: %v", err)
		return false, nil
	}
	return true, nil
}

// mysqlIndexExists MySQL 不支持 CREATE INDEX IF NOT EXISTS，先查 information_schema
func mysqlIndexExists(table, index string) bool {
	var cnt int64
	err := DB.Raw(`SELECT count(*) FROM information_schema.statistics
		WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?`, table, index).Scan(&cnt).Error
	return err == nil && cnt > 0
}

// StartLogRetention 按配置清理过期日志（database 之外的 logs.retentionDays，0 表示不清理）
// 启动时先清一次，之后每 24 小时执行一次
func StartLogRetention(stop <-chan struct{}) {
	days := config.Cfg.Logs.RetentionDays
	if days <= 0 {
		return
	}
	backgroundWG.Add(1)
	go func() {
		defer backgroundWG.Done()
		purgeLogsOnce(days)
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				purgeLogsOnce(days)
			}
		}
	}()
}

// PurgeOldLogs 删除早于 days 天的操作日志/登录日志/任务日志，返回删除行数
func PurgeOldLogs(days int) (int64, error) {
	if days <= 0 {
		return 0, nil
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	var total int64
	for _, item := range []struct {
		Model  interface{}
		Column string
	}{
		{&model.OperLog{}, "oper_time"},
		{&model.LoginLog{}, "login_time"},
		{&model.JobLog{}, "create_time"},
	} {
		res := DB.Where(item.Column+" < ?", cutoff).Delete(item.Model)
		if res.Error != nil {
			return total, fmt.Errorf("清理 %T 失败: %w", item.Model, res.Error)
		}
		total += res.RowsAffected
	}
	return total, nil
}

func purgeLogsOnce(days int) {
	n, err := PurgeOldLogs(days)
	if err != nil {
		log.Printf("⚠️ 日志清理失败: %v", err)
		return
	}
	if n > 0 {
		log.Printf("日志清理完成：删除 %d 条早于 %d 天前的记录", n, days)
	}
}
