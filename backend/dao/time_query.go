package dao

import (
	"fmt"
	"strings"
	"time"
	_ "time/tzdata" // 内嵌 IANA 时区数据：Windows 无系统 tzdata 时 LoadLocation 会失败
)

// bizLocation 业务时区（由 config.database.timezone 初始化）
var bizLocation = time.FixedZone("CST", 8*3600)

// initBizTimezone 校验并加载业务时区；配置非法时直接报错，避免静默按错误时区统计
func initBizTimezone(name string) error {
	if name == "" {
		name = "Asia/Shanghai"
	}
	for _, r := range name {
		if !(r == '/' || r == '_' || r == '+' || r == '-' || r == ':' ||
			(r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')) {
			return fmt.Errorf("database.timezone 含非法字符: %q", name)
		}
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return fmt.Errorf("加载时区 %q 失败: %w", name, err)
	}
	bizLocation = loc
	return nil
}

// tzName 当前业务时区名（已校验字符集，可安全拼入 SQL）
func tzName() string {
	return bizLocation.String()
}

// timeCond 返回时间范围查询的条件值。
// SQLite 时间列以 TEXT 存储，沿用字符串比较；MySQL/PostgreSQL 用 time.Time（绝对时刻），
// 避免 PostgreSQL 把字符串按"会话时区"隐式转换——会话未设 TimeZone 时日期边界会偏移。
func timeCond(s string, endOfDay bool) interface{} {
	if s == "" {
		return s
	}
	if Dialect() == "sqlite" {
		if endOfDay {
			return padEndTime(s)
		}
		return padBeginTime(s)
	}
	if t, ok := parseDayTime(s, endOfDay); ok {
		return t
	}
	// 非法/非标准格式退回原字符串，交由数据库报错便于定位
	if endOfDay {
		return padEndTime(s)
	}
	return padBeginTime(s)
}

// parseDayTime 按业务时区解析 "YYYY-MM-DD" / "YYYY-MM-DD HH:mm:ss"
func parseDayTime(s string, endOfDay bool) (time.Time, bool) {
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", s, bizLocation); err == nil {
		return t, true
	}
	if t, err := time.ParseInLocation("2006-01-02", s, bizLocation); err == nil {
		if endOfDay {
			t = t.Add(24*time.Hour - time.Second)
		}
		return t, true
	}
	return time.Time{}, false
}

// dayStart 返回某天 0 点（业务时区）
func dayStart(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, bizLocation)
}

// dayStartCond 返回某天 0 点作为查询条件值（SQLite 用日期字符串，网络库用 time.Time）
func dayStartCond(d time.Time) interface{} {
	start := dayStart(d)
	if Dialect() == "sqlite" {
		return start.Format("2006-01-02")
	}
	return start
}

// 时间范围查询辅助：前端日期选择器可能只传 YYYY-MM-DD，
// 补全为可比较的完整时间串（与 sqlite/mysql/pg 的 datetime 存储格式一致）

// padBeginTime 起止时间：纯日期补 00:00:00
func padBeginTime(t string) string {
	if len(t) == 10 {
		return t + " 00:00:00"
	}
	return t
}

// padEndTime 止时间：纯日期补 23:59:59，保证当天数据被包含
func padEndTime(t string) string {
	if len(t) == 10 {
		return t + " 23:59:59"
	}
	return t
}

// escapeLike 转义 LIKE 模式中的通配符（% / _ / 转义符本身）。
// SQLite / MySQL / PostgreSQL 的 LIKE 默认均以反斜杠作为转义字符，
// 避免用户输入 % 或 _ 被当作通配符导致搜索结果失真（参数化已防注入，此处防通配）。
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// normPage 分页参数保底：PageNum 至少 1、PageSize 至少 1 且最多 500
// （上限用于防止 pageSize=100000 这类入参一次性拉全表，占满连接与内存）
func normPage(pageNum, pageSize int) (int, int) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return pageNum, pageSize
}

// maxPageSize 单页最大条数
const maxPageSize = 500

// likeExpr 返回方言适配的模糊匹配表达式。
// PostgreSQL 的 LIKE 区分大小写，而 SQLite 的 LIKE 对 ASCII 不区分，
// 迁移后同样的搜索词会得到不同结果（如 "admin" 搜不到 "Admin"）；
// PostgreSQL 用 ILIKE 保持与 SQLite 一致的大小写不敏感语义。
func likeExpr(col string) string {
	if Dialect() == "postgres" {
		return col + " ILIKE ?"
	}
	return col + " LIKE ?"
}

// likeExprEsc 同 likeExpr，并显式指定反斜杠为转义符（配合 escapeLike 使用，
// 避免用户输入的 % / _ 被当成通配符）
func likeExprEsc(col string) string {
	if Dialect() == "postgres" {
		return col + " ILIKE ? ESCAPE '\\'"
	}
	return col + " LIKE ? ESCAPE '\\'"
}
