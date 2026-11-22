package middleware

import (
	"strings"
	"testing"
)

// TestMaskSensitiveNested 验证嵌套对象与数组内的密码字段都会被脱敏。
func TestMaskSensitiveNested(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		leak []string
	}{
		{"top-level", `{"password":"top-secret","name":"admin"}`, []string{"top-secret"}},
		{"nested-object", `{"user":{"password":"nested-secret"},"name":"admin"}`, []string{"nested-secret"}},
		{"array-object", `{"users":[{"newPwd":"array-secret"},{"profile":{"passwd":"deep-secret"}}]}`, []string{"array-secret", "deep-secret"}},
		{"top-level-array", `[{"password":"root-secret"}]`, []string{"root-secret"}},
		{"non-string-value", `{"password":123456}`, []string{"123456"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := maskSensitive(tc.raw)
			if !strings.Contains(got, "******") {
				t.Fatalf("脱敏结果缺少 ******：%s", got)
			}
			for _, leak := range tc.leak {
				if strings.Contains(got, leak) {
					t.Fatalf("仍含明文 %q：%s", leak, got)
				}
			}
		})
	}
}

// TestMaskSensitiveNoChange 非敏感字段不应被改写。
func TestMaskSensitiveNoChange(t *testing.T) {
	raw := `{"username":"admin","nickname":"管理员"}`
	if got := maskSensitive(raw); got != raw {
		t.Fatalf("非敏感字段被改写：%s", got)
	}
}
