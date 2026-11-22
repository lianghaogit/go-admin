package dao

import (
	"time"

	"go-admin/model"
)

// DashStats 首页统计汇总
type DashStats struct {
	UserCount   int64 `json:"userCount"`
	RoleCount   int64 `json:"roleCount"`
	MenuCount   int64 `json:"menuCount"`
	DeptCount   int64 `json:"deptCount"`
	NoticeCount int64 `json:"noticeCount"`
	TodayLogin  int64 `json:"todayLogin"`
	OnlineCount int64 `json:"onlineCount"`
}

// GetDashStats 统计首页各模块数量
func GetDashStats() DashStats {
	var s DashStats
	DB.Model(&model.User{}).Count(&s.UserCount)
	DB.Model(&model.Role{}).Count(&s.RoleCount)
	DB.Model(&model.Menu{}).Count(&s.MenuCount)
	DB.Model(&model.Dept{}).Count(&s.DeptCount)
	DB.Model(&model.Notice{}).Count(&s.NoticeCount)
	DB.Model(&model.LoginLog{}).Where("login_time >= ?", dayStartCond(time.Now())).Count(&s.TodayLogin)
	DB.Model(&model.Online{}).Count(&s.OnlineCount)
	return s
}

// LoginTrendItem 登录趋势点
type LoginTrendItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// loginDateExpr 登录时间按天分组的表达式（PostgreSQL 的 timestamp 不支持 substr）
// 显式 AT TIME ZONE：不依赖会话时区，换连接/换驱动后分组结果不偏移
func loginDateExpr() string {
	if Dialect() == "postgres" {
		return "to_char(login_time AT TIME ZONE '" + tzName() + "', 'YYYY-MM-DD')"
	}
	return "substr(login_time,1,10)"
}

// loginHourExpr 登录时间按小时分组的表达式
func loginHourExpr() string {
	if Dialect() == "postgres" {
		return "EXTRACT(HOUR FROM login_time AT TIME ZONE '" + tzName() + "')"
	}
	return "CAST(substr(login_time,12,2) AS INTEGER)"
}

// GetLoginTrend 近 n 天每日登录次数（无登录天补 0）
func GetLoginTrend(days int) []LoginTrendItem {
	start := dayStartCond(time.Now().AddDate(0, 0, -(days - 1)))
	rows, _ := DB.Table("sys_logininfor").
		Select(loginDateExpr()+" AS date, COUNT(*) AS count").
		Where("login_time >= ?", start).
		Group("date").Order("date asc").Rows()
	countMap := map[string]int64{}
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var d string
			var c int64
			_ = rows.Scan(&d, &c)
			countMap[d] = c
		}
	}
	var out []LoginTrendItem
	for i := days - 1; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i).Format("01-02")
		full := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		out = append(out, LoginTrendItem{Date: d, Count: countMap[full]})
	}
	return out
}

// OperTypeItem 操作日志类型分布
type OperTypeItem struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

// LoginHeatItem 登录热力点（近 n 天 × 24 小时）
type LoginHeatItem struct {
	Date  string `json:"date"`
	Hour  int    `json:"hour"`
	Count int64  `json:"count"`
}

// GetLoginHeatmap 近 n 天按天×小时统计登录次数
func GetLoginHeatmap(days int) []LoginHeatItem {
	start := dayStartCond(time.Now().AddDate(0, 0, -(days - 1)))
	rows, _ := DB.Table("sys_logininfor").
		Select(loginDateExpr()+" AS date, "+loginHourExpr()+" AS hour, COUNT(*) AS count").
		Where("login_time >= ?", start).
		Group("date, hour").Rows()
	var out []LoginHeatItem
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var d string
			var h int
			var c int64
			_ = rows.Scan(&d, &h, &c)
			out = append(out, LoginHeatItem{Date: d, Hour: h, Count: c})
		}
	}
	return out
}

// GetOperTypeStats 近 n 天操作日志类型分布
func GetOperTypeStats(days int) []OperTypeItem {
	start := dayStartCond(time.Now().AddDate(0, 0, -(days - 1)))
	rows, _ := DB.Table("sys_oper_log").
		Select("business_type, COUNT(*) AS count").
		Where("oper_time >= ?", start).
		Group("business_type").Rows()
	labels := map[int]string{0: "其它", 1: "新增", 2: "修改", 3: "删除", 4: "查询"}
	countMap := map[int]int64{}
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var t int
			var c int64
			_ = rows.Scan(&t, &c)
			countMap[t] += c
		}
	}
	var out []OperTypeItem
	for _, t := range []int{4, 1, 2, 3, 0} {
		if countMap[t] > 0 {
			out = append(out, OperTypeItem{Name: labels[t], Value: countMap[t]})
		}
	}
	return out
}
