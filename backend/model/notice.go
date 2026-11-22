package model

import "time"

// Notice 通知公告表
type Notice struct {
	ID        uint      `gorm:"column:notice_id;primaryKey" json:"noticeId"`
	Title     string    `gorm:"column:notice_title;size:128" json:"noticeTitle"`
	Type      string    `gorm:"column:notice_type;size:1;default:1" json:"noticeType"` // 1通知 2公告
	Content   string    `gorm:"column:notice_content;type:text" json:"noticeContent"`
	Status    string    `gorm:"size:1;default:0" json:"status"` // 0正常 1关闭
	CreateBy  string    `gorm:"column:create_by;size:64" json:"createBy"`
	Remark    string    `gorm:"size:255" json:"remark"`
	CreatedAt time.Time `gorm:"index" json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
}

// TableName 指定表名
func (Notice) TableName() string { return "sys_notice" }
