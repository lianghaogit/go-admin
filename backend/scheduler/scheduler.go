package scheduler

import (
	"sync"

	"go-admin/dao"
	"go-admin/model"

	"github.com/robfig/cron/v3"
)

var (
	cronSvc *cron.Cron
	mu      sync.Mutex
	entries = map[uint]cron.EntryID{}
	running = map[uint]bool{}
)

// ParseCron 校验 cron 表达式（6 段：秒 分 时 日 月 周）
func ParseCron(expr string) (cron.Schedule, error) {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	return parser.Parse(expr)
}

// Start 启动调度器并加载全部正常状态任务（启动失败不阻断服务）
func Start() {
	cronSvc = cron.New(cron.WithSeconds()) // 6 段：秒 分 时 日 月 周
	jobs, err := dao.ListAllActiveJobs()
	if err == nil {
		for i := range jobs {
			if err := Register(&jobs[i]); err != nil {
				continue // 表达式无效的任务跳过注册
			}
		}
	}
	cronSvc.Start()
}

// Stop 停止调度器
func Stop() {
	if cronSvc != nil {
		cronSvc.Stop()
	}
}

// Register 注册任务到调度器
func Register(j *model.Job) error {
	if cronSvc == nil {
		return nil
	}
	entryID, err := cronSvc.AddFunc(j.CronExpression, func() {
		runJob(j)
	})
	if err != nil {
		return err
	}
	mu.Lock()
	entries[j.ID] = entryID
	mu.Unlock()
	return nil
}

// Unregister 从调度器移除任务
func Unregister(id uint) {
	mu.Lock()
	eid, ok := entries[id]
	if ok {
		delete(entries, id)
	}
	mu.Unlock()
	if ok && cronSvc != nil {
		cronSvc.Remove(eid)
	}
}

// Sync 任务更新后同步调度（注销后按状态重注册）
func Sync(j *model.Job) error {
	Unregister(j.ID)
	if j.Status != "0" {
		return nil
	}
	return Register(j)
}

// RunNow 立即执行一次任务
func RunNow(j *model.Job) {
	go runJob(j)
}

// runJob 执行任务并写入执行日志（concurrent=1 时并发保护）
func runJob(j *model.Job) {
	if j.Concurrent == "1" {
		mu.Lock()
		if running[j.ID] {
			mu.Unlock()
			return
		}
		running[j.ID] = true
		mu.Unlock()
		defer func() {
			mu.Lock()
			delete(running, j.ID)
			mu.Unlock()
		}()
	}
	log := &model.JobLog{
		JobName:      j.JobName,
		JobGroup:     j.JobGroup,
		InvokeTarget: j.InvokeTarget,
		Status:       "0",
	}
	msg, err := Execute(j.InvokeTarget)
	if err != nil {
		log.Status = "1"
		log.JobMessage = err.Error()
		log.ExceptionInfo = err.Error()
	} else {
		log.JobMessage = msg
	}
	_ = dao.InsertJobLog(log)
}
