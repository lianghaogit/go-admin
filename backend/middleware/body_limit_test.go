package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBodyLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(BodyLimit())
	r.POST("/json", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	t.Run("too-large-json", func(t *testing.T) {
		body := bytes.Repeat([]byte("a"), int(defaultJSONBodyLimit)+1)
		req := httptest.NewRequest(http.MethodPost, "/json", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("status = %d，期望 413", w.Code)
		}
	})

	t.Run("normal-json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/json", bytes.NewReader([]byte(`{"a":1}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK || w.Body.String() != "ok" {
			t.Fatalf("status = %d body = %q", w.Code, w.Body.String())
		}
	})
}
