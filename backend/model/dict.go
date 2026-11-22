package model

import "time"

// DictType 字典类型表
type DictType struct {
	ID        uint      `gorm:"primaryKey" json:"dictId"`
	DictName  string    `gorm:"size:64" json:"dictName"`
	DictType  string    `gorm:"size:64;uniqueIndex" json:"dictType"`
	Status    string    `gorm:"size:8;default:0" json:"status"` // 0正常 1停用
	Remark    string    `gorm:"size:255" json:"remark"`
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
}

// TableName 指定表名
func (DictType) TableName() string { return "sys_dict_type" }

// DictData 字典数据表
type DictData struct {
	ID        uint      `gorm:"primaryKey" json:"dictCode"`
	DictSort  int       `gorm:"default:0" json:"dictSort"`
	DictLabel string    `gorm:"size:64" json:"dictLabel"`
	DictValue string    `gorm:"size:64" json:"dictValue"`
	DictType  string    `gorm:"size:64;index" json:"dictType"`
	Status    string    `gorm:"size:8;default:0" json:"status"` // 0正常 1停用
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
}

// TableName 指定表名
func (DictData) TableName() string { return "sys_dict_data" }
