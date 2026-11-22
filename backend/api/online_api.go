package api

import (
	"go-admin/dao"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// ListOnlines 分页查询在线用户
func ListOnlines(c *gin.Context) {
	q := dao.OnlineQuery{
		PageNum:  pageNum(c),
		PageSize: pageSize(c),
		UserName: c.Query("userName"),
	}
	list, total, err := dao.ListOnlines(q)
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, list, total)
}

// ForceLogout 强退单个在线用户（删除会话，token 立即失效）
func ForceLogout(c *gin.Context) {
	tokenHash := c.Param("tokenHash")
	if tokenHash == "" || tokenHash == "all" {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	if err := dao.DeleteOnlineByTokenHash(tokenHash); err != nil {
		utils.Fail(c, 500, "强退失败")
		return
	}
	utils.Ok(c)
}

// ForceLogoutAll 强退全部在线用户
func ForceLogoutAll(c *gin.Context) {
	if err := dao.DeleteAllOnlines(); err != nil {
		utils.Fail(c, 500, "强退失败")
		return
	}
	utils.Ok(c)
}
