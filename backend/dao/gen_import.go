package dao

import (
	"fmt"
	"regexp"
)

// identRe 仅允许普通标识符，避免把外部输入拼进 SQL
var identRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// SyncImportSequence 导入时若写入了显式主键，PostgreSQL 的序列不会自动前移，
// 之后自增插入会撞主键；这里把序列同步到当前最大主键。MySQL/SQLite 无需处理。
func SyncImportSequence(table, pkColumn string) error {
	if Dialect() != "postgres" {
		return nil
	}
	if !identRe.MatchString(table) || !identRe.MatchString(pkColumn) {
		return fmt.Errorf("非法表名或主键列名")
	}
	sql := fmt.Sprintf(
		`SELECT setval(pg_get_serial_sequence('%s', '%s'), COALESCE((SELECT MAX(%s) FROM %s), 1))`,
		table, pkColumn, pkColumn, table,
	)
	return DB.Exec(sql).Error
}
