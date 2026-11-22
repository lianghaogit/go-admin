package dao

import (
	"time"

	"go-admin/model"
)

// JobQuery 任务查询条件
type JobQuery struct {
	PageNum  int
	PageSize int
	JobName  string
	JobGroup string
	Status   string
}

// ListJobs 分页查询任务
func ListJobs(q JobQuery) ([]model.Job, int64, error) {
	var jobs []model.Job
	var total int64
	db := DB.Model(&model.Job{})
	if q.JobName != "" {
		db = db.Where(likeExprEsc("job_name"), "%"+escapeLike(q.JobName)+"%")
	}
	if q.JobGroup != "" {
		db = db.Where("job_group = ?", q.JobGroup)
	}
	if q.Status != "" {
		db = db.Where("status = ?", q.Status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("job_id desc").Offset((q.PageNum - 1) * q.PageSize).Limit(q.PageSize).Find(&jobs).Error
	return jobs, total, err
}

// GetJobByID 按 ID 查询任务
func GetJobByID(id uint) (*model.Job, error) {
	var j model.Job
	err := DB.First(&j, id).Error
	return &j, err
}

// CreateJob 新增任务
func CreateJob(j *model.Job) error {
	return DB.Create(j).Error
}

// UpdateJob 更新任务
func UpdateJob(j *model.Job) error {
	return DB.Save(j).Error
}

// DeleteJob 删除任务
func DeleteJob(id uint) error {
	return DB.Delete(&model.Job{}, id).Error
}

// ListAllActiveJobs 查询全部正常状态任务（调度器加载用）
func ListAllActiveJobs() ([]model.Job, error) {
	var jobs []model.Job
	err := DB.Where("status = '0'").Find(&jobs).Error
	return jobs, err
}

// JobLogQuery 执行日志查询条件
type JobLogQuery struct {
	PageNum  int
	PageSize int
	JobName  string
	Status   string
}

// ListJobLogs 分页查询执行日志
func ListJobLogs(q JobLogQuery) ([]model.JobLog, int64, error) {
	var logs []model.JobLog
	var total int64
	db := DB.Model(&model.JobLog{})
	if q.JobName != "" {
		db = db.Where(likeExprEsc("job_name"), "%"+escapeLike(q.JobName)+"%")
	}
	if q.Status != "" {
		db = db.Where("status = ?", q.Status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("job_log_id desc").Offset((q.PageNum - 1) * q.PageSize).Limit(q.PageSize).Find(&logs).Error
	return logs, total, err
}

// InsertJobLog 写入执行日志
func InsertJobLog(l *model.JobLog) error {
	if l.CreateTime.IsZero() {
		l.CreateTime = time.Now()
	}
	return DB.Create(l).Error
}

// DeleteJobLog 删除单条日志
func DeleteJobLog(id uint) error {
	return DB.Delete(&model.JobLog{}, id).Error
}

// CleanJobLogs 清空执行日志
func CleanJobLogs() error {
	return DB.Exec("DELETE FROM sys_job_log").Error
}

// DeleteExpiredOperLogs 清理 N 天前的操作日志与登录日志（内置任务 clearExpiredLogs）
func DeleteExpiredOperLogs(days int) (int64, error) {
	cut := time.Now().AddDate(0, 0, -days)
	operRes := DB.Where("oper_time < ?", cut).Delete(&model.OperLog{})
	if operRes.Error != nil {
		return 0, operRes.Error
	}
	loginRes := DB.Where("login_time < ?", cut).Delete(&model.LoginLog{})
	if loginRes.Error != nil {
		return 0, loginRes.Error
	}
	return operRes.RowsAffected + loginRes.RowsAffected, nil
}
