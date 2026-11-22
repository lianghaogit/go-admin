package api

import (
	"go-admin/dao"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// DashStats 首页统计汇总（数量卡片）
func DashStats(c *gin.Context) {
	utils.OkData(c, dao.GetDashStats())
}

// LoginTrend 近 7 天登录趋势
func LoginTrend(c *gin.Context) {
	utils.OkData(c, dao.GetLoginTrend(7))
}

// OperTypeStats 近 7 天操作日志类型分布
func OperTypeStats(c *gin.Context) {
	utils.OkData(c, dao.GetOperTypeStats(7))
}

// LoginHeatmap 近 7 天登录热力（天 × 小时）
func LoginHeatmap(c *gin.Context) {
	utils.OkData(c, dao.GetLoginHeatmap(7))
}
