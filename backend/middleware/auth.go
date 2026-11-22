package middleware

import (
	"strings"

	"go-admin/config"
	"go-admin/dao"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// Auth JWT 鉴权中间件：校验 token 并确认用户仍存在且未停用
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			utils.Fail(c, 401, "认证失败，无法访问系统资源")
			c.Abort()
			return
		}
		claims, err := utils.ParseToken(strings.TrimPrefix(header, "Bearer "), config.Cfg.JWT.Secret)
		if err != nil {
			utils.Fail(c, 401, "认证失败，无法访问系统资源")
			c.Abort()
			return
		}
		// 校验用户存在且状态正常（删除/停用后 token 立即失效）。
		// 只取鉴权字段，避免 GetUserByID 的 Roles/Posts/Dept 三次预加载。
		user, err := dao.GetUserAuthInfo(claims.UserID)
		if err != nil || user.Status != "0" {
			utils.Fail(c, 401, "账号已被停用或不存在，请重新登录")
			c.Abort()
			return
		}
		// 在线会话校验（强退后 token 立即失效）
		token := strings.TrimPrefix(header, "Bearer ")
		online, err := dao.OnlineExists(utils.TokenHash(token))
		if err != nil || !online {
			utils.Fail(c, 401, "登录状态已失效，请重新登录")
			c.Abort()
			return
		}
		c.Set("userId", claims.UserID)
		// 用数据库中的 username 覆盖 JWT 里的旧值：账号改名后仍能正确判断 admin。
		c.Set("username", user.Username)
		c.Next()
	}
}
