package api

import (
	"go-admin/dao"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// ListOperLogs 分页查询操作日志
func ListOperLogs(c *gin.Context) {
	q := dao.OperLogQuery{
		PageNum:   pageNum(c),
		PageSize:  pageSize(c),
		Title:     c.Query("title"),
		OperName:  c.Query("operName"),
		Status:    c.Query("status"),
		BeginTime: c.Query("beginTime"),
		EndTime:   c.Query("endTime"),
	}
	logs, total, err := dao.ListOperLogs(q)
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, logs, total)
}

// DeleteOperLog 删除单条操作日志
func DeleteOperLog(c *gin.Context) {
	if err := dao.DeleteOperLog(paramUint(c, "operId")); err != nil {
		utils.Fail(c, 500, "删除日志失败")
		return
	}
	utils.Ok(c)
}

// CleanOperLog 清空操作日志
func CleanOperLog(c *gin.Context) {
	if err := dao.CleanOperLog(); err != nil {
		utils.Fail(c, 500, "清空日志失败")
		return
	}
	utils.Ok(c)
}
