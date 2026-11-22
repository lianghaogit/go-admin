package model

import "time"

// OperLog 操作日志表
type OperLog struct {
	ID            uint      `gorm:"primaryKey" json:"operId"`
	Title         string    `gorm:"size:128" json:"title"` // 模块标题（路由）
	BusinessType  int       `gorm:"default:0" json:"businessType"`
	Method        string    `gorm:"size:255" json:"method"` // 处理器名
	RequestMethod string    `gorm:"size:16" json:"requestMethod"`
	OperIP        string    `gorm:"size:64" json:"operIp"`
	OperURL       string    `gorm:"size:255" json:"operUrl"`
	OperParam     string    `gorm:"size:2000" json:"operParam"`
	Status        string    `gorm:"size:8;default:0" json:"status"` // 0成功 1失败
	ErrorMsg      string    `gorm:"size:2000" json:"errorMsg"`
	CostTime      int64     `json:"costTime"` // 毫秒
	OperTime      time.Time `gorm:"index;index:idx_oper_name" json:"operTime"`
	OperName      string    `gorm:"size:32;index:idx_oper_name" json:"operName"`
	RequestID     string    `gorm:"size:64;index" json:"requestId"` // 按请求串访问日志与操作日志
}

// TableName 指定表名
func (OperLog) TableName() string { return "sys_oper_log" }
