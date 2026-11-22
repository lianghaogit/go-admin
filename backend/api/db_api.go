package api

import (
	"database/sql"
	"os"
	"strings"

	"go-admin/dao"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// DbTableRow 表统计
type DbTableRow struct {
	TableName string `json:"tableName"`
	RowCount  int64  `json:"rowCount"`
}

// listTablesSQL 按方言返回"全部业务表名"查询
func listTablesSQL() string {
	switch dao.Dialect() {
	case "mysql":
		return "SELECT table_name AS name FROM information_schema.tables WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE' ORDER BY table_name"
	case "postgres":
		return "SELECT tablename AS name FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename"
	default:
		return "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name"
	}
}

// sqlIdent 双引号包裹标识符（表名来自数据库元数据，仍做转义防御）
func sqlIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// sqlLiteral 单引号包裹字符串字面量
func sqlLiteral(s string) string {
	return `'` + strings.ReplaceAll(s, `'`, `''`) + `'`
}

// countTablesSQL 生成一条 UNION ALL 语句统计所有表行数：
// 原先逐表 COUNT(*) 需要 N 次往返（26 张表 = 26 次），合并后只往返一次。
func countTablesSQL(names []string) string {
	if len(names) == 0 {
		return ""
	}
	parts := make([]string, 0, len(names))
	for _, n := range names {
		parts = append(parts, "SELECT "+sqlLiteral(n)+" AS name, COUNT(*) AS cnt FROM "+sqlIdent(n))
	}
	return strings.Join(parts, " UNION ALL ")
}

// shortVersion 把驱动的原始版本串收敛成简短版本号，便于前端单行展示
func shortVersion(dialect, raw string) string {
	raw = strings.TrimSpace(raw)
	if dialect == "sqlite" {
		return "SQLite " + raw
	}
	if i := strings.Index(raw, " on "); i > 0 { // PostgreSQL: "PostgreSQL 18.4 (...) on x86_64..."
		raw = raw[:i]
	}
	if i := strings.Index(raw, " ("); i > 0 { // 去掉发行版括号
		raw = raw[:i]
	}
	return raw
}

// DbInfo 数据监控：数据库概览与各表行数（sqlite / mysql / postgres）
func DbInfo(c *gin.Context) {
	dialect := dao.Dialect()
	info := gin.H{"dbType": dialect}

	var rawVersion string
	switch dialect {
	case "mysql":
		_ = dao.DB.Raw("SELECT VERSION()").Scan(&rawVersion).Error
		info["journalMode"] = "InnoDB"
	case "postgres":
		_ = dao.DB.Raw("SELECT version()").Scan(&rawVersion).Error
		// PostgreSQL 固定使用 WAL，这里改为给出运维真正关注的指标
		var dbSize int64
		if err := dao.DB.Raw("SELECT pg_database_size(current_database())").Scan(&dbSize).Error; err == nil {
			info["dbSize"] = dbSize
		}
		var conns int
		if err := dao.DB.Raw("SELECT COUNT(*) FROM pg_stat_activity WHERE datname = current_database()").Scan(&conns).Error; err == nil {
			info["connections"] = conns
		}
		var maxConns int
		if err := dao.DB.Raw("SELECT current_setting('max_connections')::int").Scan(&maxConns).Error; err == nil {
			info["maxConnections"] = maxConns
		}
		// 缓存命中率：命中块 / (命中块 + 读盘块)
		var hit sql.NullFloat64
		_ = dao.DB.Raw("SELECT ROUND(100.0 * SUM(blks_hit) / NULLIF(SUM(blks_hit + blks_read), 0), 2) " +
			"FROM pg_stat_database WHERE datname = current_database()").Scan(&hit).Error
		if hit.Valid {
			info["cacheHitRatio"] = hit.Float64
		}
		// 未提交事务数（长事务/连接泄漏的先兆）
		var longTx int
		if err := dao.DB.Raw("SELECT COUNT(*) FROM pg_stat_activity WHERE datname = current_database() " +
			"AND state = 'idle in transaction'").Scan(&longTx).Error; err == nil {
			info["idleInTransaction"] = longTx
		}
	default:
		// SQLite 版本与 journal 模式
		var journal string
		_ = dao.DB.Raw("SELECT sqlite_version()").Scan(&rawVersion).Error
		if err := dao.DB.Raw("PRAGMA journal_mode").Scan(&journal).Error; err == nil {
			info["journalMode"] = journal
		}
		// 数据库文件大小
		if f, err := os.Stat(dao.DbPath); err == nil {
			info["fileSize"] = f.Size()
		}
	}
	if rawVersion != "" {
		info["version"] = shortVersion(dialect, rawVersion)
		info["versionRaw"] = rawVersion
	}

	// 全部业务表
	var names []string
	if err := dao.DB.Raw(listTablesSQL()).Scan(&names).Error; err != nil {
		utils.Fail(c, 500, "查询表失败")
		return
	}
	rows := make([]DbTableRow, 0, len(names))
	var total int64
	if sqlText := countTablesSQL(names); sqlText != "" {
		var res []struct {
			Name string
			Cnt  int64
		}
		if err := dao.DB.Raw(sqlText).Scan(&res).Error; err == nil {
			cntMap := make(map[string]int64, len(res))
			for _, r := range res {
				cntMap[r.Name] = r.Cnt
			}
			for _, n := range names {
				cnt := cntMap[n]
				rows = append(rows, DbTableRow{TableName: n, RowCount: cnt})
				total += cnt
			}
		} else {
			// 合并查询失败（如方言不支持）时退回逐表统计，保证页面仍可用
			for _, n := range names {
				var cnt int64
				if err := dao.DB.Raw("SELECT COUNT(*) FROM " + sqlIdent(n)).Scan(&cnt).Error; err != nil {
					cnt = -1
				}
				rows = append(rows, DbTableRow{TableName: n, RowCount: cnt})
				total += cnt
			}
		}
	}
	info["tableCount"] = len(rows)
	info["totalRows"] = total
	info["tables"] = rows
	utils.OkData(c, info)
}
