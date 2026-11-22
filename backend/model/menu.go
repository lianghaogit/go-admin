package model

import "time"

// Menu 系统菜单/权限表
type Menu struct {
	ID        uint      `gorm:"primaryKey" json:"menuId"`
	MenuName  string    `gorm:"size:32" json:"menuName"`
	ParentID  uint      `gorm:"default:0" json:"parentId"`
	OrderNum  int       `gorm:"default:0" json:"orderNum"`
	Path      string    `gorm:"size:128" json:"path"`
	Component string    `gorm:"size:255" json:"component"`
	MenuType  string    `gorm:"size:4;default:C" json:"menuType"` // M目录 C菜单 F按钮
	Perms     string    `gorm:"size:128" json:"perms"`
	Icon      string    `gorm:"size:64" json:"icon"`
	Visible   string    `gorm:"size:8;default:0" json:"visible"` // 0显示 1隐藏
	Status    string    `gorm:"size:8;default:0" json:"status"`  // 0正常 1停用
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
	Children  []Menu    `gorm:"-" json:"children"`
}

// TableName 指定表名
func (Menu) TableName() string { return "sys_menu" }
