package middleware

import (
	"go-admin/dao"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// CheckPerm 权限校验中间件（RuoYi 风格按钮权限标识，如 system:user:list）
func CheckPerm(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userId")
		// 内置超级管理员放行；复用 Auth 写入的 DB username，不再每请求查一次 sys_user。
		if isSuperAdmin(c) {
			c.Next()
			return
		}
		perms, err := dao.GetPermsByUser(userID)
		if err != nil || !contains(perms, perm) {
			utils.Fail(c, 403, "没有权限，请联系管理员")
			c.Abort()
			return
		}
		c.Next()
	}
}

func contains(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

// isSuperAdmin 统一超管判定入口：Auth 已把数据库中的 username 写入上下文。
func isSuperAdmin(c *gin.Context) bool {
	return dao.IsSuperAdminName(c.GetString("username"))
}
