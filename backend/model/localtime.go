package model

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// LocalTime 业务时间类型：数据库里用原生时间列，JSON 仍按 "YYYY-MM-DD HH:mm:ss" 输出。
// 需要它而不是直接用 time.Time 的原因：
//   - SQLite 与历史数据把时间存成 TEXT（"YYYY-MM-DD HH:MM:SS"），直接 time.Time 会 Scan 失败；
//   - 前端与既有接口约定用 "YYYY-MM-DD HH:mm:ss"（不带时区后缀），换 RFC3339 会改变接口格式。
//
// 各方言的列类型由 GormDBDataType 决定：postgres → timestamptz，mysql → datetime，sqlite → datetime。
type LocalTime time.Time

// layouts 兼容旧库字符串格式与标准格式
var localTimeLayouts = []string{
	"2006-01-02 15:04:05.999999999-07:00",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05Z07:00",
	time.RFC3339Nano,
	"2006-01-02",
}

// parseLocalTime 依次尝试各格式，统一按本地时区解释无时区信息的时间
func parseLocalTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "0001-01-01 00:00:00" {
		return time.Time{}, nil
	}
	for _, layout := range localTimeLayouts {
		if layout == "2006-01-02" {
			if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
				return t, nil
			}
			continue
		}
		if strings.Contains(layout, "Z07:00") || strings.HasSuffix(layout, "-07:00") {
			if t, err := time.Parse(layout, s); err == nil {
				return t.Local(), nil
			}
			continue
		}
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法解析时间: %q", s)
}

// Scan 实现 sql.Scanner：兼容 time.Time、字符串（SQLite TEXT / 驱动返回的文本）
func (t *LocalTime) Scan(src interface{}) error {
	switch v := src.(type) {
	case nil:
		*t = LocalTime(time.Time{})
	case time.Time:
		*t = LocalTime(v)
	case string:
		parsed, err := parseLocalTime(v)
		if err != nil {
			return err
		}
		*t = LocalTime(parsed)
	case []byte:
		parsed, err := parseLocalTime(string(v))
		if err != nil {
			return err
		}
		*t = LocalTime(parsed)
	default:
		return fmt.Errorf("LocalTime 不支持的类型: %T", src)
	}
	return nil
}

// Value 实现 driver.Valuer：零值写 NULL，避免出现 0001-01-01
func (t LocalTime) Value() (driver.Value, error) {
	tt := time.Time(t)
	if tt.IsZero() {
		return nil, nil
	}
	return tt, nil
}

// GormDBDataType 按方言给出原生时间列类型
func (LocalTime) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		return "timestamptz"
	case "mysql":
		return "datetime"
	default:
		return "datetime"
	}
}

// MarshalJSON 输出 "YYYY-MM-DD HH:mm:ss"（零值输出 null）
func (t LocalTime) MarshalJSON() ([]byte, error) {
	tt := time.Time(t)
	if tt.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + tt.Format("2006-01-02 15:04:05") + `"`), nil
}

// UnmarshalJSON 接受 "YYYY-MM-DD HH:mm:ss"、RFC3339、null 与空串
func (t *LocalTime) UnmarshalJSON(data []byte) error {
	s := strings.TrimSpace(string(data))
	if s == "null" || s == `""` {
		*t = LocalTime(time.Time{})
		return nil
	}
	s = strings.Trim(s, `"`)
	parsed, err := parseLocalTime(s)
	if err != nil {
		return err
	}
	*t = LocalTime(parsed)
	return nil
}

// Time 返回底层 time.Time
func (t LocalTime) Time() time.Time { return time.Time(t) }

// String 返回 "YYYY-MM-DD HH:mm:ss"
func (t LocalTime) String() string {
	tt := time.Time(t)
	if tt.IsZero() {
		return ""
	}
	return tt.Format("2006-01-02 15:04:05")
}

// ExcelString 导出到 Excel 时的文本（空值输出空字符串）
func (t LocalTime) ExcelString() string { return t.String() }

// UnmarshalExcel 从 Excel 单元格文本解析时间（支持空值）
func (t *LocalTime) UnmarshalExcel(s string) error {
	parsed, err := parseLocalTime(s)
	if err != nil {
		return fmt.Errorf("时间格式不正确（示例：2006-01-02 15:04:05）")
	}
	*t = LocalTime(parsed)
	return nil
}
