package model

import "time"

// User 系统用户表
type User struct {
	ID        uint      `gorm:"primaryKey" json:"userId"`
	DeptID    uint      `gorm:"default:0" json:"deptId"`
	Dept      *Dept     `gorm:"foreignKey:DeptID" json:"dept"`
	Username  string    `gorm:"size:32;uniqueIndex" json:"userName"`
	Password  string    `gorm:"size:128" json:"-"`
	Nickname  string    `gorm:"size:32" json:"nickName"`
	Email     string    `gorm:"size:64" json:"email"`
	Phone     string    `gorm:"size:16" json:"phonenumber"`
	Status    string    `gorm:"size:8;default:0" json:"status"` // 0正常 1停用
	Avatar    string    `gorm:"size:255" json:"avatar"`         // 头像 URL
	Remark    string    `gorm:"size:255" json:"remark"`
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
	Roles     []Role    `gorm:"many2many:sys_user_role" json:"roles"`
	Posts     []Post    `gorm:"many2many:sys_user_post" json:"posts"`
}

// TableName 指定表名（对齐 RuoYi 命名）
func (User) TableName() string { return "sys_user" }
