package dao

import (
	"fmt"
	"strings"
)

// dropForeignKeys 删除表结构上的全部外键约束，保持"数据有主外键语义、结构不建约束"的约定。
//
// 背景：项目改由业务代码维护关联（删除用户时清理 user_post/user_role，删除部门时清理
// role_dept，删除流程定义时清理 node，删除生成配置时清理 column），数据库层不再用
// FOREIGN KEY 做级联/限制；这样历史脏数据、跨库导入和 DDL 变更都不会被约束卡住。
//
// 兼容性：
//   - PostgreSQL：ALTER TABLE ... DROP CONSTRAINT，可原地删除；
//   - MySQL：ALTER TABLE ... DROP FOREIGN KEY，可原地删除；
//   - SQLite：外键写在建表 DDL 里，SQLite 不支持原地删除约束；新库由
//     DisableForeignKeyConstraintWhenMigrating 保证不再生成。老库若仍有 FK，
//     这里只统计并提示（需要时可重建库或按表重建）。
//
// 该函数幂等：没有外键时什么也不做。
func dropForeignKeys() error {
	switch Dialect() {
	case "postgres":
		type fkRow struct {
			TableName string `gorm:"column:table_name"`
			ConName   string `gorm:"column:con_name"`
		}
		var rows []fkRow
		if err := DB.Raw(`SELECT conrelid::regclass::text AS table_name, conname AS con_name
			FROM pg_constraint
			WHERE contype = 'f' AND connamespace = 'public'::regnamespace
			ORDER BY 1, 2`).Scan(&rows).Error; err != nil {
			return fmt.Errorf("查询 PG 外键失败: %w", err)
		}
		dropped := 0
		for _, r := range rows {
			stmt := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", quotePgIdent(r.TableName), quotePgIdent(r.ConName))
			if err := DB.Exec(stmt).Error; err != nil {
				return fmt.Errorf("删除外键 %s.%s 失败: %w", r.TableName, r.ConName, err)
			}
			dropped++
		}
		if dropped > 0 {
			fmt.Printf("[db] 已删除 PostgreSQL 外键约束 %d 个（表结构不再带主外键）\n", dropped)
		}
		return nil

	case "mysql":
		type fkRow struct {
			TableName string `gorm:"column:table_name"`
			ConName   string `gorm:"column:constraint_name"`
		}
		var rows []fkRow
		if err := DB.Raw(`SELECT table_name, constraint_name
			FROM information_schema.table_constraints
			WHERE constraint_type = 'FOREIGN KEY' AND table_schema = DATABASE()
			ORDER BY 1, 2`).Scan(&rows).Error; err != nil {
			return fmt.Errorf("查询 MySQL 外键失败: %w", err)
		}
		dropped := 0
		for _, r := range rows {
			stmt := fmt.Sprintf("ALTER TABLE `%s` DROP FOREIGN KEY `%s`", r.TableName, r.ConName)
			if err := DB.Exec(stmt).Error; err != nil {
				return fmt.Errorf("删除外键 %s.%s 失败: %w", r.TableName, r.ConName, err)
			}
			dropped++
		}
		if dropped > 0 {
			fmt.Printf("[db] 已删除 MySQL 外键约束 %d 个（表结构不再带主外键）\n", dropped)
		}
		return nil

	default: // sqlite
		tables, err := sqliteTables()
		if err != nil {
			return nil
		}
		total := 0
		for _, t := range tables {
			var fks []struct {
				Table string `gorm:"column:table"`
			}
			if err := DB.Raw("PRAGMA foreign_key_list(" + quoteSqliteIdent(t) + ")").Scan(&fks).Error; err != nil {
				continue
			}
			total += len(fks)
		}
		if total > 0 {
			fmt.Printf("[db] 提示：当前 SQLite 老库仍有 %d 条外键定义（SQLite 不支持原地删除，新库不会再生成）\n", total)
		}
		return nil
	}
}

// sqliteTables 列出 SQLite 用户表（排除 sqlite_ 内部表）
func sqliteTables() ([]string, error) {
	var names []string
	err := DB.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&names).Error
	return names, err
}

// quotePgIdent 为 PostgreSQL 标识符加双引号（表名可能带 schema 前缀，如 public.sys_user）
func quotePgIdent(name string) string {
	parts := strings.Split(name, ".")
	for i, p := range parts {
		parts[i] = `"` + strings.ReplaceAll(p, `"`, `""`) + `"`
	}
	return strings.Join(parts, ".")
}

// quoteSqliteIdent 为 SQLite 表名加引号（PRAGMA 参数位置无法用占位符）
func quoteSqliteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
