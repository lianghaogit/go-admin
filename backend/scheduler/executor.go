package scheduler

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"go-admin/cache"
	"go-admin/dao"
)

// Executor 内置任务执行器定义
type Executor struct {
	Name string                 `json:"name"` // 调用目标标识（invokeTarget）
	Desc string                 `json:"desc"` // 说明（前端下拉展示）
	Fn   func() (string, error) `json:"-"`    // 执行函数（不参与 JSON 序列化）
}

// Executors 内置任务执行器注册表
// invokeTarget 取值：①内置执行器名；②http:// 或 https:// 开头的 URL（发起 HTTP GET 请求）
var Executors = []Executor{
	{
		Name: "demoTask",
		Desc: "演示任务（仅记录执行）",
		Fn: func() (string, error) {
			return "演示任务执行完成", nil
		},
	},
	{
		Name: "clearExpiredLogs",
		Desc: "清理 30 天前的操作日志与登录日志",
		Fn: func() (string, error) {
			n, err := dao.DeleteExpiredOperLogs(30)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("已清理 %d 条 30 天前的操作/登录日志", n), nil
		},
	},
	{
		Name: "memorySnapshot",
		Desc: "记录当前进程内存快照",
		Fn: func() (string, error) {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return fmt.Sprintf("进程内存：%.1f MB（堆 %.1f MB）", float64(m.Alloc)/1024/1024, float64(m.HeapAlloc)/1024/1024), nil
		},
	},
	{
		Name: "cleanExpiredOnlines",
		Desc: "清理过期的在线用户会话",
		Fn: func() (string, error) {
			n, err := dao.DeleteExpiredOnlines()
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("已清理 %d 条过期在线会话", n), nil
		},
	},
	{
		Name: "cacheSnapshot",
		Desc: "记录缓存命中率与占用统计",
		Fn: func() (string, error) {
			keyCount, hitCount, missCount, hitRate, memBytes := cache.Info()
			return fmt.Sprintf("缓存条目 %d，命中率 %.1f%%（命中 %d / 未命中 %d），占用 %.1f KB",
				keyCount, float64(hitRate), hitCount, missCount, float64(memBytes)/1024), nil
		},
	},
}

// Execute 执行任务目标：内置执行器或 HTTP URL（未注册目标仅记录）
func Execute(target string) (string, error) {
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return execHTTP(target)
	}
	for _, e := range Executors {
		if e.Name == target {
			return e.Fn()
		}
	}
	return "未注册的任务目标，仅记录本次执行", nil
}

// ListExecutors 返回内置执行器列表（前端下拉）
func ListExecutors() []Executor {
	return Executors
}

// execHTTP 发起 HTTP GET 请求并记录状态码与耗时
func execHTTP(url string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	start := time.Now()
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("HTTP 请求失败: %v", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return fmt.Sprintf("HTTP %d，耗时 %d ms", resp.StatusCode, time.Since(start).Milliseconds()), nil
}
