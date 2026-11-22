package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "%s", c.GetString(RequestIDKey))
	})

	t.Run("pass-through-valid-id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set("X-Request-ID", "abc-123")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if got := w.Header().Get("X-Request-ID"); got != "abc-123" {
			t.Fatalf("响应头 = %q，期望 abc-123", got)
		}
		if got := w.Body.String(); got != "abc-123" {
			t.Fatalf("上下文中 = %q，期望 abc-123", got)
		}
	})

	t.Run("replace-invalid-id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set("X-Request-ID", "bad id with space")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		got := w.Header().Get("X-Request-ID")
		if got == "" || got == "bad id with space" || !requestIDPattern.MatchString(got) {
			t.Fatalf("非法 request id 未被替换: %q", got)
		}
	})
}
