package middleware

import (
	"net/http"
	"testing"
)

// TestResolveBusinessType 验证按路径语义修正“POST 一律新增”的粗糙映射。
func TestResolveBusinessType(t *testing.T) {
	cases := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodGet, "/api/system/user/export", BizTypeQuery},
		{http.MethodPost, "/api/system/user/importData", BizTypeInsert},
		{http.MethodPost, "/api/common/upload", BizTypeOther},
		{http.MethodPost, "/api/flow/apply", BizTypeOther},
		{http.MethodPost, "/api/system/user", BizTypeInsert},
		{http.MethodPut, "/api/system/user", BizTypeUpdate},
		{http.MethodDelete, "/api/system/user/1", BizTypeDelete},
		{http.MethodPatch, "/api/system/user/1", BizTypeOther},
	}
	for _, tc := range cases {
		if got := resolveBusinessType(tc.method, tc.path); got != tc.want {
			t.Fatalf("%s %s 业务类型 = %d，期望 %d", tc.method, tc.path, got, tc.want)
		}
	}
}
