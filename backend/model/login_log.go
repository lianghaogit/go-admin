package model

import "time"

// LoginLog 登录日志表
type LoginLog struct {
	ID        uint      `gorm:"primaryKey" json:"infoId"`
	Username  string    `gorm:"size:32" json:"userName"`
	IP        string    `gorm:"size:64" json:"ipaddr"`
	Status    string    `gorm:"size:8;default:0" json:"status"` // 0成功 1失败
	Msg       string    `gorm:"size:255" json:"msg"`
	LoginTime time.Time `gorm:"index" json:"loginTime"`
}

// TableName 指定表名
func (LoginLog) TableName() string { return "sys_logininfor" }
