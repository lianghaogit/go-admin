package model

import "time"

// Dept 系统部门表
type Dept struct {
	ID        uint      `gorm:"primaryKey" json:"deptId"`
	ParentID  uint      `gorm:"default:0" json:"parentId"`
	DeptName  string    `gorm:"size:32" json:"deptName"`
	OrderNum  int       `gorm:"default:0" json:"orderNum"`
	Leader    string    `gorm:"size:32" json:"leader"`
	Phone     string    `gorm:"size:16" json:"phone"`
	Email     string    `gorm:"size:64" json:"email"`
	Status    string    `gorm:"size:8;default:0" json:"status"` // 0正常 1停用
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
	Children  []Dept    `gorm:"-" json:"children"`
}

// TableName 指定表名
func (Dept) TableName() string { return "sys_dept" }
