package utils

import (
	"testing"
	"time"
)

func TestLoginLimiterFail(t *testing.T) {
	// 连续 5 次失败触发锁定
	for i := 0; i < 5; i++ {
		LoginUserFail("lim_user", "10.9.9.9")
	}
	if LoginUserAllowed("lim_user", "10.9.9.9") {
		t.Fatalf("5 次失败后应锁定")
	}
	// 锁定期间即使换 IP 也应被账号维度拦截
	if LoginUserAllowed("lim_user", "10.9.9.10") {
		t.Fatalf("账号维度锁定应跨 IP 生效")
	}
	// 其他账号不受影响
	if !LoginUserAllowed("other_user", "10.9.9.9") {
		t.Fatalf("其他账号不应被锁")
	}
}

func TestLoginLimiterExpire(t *testing.T) {
	// 锁定到期自动解锁
	key := "exp_user"
	for i := 0; i < 5; i++ {
		LoginUserFail(key, "10.9.9.11")
	}
	// 直接改锁定期望到过去
	loginLimiterInst.mu.Lock()
	e := loginLimiterInst.fails["u:"+key]
	e.lockUntil = time.Now().Add(-time.Second)
	loginLimiterInst.mu.Unlock()
	if !LoginUserAllowed(key, "10.9.9.11") {
		t.Fatalf("锁定到期应自动解锁")
	}
}

func TestLoginLimiterGC(t *testing.T) {
	// 先分别制造失败记录（不在持锁期间调用 LoginUserFail，避免自死锁）
	k1 := "gc_lock_user"
	for i := 0; i < 5; i++ {
		LoginUserFail(k1, "10.9.9.12")
	}
	k2 := "gc_idle_user"
	LoginUserFail(k2, "10.9.9.13")
	k3 := "gc_fresh_user"
	LoginUserFail(k3, "10.9.9.14")

	// 持锁修改时间戳：k1 锁定已过期、k2 空闲超时、k3 保留
	loginLimiterInst.mu.Lock()
	loginLimiterInst.fails["u:"+k1].lockUntil = time.Now().Add(-2 * time.Minute)
	loginLimiterInst.fails["u:"+k2].lastFail = time.Now().Add(-loginEntryIdle - time.Minute)
	loginLimiterInst.mu.Unlock()

	loginLimiterInst.gc()

	loginLimiterInst.mu.Lock()
	defer loginLimiterInst.mu.Unlock()
	if _, ok := loginLimiterInst.fails["u:"+k1]; ok {
		t.Fatalf("过期锁定条目应被 gc 回收")
	}
	if _, ok := loginLimiterInst.fails["u:"+k2]; ok {
		t.Fatalf("超时空闲条目应被 gc 回收")
	}
	if _, ok := loginLimiterInst.fails["u:"+k3]; !ok {
		t.Fatalf("未超时空闲条目应保留")
	}
}
