package dao

import (
	"context"
	"sync"
)

// backgroundWG 跟踪进程内后台 watcher（审批超时催办、日志保留清理）。
// main 在关闭 stop channel 后调用 WaitBackground，等待现有扫描退出再关闭 DB。
var backgroundWG sync.WaitGroup

// WaitBackground 等待所有已启动的后台 watcher 退出；ctx 控制等待上限。
func WaitBackground(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	done := make(chan struct{})
	go func() {
		backgroundWG.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
