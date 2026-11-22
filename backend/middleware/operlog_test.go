package middleware

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// TestTruncateUTF8 校验超长/非法 UTF-8 文本被按字符边界截断
// 背景：PG/MySQL 的 varchar 严格校验长度与编码，按字节切中文会产生非法 UTF-8 导致日志写入失败
func TestTruncateUTF8(t *testing.T) {
	// 700 个中文字符 = 2100 字节，超过 2000 字节上限
	longCN := strings.Repeat("中", 700)
	got := truncateUTF8(longCN, 2000)
	if !utf8.ValidString(got) {
		t.Fatalf("截断结果不是合法 UTF-8")
	}
	if len(got) > 2000 {
		t.Fatalf("截断结果 %d 字节，超过上限 2000", len(got))
	}
	if !strings.HasSuffix(got, truncMarker) {
		t.Fatalf("截断结果缺少截断标记: %q", got[len(got)-12:])
	}
	// 去掉标记后必须是 "中" 的完整倍数，即字符边界正确
	body := strings.TrimSuffix(got, truncMarker)
	if strings.Count(body, "中")*3 != len(body) {
		t.Fatalf("截断点落在多字节字符中间")
	}
	if len(got) < 1990 {
		t.Fatalf("截断过于保守：%d 字节（应接近 2000）", len(got))
	}

	// 非法字节应被清洗（0x89 非法被丢弃，0x50 是合法 ASCII 'P' 保留）
	if got := truncateUTF8("abc\x89\x50def", 10); got != "abcPdef" {
		t.Fatalf("非法字节清洗失败: %q", got)
	}
	// 连续非法字节 + 截断长度内：结果必须合法且不超长
	if got := truncateUTF8("中文\x89\x89\x50\x4E\x47中文", 8); !utf8.ValidString(got) || len(got) > 8 {
		t.Fatalf("混合非法字节处理异常: %q", got)
	}

	// 未超长时原样返回
	if got := truncateUTF8("hello 中文", 100); got != "hello 中文" {
		t.Fatalf("未超长内容被改写: %q", got)
	}

	// 空串与零上限
	if got := truncateUTF8("", 100); got != "" {
		t.Fatalf("空串处理异常: %q", got)
	}
	if got := truncateUTF8("abc", 0); got != "" {
		t.Fatalf("零上限处理异常: %q", got)
	}
}

// TestTruncateUTF8CapsMatchModel 确认截断上限与 model.OperLog 的 size 定义一致
func TestTruncateUTF8CapsMatchModel(t *testing.T) {
	cases := []struct {
		name string
		got  int
		want int
	}{
		{"title", maxOperTitle, 128},
		{"method", maxOperMethod, 255},
		{"ip", maxOperIP, 64},
		{"url", maxOperURL, 255},
		{"operName", maxOperName, 32},
		{"msg", maxOperMsg, 2000},
		{"param", LogMaxParamLen, 2000},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Fatalf("%s 上限 %d，与 model size %d 不一致", c.name, c.got, c.want)
		}
	}
}
