package model

import "time"

// Online 在线用户会话表
type Online struct {
	TokenHash string    `gorm:"primaryKey;size:64" json:"tokenHash"`
	UserID    uint      `gorm:"index" json:"userId"`
	UserName  string    `gorm:"size:64" json:"userName"`
	IP        string    `gorm:"size:64" json:"ip"`
	Browser   string    `gorm:"size:64" json:"browser"`
	OS        string    `gorm:"size:64" json:"os"`
	LoginTime time.Time `json:"loginTime"`
	ExpireAt  time.Time `json:"expireAt"`
}

// TableName 指定表名
func (Online) TableName() string { return "sys_online" }
