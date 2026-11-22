package dao

import (
	"testing"
	"time"
)

func TestEscapeLike(t *testing.T) {
	cases := []struct{ in, want string }{
		{"abc", "abc"},
		{"a%b", `a\%b`},
		{"a_b", `a\_b`},
		{`a\b`, `a\\b`},
		{`%_\%`, `\%\_\\\%`},
	}
	for _, c := range cases {
		if got := escapeLike(c.in); got != c.want {
			t.Fatalf("escapeLike(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormPage(t *testing.T) {
	cases := []struct{ pn, ps, wPN, wPS int }{
		{0, 0, 1, 10},
		{-5, -1, 1, 10},
		{1, 10, 1, 10},
		{3, 20, 3, 20},
		{2, 100000, 2, maxPageSize},
		{2, maxPageSize + 1, 2, maxPageSize},
	}
	for _, c := range cases {
		pn, ps := normPage(c.pn, c.ps)
		if pn != c.wPN || ps != c.wPS {
			t.Fatalf("normPage(%d,%d) = (%d,%d), want (%d,%d)", c.pn, c.ps, pn, ps, c.wPN, c.wPS)
		}
	}
}

func TestPadTimes(t *testing.T) {
	if got := padBeginTime("2026-10-31"); got != "2026-10-31 00:00:00" {
		t.Fatalf("padBeginTime = %q", got)
	}
	if got := padEndTime("2026-10-31"); got != "2026-10-31 23:59:59" {
		t.Fatalf("padEndTime = %q", got)
	}
	if got := padBeginTime("2026-10-31 08:00:00"); got != "2026-10-31 08:00:00" {
		t.Fatalf("padBeginTime full = %q", got)
	}
}

// TestParseDayTime 校验前端传来的日期字符串按业务时区解析（不依赖机器/数据库会话时区）
func TestParseDayTime(t *testing.T) {
	if err := initBizTimezone("Asia/Shanghai"); err != nil {
		t.Fatalf("加载时区失败: %v", err)
	}
	start, ok := parseDayTime("2026-11-05", false)
	if !ok {
		t.Fatalf("纯日期解析失败")
	}
	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Fatalf("起始时间应为 00:00:00，实际 %v", start)
	}
	if start.Location().String() != "Asia/Shanghai" {
		t.Fatalf("解析时区应为 Asia/Shanghai，实际 %v", start.Location())
	}
	end, ok := parseDayTime("2026-11-05", true)
	if !ok || end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Fatalf("截止时间应为 23:59:59，实际 %v", end)
	}
	full, ok := parseDayTime("2026-11-05 18:24:56", false)
	if !ok || full.Hour() != 18 || full.Minute() != 24 || full.Second() != 56 {
		t.Fatalf("完整时间解析异常: %v", full)
	}
	if _, ok := parseDayTime("not-a-date", false); ok {
		t.Fatalf("非法字符串不应解析成功")
	}
}

// TestDayStart 校验按天取整使用业务时区
func TestDayStart(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("加载时区失败: %v", err)
	}
	in := time.Date(2026, 11, 5, 18, 24, 56, 0, loc)
	got := dayStart(in)
	if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 || got.Day() != 5 {
		t.Fatalf("dayStart 结果异常: %v", got)
	}
}

// TestInitBizTimezoneRejectsInjection 时区名会拼入 SQL，必须拒绝非法字符
func TestInitBizTimezoneRejectsInjection(t *testing.T) {
	if err := initBizTimezone("Asia/Shanghai'--"); err == nil {
		t.Fatalf("含引号的时区名应被拒绝")
	}
	if err := initBizTimezone("Asia/Shanghai"); err != nil {
		t.Fatalf("合法时区应通过: %v", err)
	}
}
