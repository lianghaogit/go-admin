package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Ok 成功无数据
func Ok(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "操作成功"})
}

// OkMsg 成功响应（自定义提示消息）
func OkMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": msg})
}

// OkData 成功带数据
func OkData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "操作成功", "data": data})
}

// OkTable 成功返回表格数据（RuoYi 风格 rows/total）
func OkTable(c *gin.Context, rows interface{}, total int64) {
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "查询成功", "rows": rows, "total": total})
}

// Fail 失败（HTTP 始终 200，业务码在 body，与 RuoYi 前端约定一致）
func Fail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, gin.H{"code": code, "msg": msg})
}
