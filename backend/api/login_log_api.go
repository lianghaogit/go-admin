package api

import (
	"go-admin/dao"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// ListLoginLogs 分页查询登录日志（支持用户名/IP/状态/时间范围过滤）
func ListLoginLogs(c *gin.Context) {
	q := dao.LoginLogQuery{
		PageNum:   pageNum(c),
		PageSize:  pageSize(c),
		Username:  c.Query("username"),
		IP:        c.Query("ipaddr"),
		Status:    c.Query("status"),
		BeginTime: c.Query("beginTime"),
		EndTime:   c.Query("endTime"),
	}
	logs, total, err := dao.ListLoginLogs(q)
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, logs, total)
}

// DeleteLoginLog 删除单条登录日志
func DeleteLoginLog(c *gin.Context) {
	if err := dao.DeleteLoginLog(paramUint(c, "infoId")); err != nil {
		utils.Fail(c, 500, "删除日志失败")
		return
	}
	utils.Ok(c)
}

// CleanLoginLog 清空登录日志
func CleanLoginLog(c *gin.Context) {
	if err := dao.CleanLoginLog(); err != nil {
		utils.Fail(c, 500, "清空日志失败")
		return
	}
	utils.Ok(c)
}
