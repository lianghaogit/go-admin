package dao

import (
	"go-admin/model"
)

// InsertOperLog 记录单条操作日志
func InsertOperLog(log *model.OperLog) error {
	return DB.Create(log).Error
}

// InsertOperLogs 批量记录操作日志（中间件队列聚合后调用）
func InsertOperLogs(logs []*model.OperLog) error {
	if len(logs) == 0 {
		return nil
	}
	return DB.CreateInBatches(logs, 100).Error
}

// OperLogQuery 操作日志查询条件
type OperLogQuery struct {
	PageNum   int
	PageSize  int
	Title     string
	OperName  string
	Status    string
	BeginTime string // 起（YYYY-MM-DD 或 YYYY-MM-DD HH:mm:ss）
	EndTime   string // 止
}

// ListOperLogs 分页查询操作日志
func ListOperLogs(q OperLogQuery) ([]model.OperLog, int64, error) {
	var logs []model.OperLog
	var total int64
	pageNum, pageSize := normPage(q.PageNum, q.PageSize)
	db := DB.Model(&model.OperLog{})
	if q.Title != "" {
		db = db.Where(likeExprEsc("title"), "%"+escapeLike(q.Title)+"%")
	}
	if q.OperName != "" {
		db = db.Where(likeExprEsc("oper_name"), "%"+escapeLike(q.OperName)+"%")
	}
	if q.Status != "" {
		db = db.Where("status = ?", q.Status)
	}
	if q.BeginTime != "" {
		db = db.Where("oper_time >= ?", timeCond(q.BeginTime, false))
	}
	if q.EndTime != "" {
		db = db.Where("oper_time <= ?", timeCond(q.EndTime, true))
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("id desc").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

// DeleteOperLog 删除单条操作日志
func DeleteOperLog(id uint) error {
	return DB.Delete(&model.OperLog{}, id).Error
}

// CleanOperLog 清空操作日志
func CleanOperLog() error {
	return DB.Where("1 = 1").Delete(&model.OperLog{}).Error
}
