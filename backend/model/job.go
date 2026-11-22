package model

import "time"

// Job 定时任务表
type Job struct {
	ID             uint      `gorm:"column:job_id;primaryKey" json:"jobId"`
	JobName        string    `gorm:"column:job_name;size:64" json:"jobName"`
	JobGroup       string    `gorm:"column:job_group;size:64;default:DEFAULT" json:"jobGroup"`
	InvokeTarget   string    `gorm:"column:invoke_target;size:255" json:"invokeTarget"`
	CronExpression string    `gorm:"column:cron_expression;size:255" json:"cronExpression"`
	MisfirePolicy  string    `gorm:"column:misfire_policy;size:20;default:3" json:"misfirePolicy"` // 错峰策略（1立即 2执行一次 3放弃，暂存不生效）
	Concurrent     string    `gorm:"size:1;default:1" json:"concurrent"`                           // 0允许 1禁止
	Status         string    `gorm:"size:1;default:0" json:"status"`                               // 0正常 1暂停
	CreateBy       string    `gorm:"column:create_by;size:64" json:"createBy"`
	Remark         string    `gorm:"size:255" json:"remark"`
	CreatedAt      time.Time `json:"createTime"`
	UpdatedAt      time.Time `json:"updateTime"`
}

// TableName 指定表名
func (Job) TableName() string { return "sys_job" }
