package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"go-admin/config"
	"go-admin/dao"
	"go-admin/middleware"
	"go-admin/router"
	"go-admin/scheduler"

	"github.com/gin-gonic/gin"
)

// exeDir 返回可执行文件所在目录
func exeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

// resolveConfigPath 定位配置文件：
// 优先可执行文件目录（go run 编译到临时目录时此处无效），其次当前工作目录（go run 从 backend 目录执行时生效）
func resolveConfigPath() string {
	candidates := make([]string, 0, 2)
	if dir := exeDir(); dir != "." {
		candidates = append(candidates, filepath.Join(dir, "config.yaml"))
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "config.yaml"))
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if len(candidates) > 0 {
		return candidates[0]
	}
	return "config.yaml"
}

func main() {
	configPath := resolveConfigPath()
	if err := config.Load(configPath); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	// 数据库相对路径基于配置文件所在目录解析（仅 SQLite 生效，与 go run / 直接运行 exe 两种方式保持一致）
	if config.Cfg.Database.Type == "" || config.Cfg.Database.Type == "sqlite" {
		if !filepath.IsAbs(config.Cfg.Database.Path) {
			config.Cfg.Database.Path = filepath.Join(filepath.Dir(configPath), config.Cfg.Database.Path)
		}
	}
	// 上传目录同样基于配置文件所在目录
	config.Cfg.Server.UploadDir = filepath.Join(filepath.Dir(configPath), "uploads")

	switch config.Cfg.Server.Mode {
	case "release":
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
	if err := dao.InitDB(config.Cfg.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	// pprof 仅在显式设置地址时开启，默认关闭；建议只绑 127.0.0.1。
	if addr := os.Getenv("GO_ADMIN_PPROF_ADDR"); addr != "" {
		go func() {
			log.Printf("pprof 监听 %s（仅建议本机访问）", addr)
			if err := http.ListenAndServe(addr, nil); err != nil {
				log.Printf("pprof 监听退出: %v", err)
			}
		}()
	}
	r := router.InitRouter()
	// 启动定时任务调度器（加载正常状态任务）
	scheduler.Start()
	// 启动审批超时催办 watcher（每分钟扫描运行中实例的待办任务）
	stopOverdue := make(chan struct{})
	dao.StartOverdueWatcher(stopOverdue)
	// 启动日志保留清理（logs.retentionDays > 0 时生效，默认关闭）
	dao.StartLogRetention(stopOverdue)
	// 启动在线会话过期清理（单实例也避免 sys_online 只增不减）
	dao.StartOnlineCleanup(stopOverdue, 10*time.Minute)
	addr := fmt.Sprintf(":%d", config.Cfg.Server.Port)

	srv := &http.Server{Addr: addr, Handler: r}
	go func() {
		log.Printf("Go Admin 服务启动，监听 %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 优雅退出：等待 SIGINT/SIGTERM，先停 HTTP，再停调度器与数据库连接
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("收到退出信号，正在优雅关闭...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	// 先停后台 watcher，再关操作日志队列（flush 剩余批次），最后才关库。
	close(stopOverdue)
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 5*time.Second)
	_ = dao.WaitBackground(waitCtx)
	waitCancel()
	operCtx, operCancel := context.WithTimeout(context.Background(), 3*time.Second)
	if err := middleware.StopOperLog(operCtx); err != nil {
		log.Printf("操作日志队列未在超时内排空: %v", err)
	}
	operCancel()
	scheduler.Stop()
	if sqlDB, err := dao.DB.DB(); err == nil {
		_ = sqlDB.Close()
	}
	log.Println("服务已退出")
}
