package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"

	"github.com/gin-gonic/gin"
)

// RequestIDKey 请求 ID 在 gin.Context / 访问日志中的键名。
const RequestIDKey = "request_id"

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,64}$`)

// RequestID 透传或生成 X-Request-ID，写入响应头与上下文，便于把访问日志和业务日志串起来。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if !requestIDPattern.MatchString(id) {
			id = newRequestID()
		}
		c.Set(RequestIDKey, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

func newRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "req-fallback"
	}
	return hex.EncodeToString(b)
}
