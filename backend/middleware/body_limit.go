package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	defaultJSONBodyLimit   = int64(4 << 20)  // 4MB
	defaultUploadBodyLimit = int64(64 << 20) // 64MB，multipart 上传/Excel 导入
)

// BodyLimit 限制请求体大小：普通 JSON 4MB，multipart 64MB。
// 先看 Content-Length 快速拒绝；chunked 请求由 MaxBytesReader 兜底。
func BodyLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil || c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead {
			c.Next()
			return
		}
		limit := defaultJSONBodyLimit
		if strings.HasPrefix(strings.ToLower(c.GetHeader("Content-Type")), "multipart/form-data") {
			limit = defaultUploadBodyLimit
		}
		if c.Request.ContentLength > limit {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"code": 413,
				"msg":  "请求体过大",
			})
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}
