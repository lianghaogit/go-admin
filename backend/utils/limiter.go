package utils

import (
	"sync"
	"time"
)

// 登录防爆破限流（进程内存实现，无需外部依赖）
// 维度：账号 + IP 双维度独立计数；连续失败达阈值后锁定，锁定到期自动解锁。
// 阈值与锁定时长为常量，如需调整直接改这里（生产可改为 Redis 分布式计数）。

const (
	// loginUserMaxFail 单账号连续失败锁定阈值
	loginUserMaxFail = 5
	// loginIPMaxFail 单 IP 连续失败锁定阈值（比账号维度宽松，避免同网段误伤）
	loginIPMaxFail = 20
	// loginLockMinute 锁定分钟数
	loginLockMinute = 15
	// loginEntryIdle 无锁定失败条目（未达阈值）的最长保留时长，超时回收防内存增长
	loginEntryIdle = 30 * time.Minute
)

// failEntry 单键失败计数
type failEntry struct {
	count     int
	lockUntil time.Time
	lastFail  time.Time
}

// loginLimiter 登录失败限流器（并发安全）
type loginLimiter struct {
	mu    sync.Mutex
	fails map[string]*failEntry
}

var loginLimiterInst = &loginLimiter{fails: make(map[string]*failEntry)}

func init() {
	loginLimiterInst.startGCLoop()
}

// startGCLoop 后台周期性回收过期条目，防止 map 无限增长（内存防泄漏）：
// - 已触发锁定且锁定到期超过 1 分钟的条目删除
// - 未触发锁定但超过 loginEntryIdle（30 分钟）无新失败的条目删除
func (l *loginLimiter) startGCLoop() {
	go func() {
		t := time.NewTicker(5 * time.Minute)
		defer t.Stop()
		for range t.C {
			l.gc()
		}
	}()
}

// gc 清理过期条目（持锁执行，保证并发安全）
func (l *loginLimiter) gc() {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	for k, e := range l.fails {
		if !e.lockUntil.IsZero() {
			// 锁定已到期：到期后仍需保留少量时间（allow 依赖到期后自删），超时后回收
			if now.Sub(e.lockUntil) > 1*time.Minute {
				delete(l.fails, k)
			}
		} else if now.Sub(e.lastFail) > loginEntryIdle {
			delete(l.fails, k)
		}
	}
}

// allow 判断 key 当前是否放行
func (l *loginLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.fails[key]
	if !ok {
		return true
	}
	// 有失败记录但尚未触发锁定 → 放行（不能误删计数）
	if e.lockUntil.IsZero() {
		return true
	}
	if time.Now().Before(e.lockUntil) {
		return false
	}
	// 锁定已到期，清除计数
	delete(l.fails, key)
	return true
} // fail 记录一次失败；返回剩余锁定时长（0 表示尚未触发锁定）
func (l *loginLimiter) fail(key string, max int) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	e, ok := l.fails[key]
	if !ok {
		e = &failEntry{}
		l.fails[key] = e
	} else if !e.lockUntil.IsZero() && now.After(e.lockUntil) {
		// 上一轮锁定已到期，重置计数重新累计
		e.count = 0
		e.lockUntil = time.Time{}
	}
	e.count++
	e.lastFail = now
	if e.count >= max {
		e.lockUntil = now.Add(loginLockMinute * time.Minute)
	}
	if now.Before(e.lockUntil) {
		return e.lockUntil.Sub(now)
	}
	return 0
}

// reset 登录成功时清除账号维度计数
func (l *loginLimiter) reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.fails, key)
}

// LoginUserAllowed 登录放行判断：账号与 IP 均未锁定时放行
func LoginUserAllowed(username, ip string) bool {
	return loginLimiterInst.allow("u:"+username) && loginLimiterInst.allow("i:"+ip)
}

// LoginUserFail 记录一次登录失败（账号 + IP 双维度），返回剩余锁定时长
func LoginUserFail(username, ip string) time.Duration {
	d1 := loginLimiterInst.fail("u:"+username, loginUserMaxFail)
	d2 := loginLimiterInst.fail("i:"+ip, loginIPMaxFail)
	if d1 > d2 {
		return d1
	}
	return d2
}

// LoginUserReset 登录成功后清除该账号维度计数（IP 维度保留，防同 IP 换账号继续爆破）
func LoginUserReset(username string) {
	loginLimiterInst.reset("u:" + username)
}
