package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// Cors 跨域中间件。
// 默认仅允许本地开发前端（5173）；生产环境通过 GO_ADMIN_ALLOW_ORIGIN 配置具体来源，
// 多个来源用逗号分隔，例如 GO_ADMIN_ALLOW_ORIGIN=https://admin.example.com
func Cors() gin.HandlerFunc {
	allowOrigin := os.Getenv("GO_ADMIN_ALLOW_ORIGIN")
	if allowOrigin == "" {
		allowOrigin = "http://localhost:5173,http://127.0.0.1:5173"
	}
	return func(c *gin.Context) {
		if allowOrigin == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			origin := c.GetHeader("Origin")
			for _, o := range strings.Split(allowOrigin, ",") {
				if strings.TrimSpace(o) == origin {
					c.Header("Access-Control-Allow-Origin", origin)
					c.Header("Vary", "Origin")
					break
				}
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
