package api

import (
	"time"

	"go-admin/dao"
	"go-admin/model"
	"go-admin/scheduler"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// ListJobExecutors 内置任务执行器列表（前端调用目标下拉）
func ListJobExecutors(c *gin.Context) {
	utils.OkData(c, scheduler.ListExecutors())
}

// ListJobs 分页查询定时任务
func ListJobs(c *gin.Context) {
	q := dao.JobQuery{
		PageNum:  pageNum(c),
		PageSize: pageSize(c),
		JobName:  c.Query("jobName"),
		JobGroup: c.Query("jobGroup"),
		Status:   c.Query("status"),
	}
	jobs, total, err := dao.ListJobs(q)
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, jobs, total)
}

// GetJob 查询任务详情
func GetJob(c *gin.Context) {
	j, err := dao.GetJobByID(idParam(c))
	if err != nil {
		utils.Fail(c, 500, "任务不存在")
		return
	}
	utils.OkData(c, j)
}

// JobReq 任务请求体
type JobReq struct {
	JobID          uint   `json:"jobId"`
	JobName        string `json:"jobName"`
	JobGroup       string `json:"jobGroup"`
	InvokeTarget   string `json:"invokeTarget"`
	CronExpression string `json:"cronExpression"`
	MisfirePolicy  string `json:"misfirePolicy"`
	Concurrent     string `json:"concurrent"`
	Status         string `json:"status"`
	Remark         string `json:"remark"`
}

// CreateJob 新增任务
func CreateJob(c *gin.Context) {
	var req JobReq
	if err := c.ShouldBindJSON(&req); err != nil || req.JobName == "" || req.InvokeTarget == "" || req.CronExpression == "" {
		utils.Fail(c, 500, "任务名称、调用目标和 cron 表达式不能为空")
		return
	}
	if _, err := scheduler.ParseCron(req.CronExpression); err != nil {
		utils.Fail(c, 500, "cron 表达式不正确："+err.Error())
		return
	}
	j := &model.Job{
		JobName: req.JobName, JobGroup: req.JobGroup, InvokeTarget: req.InvokeTarget,
		CronExpression: req.CronExpression, MisfirePolicy: req.MisfirePolicy,
		Concurrent: req.Concurrent, Status: req.Status, Remark: req.Remark,
		CreateBy: c.GetString("username"),
	}
	fillJobDefaults(j)
	if err := dao.CreateJob(j); err != nil {
		utils.Fail(c, 500, "新增任务失败")
		return
	}
	if j.Status == "0" {
		if err := scheduler.Register(j); err != nil {
			utils.Fail(c, 500, "任务已保存但注册失败："+err.Error())
			return
		}
	}
	utils.Ok(c)
}

// UpdateJob 修改任务（表达式变化时重新注册调度）
func UpdateJob(c *gin.Context) {
	var req JobReq
	if err := c.ShouldBindJSON(&req); err != nil || req.JobID == 0 {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	old, err := dao.GetJobByID(req.JobID)
	if err != nil {
		utils.Fail(c, 500, "任务不存在")
		return
	}
	if _, err := scheduler.ParseCron(req.CronExpression); err != nil {
		utils.Fail(c, 500, "cron 表达式不正确："+err.Error())
		return
	}
	j := &model.Job{
		ID: old.ID, JobName: req.JobName, JobGroup: req.JobGroup, InvokeTarget: req.InvokeTarget,
		CronExpression: req.CronExpression, MisfirePolicy: req.MisfirePolicy,
		Concurrent: req.Concurrent, Status: req.Status, Remark: req.Remark,
		CreateBy: old.CreateBy, CreatedAt: old.CreatedAt, UpdatedAt: time.Now(),
	}
	fillJobDefaults(j)
	if err := dao.UpdateJob(j); err != nil {
		utils.Fail(c, 500, "修改任务失败")
		return
	}
	if err := scheduler.Sync(j); err != nil {
		utils.Fail(c, 500, "任务已保存但调度同步失败："+err.Error())
		return
	}
	utils.Ok(c)
}

// fillJobDefaults 补齐默认值
func fillJobDefaults(j *model.Job) {
	if j.JobGroup == "" {
		j.JobGroup = "DEFAULT"
	}
	if j.MisfirePolicy == "" {
		j.MisfirePolicy = "3"
	}
	if j.Concurrent == "" {
		j.Concurrent = "1"
	}
	if j.Status == "" {
		j.Status = "0"
	}
}

// DeleteJob 删除任务
func DeleteJob(c *gin.Context) {
	id := idParam(c)
	if _, err := dao.GetJobByID(id); err != nil {
		utils.Fail(c, 500, "任务不存在")
		return
	}
	scheduler.Unregister(id)
	if err := dao.DeleteJob(id); err != nil {
		utils.Fail(c, 500, "删除任务失败")
		return
	}
	utils.Ok(c)
}

// ChangeJobStatus 启停任务
func ChangeJobStatus(c *gin.Context) {
	var req struct {
		JobID  uint   `json:"jobId"`
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.JobID == 0 {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	j, err := dao.GetJobByID(req.JobID)
	if err != nil {
		utils.Fail(c, 500, "任务不存在")
		return
	}
	j.Status = req.Status
	if err := dao.UpdateJob(j); err != nil {
		utils.Fail(c, 500, "修改任务状态失败")
		return
	}
	if err := scheduler.Sync(j); err != nil {
		utils.Fail(c, 500, "调度同步失败："+err.Error())
		return
	}
	utils.Ok(c)
}

// RunJob 立即执行一次任务
func RunJob(c *gin.Context) {
	var req struct {
		JobID uint `json:"jobId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.JobID == 0 {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	j, err := dao.GetJobByID(req.JobID)
	if err != nil {
		utils.Fail(c, 500, "任务不存在")
		return
	}
	scheduler.RunNow(j)
	utils.Ok(c)
}

// ListJobLogs 分页查询执行日志
func ListJobLogs(c *gin.Context) {
	q := dao.JobLogQuery{
		PageNum:  pageNum(c),
		PageSize: pageSize(c),
		JobName:  c.Query("jobName"),
		Status:   c.Query("status"),
	}
	logs, total, err := dao.ListJobLogs(q)
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, logs, total)
}

// DeleteJobLog 删除单条执行日志
func DeleteJobLog(c *gin.Context) {
	if err := dao.DeleteJobLog(idParam(c)); err != nil {
		utils.Fail(c, 500, "删除失败")
		return
	}
	utils.Ok(c)
}

// CleanJobLogs 清空执行日志
func CleanJobLogs(c *gin.Context) {
	if err := dao.CleanJobLogs(); err != nil {
		utils.Fail(c, 500, "清空失败")
		return
	}
	utils.Ok(c)
}
