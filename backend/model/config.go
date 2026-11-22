package model

import "time"

// Config 系统参数配置表
type Config struct {
	ID          uint      `gorm:"primaryKey" json:"configId"`
	ConfigName  string    `gorm:"size:64" json:"configName"`
	ConfigKey   string    `gorm:"size:64;uniqueIndex" json:"configKey"`
	ConfigValue string    `gorm:"size:255" json:"configValue"`
	ConfigType  string    `gorm:"size:8;default:Y" json:"configType"` // Y内置 N非内置
	Remark      string    `gorm:"size:255" json:"remark"`
	CreatedAt   time.Time `json:"createTime"`
	UpdatedAt   time.Time `json:"updateTime"`
}

// TableName 指定表名
func (Config) TableName() string { return "sys_config" }
