package dao

import (
	"go-admin/model"
)

// NoticeQuery 公告查询条件
type NoticeQuery struct {
	PageNum  int
	PageSize int
	Title    string
	Type     string
	Status   string
}

// ListNotices 分页查询公告
func ListNotices(q NoticeQuery) ([]model.Notice, int64, error) {
	var notices []model.Notice
	var total int64
	db := DB.Model(&model.Notice{})
	if q.Title != "" {
		db = db.Where(likeExprEsc("notice_title"), "%"+escapeLike(q.Title)+"%")
	}
	if q.Type != "" {
		db = db.Where("notice_type = ?", q.Type)
	}
	if q.Status != "" {
		db = db.Where("status = ?", q.Status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("notice_id desc").Offset((q.PageNum - 1) * q.PageSize).Limit(q.PageSize).Find(&notices).Error
	return notices, total, err
}

// GetNoticeByID 按 ID 查询公告
func GetNoticeByID(id uint) (*model.Notice, error) {
	var n model.Notice
	err := DB.First(&n, id).Error
	return &n, err
}

// CreateNotice 新增公告
func CreateNotice(n *model.Notice) error {
	return DB.Create(n).Error
}

// UpdateNotice 更新公告
func UpdateNotice(n *model.Notice) error {
	return DB.Save(n).Error
}

// DeleteNotice 删除公告
func DeleteNotice(id uint) error {
	return DB.Delete(&model.Notice{}, id).Error
}

// ListActiveNotices 查询正常状态公告（首页展示用）
func ListActiveNotices(limit int) ([]model.Notice, error) {
	var notices []model.Notice
	err := DB.Where("status = '0'").Order("notice_id desc").Limit(limit).Find(&notices).Error
	return notices, err
}
