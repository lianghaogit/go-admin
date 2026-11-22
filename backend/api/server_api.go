package api

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"go-admin/middleware"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

// startTime 服务启动时间
var startTime = time.Now()

// ServerInfo 查询服务监控信息（CPU/内存/磁盘/系统）
func ServerInfo(c *gin.Context) {
	info := gin.H{}
	// 系统信息
	if h, err := host.Info(); err == nil {
		info["os"] = h.Platform + " " + h.PlatformVersion
		info["hostName"] = h.Hostname
		info["arch"] = h.KernelArch
		info["kernel"] = h.KernelVersion
	}
	info["goVersion"] = runtime.Version()
	info["cpuNum"] = runtime.NumCPU()
	info["startTime"] = startTime.Format("2006-01-02 15:04:05")
	info["runDays"] = int(time.Since(startTime).Hours() / 24)
	info["runHours"] = int(time.Since(startTime).Hours()) % 24
	info["runMinutes"] = int(time.Since(startTime).Minutes()) % 60
	// CPU 使用率（采样 200ms）
	if ps, err := cpu.Percent(200*time.Millisecond, false); err == nil && len(ps) > 0 {
		info["cpuUsage"] = round1(ps[0])
	}
	// 内存
	if m, err := mem.VirtualMemory(); err == nil {
		info["memTotal"] = m.Total
		info["memUsed"] = m.Used
		info["memUsage"] = round1(m.UsedPercent)
	}
	// 磁盘（取当前工作目录所在分区：Windows 可能是 D: 等，非 Windows 用 /）
	diskPath := "/"
	if runtime.GOOS == "windows" {
		if wd, err := os.Getwd(); err == nil {
			if v := filepath.VolumeName(wd); v != "" {
				diskPath = v
			}
		}
	}
	if ds, err := disk.Usage(diskPath); err == nil {
		info["diskTotal"] = ds.Total
		info["diskUsed"] = ds.Used
		info["diskUsage"] = round1(ds.UsedPercent)
	}
	// 进程内存
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	info["procMemUsed"] = ms.Alloc
	info["procMemTotal"] = ms.Sys
	// 操作日志异步队列观测：积压、丢弃、批量写入失败
	info["operLog"] = middleware.OperLogStats()
	utils.OkData(c, info)
}

// round1 保留 1 位小数
func round1(v float64) float64 {
	return float64(int(v*10)) / 10
}
