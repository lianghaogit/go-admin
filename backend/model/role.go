package model

import "time"

// Role 系统角色表
type Role struct {
	ID        uint      `gorm:"primaryKey" json:"roleId"`
	RoleName  string    `gorm:"size:32" json:"roleName"`
	RoleKey   string    `gorm:"size:64;uniqueIndex" json:"roleKey"`
	Sort      int       `gorm:"default:0" json:"roleSort"`
	Status    string    `gorm:"size:8;default:0" json:"status"`    // 0正常 1停用
	DataScope string    `gorm:"size:8;default:1" json:"dataScope"` // 1全部 2自定义 3本部门 4本部门及以下 5仅本人
	Remark    string    `gorm:"size:255" json:"remark"`
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
	Menus     []Menu    `gorm:"many2many:sys_role_menu" json:"-"`
	Depts     []Dept    `gorm:"many2many:sys_role_dept" json:"-"`
	Users     []User    `gorm:"many2many:sys_user_role" json:"-"`
}

// TableName 指定表名
func (Role) TableName() string { return "sys_role" }
