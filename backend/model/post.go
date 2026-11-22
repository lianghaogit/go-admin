package model

import "time"

// Post 系统岗位表
type Post struct {
	ID        uint      `gorm:"primaryKey" json:"postId"`
	PostCode  string    `gorm:"size:64;uniqueIndex" json:"postCode"`
	PostName  string    `gorm:"size:64" json:"postName"`
	PostSort  int       `gorm:"default:0" json:"postSort"`
	Status    string    `gorm:"size:8;default:0" json:"status"` // 0正常 1停用
	Remark    string    `gorm:"size:255" json:"remark"`
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
}

// TableName 指定表名
func (Post) TableName() string { return "sys_post" }
