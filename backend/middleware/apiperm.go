package middleware

import (
	"go-admin/dao"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// CheckApiPerm 接口权限中间件（数据库可配置）：匹配 sys_api 规则并校验权限字符
// admin 放行；未配置规则或规则停用则放行；perms 为空表示仅需登录
func CheckApiPerm() gin.HandlerFunc {
	return func(c *gin.Context) {
		rule := dao.MatchApiRule(c.Request.Method, c.Request.URL.Path)
		if rule == nil || rule.Perms == "" {
			c.Next()
			return
		}
		userID := c.GetUint("userId")
		// Auth 已把数据库中的 username 放入上下文；超管判定统一走 IsSuperAdminName。
		if isSuperAdmin(c) {
			c.Next()
			return
		}
		perms, err := dao.GetPermsByUser(userID)
		if err != nil || !contains(perms, rule.Perms) {
			utils.Fail(c, 403, "没有该接口的访问权限，请联系管理员")
			c.Abort()
			return
		}
		c.Next()
	}
}
