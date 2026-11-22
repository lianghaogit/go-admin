package model

import "time"

// JobLog 定时任务执行日志表
type JobLog struct {
	ID            uint      `gorm:"column:job_log_id;primaryKey" json:"jobLogId"`
	JobName       string    `gorm:"column:job_name;size:64" json:"jobName"`
	JobGroup      string    `gorm:"column:job_group;size:64" json:"jobGroup"`
	InvokeTarget  string    `gorm:"column:invoke_target;size:255" json:"invokeTarget"`
	Status        string    `gorm:"size:1;default:0" json:"status"` // 0成功 1失败
	JobMessage    string    `gorm:"column:job_message;type:text" json:"jobMessage"`
	ExceptionInfo string    `gorm:"column:exception_info;type:text" json:"exceptionInfo"`
	CreateTime    time.Time `gorm:"column:create_time" json:"createTime"`
}

// TableName 指定表名
func (JobLog) TableName() string { return "sys_job_log" }
