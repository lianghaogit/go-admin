package dao

import (
	"go-admin/model"
)

// LoginLogQuery 登录日志查询条件
type LoginLogQuery struct {
	PageNum   int
	PageSize  int
	Username  string
	IP        string
	Status    string
	BeginTime string // 起（YYYY-MM-DD 或 YYYY-MM-DD HH:mm:ss）
	EndTime   string // 止
}

// InsertLoginLog 记录登录日志
func InsertLoginLog(log *model.LoginLog) error {
	return DB.Create(log).Error
}

// ListLoginLogs 分页查询登录日志
func ListLoginLogs(q LoginLogQuery) ([]model.LoginLog, int64, error) {
	var logs []model.LoginLog
	var total int64
	pageNum, pageSize := normPage(q.PageNum, q.PageSize)
	db := DB.Model(&model.LoginLog{})
	if q.Username != "" {
		db = db.Where(likeExprEsc("username"), "%"+escapeLike(q.Username)+"%")
	}
	if q.IP != "" {
		db = db.Where(likeExprEsc("ipaddr"), "%"+escapeLike(q.IP)+"%")
	}
	if q.Status != "" {
		db = db.Where("status = ?", q.Status)
	}
	if q.BeginTime != "" {
		db = db.Where("login_time >= ?", timeCond(q.BeginTime, false))
	}
	if q.EndTime != "" {
		db = db.Where("login_time <= ?", timeCond(q.EndTime, true))
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("id desc").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

// DeleteLoginLog 删除单条登录日志
func DeleteLoginLog(id uint) error {
	return DB.Where("id = ?", id).Delete(&model.LoginLog{}).Error
}

// CleanLoginLog 清空登录日志
func CleanLoginLog() error {
	return DB.Exec("DELETE FROM sys_logininfor").Error
}
