package dao

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go-admin/model"
)

// TestOnlineExistsExcludesExpired 验证鉴权不再把过期/空 expire_at 的会话算作在线。
func TestOnlineExistsExcludesExpired(t *testing.T) {
	token := fmt.Sprintf("expired-test-%d", time.Now().UnixNano())
	if err := DB.Create(&model.Online{
		TokenHash: token,
		UserID:    1,
		UserName:  "admin",
		LoginTime: time.Now().Add(-2 * time.Hour),
		ExpireAt:  time.Now().Add(-time.Hour),
	}).Error; err != nil {
		t.Fatalf("create expired online: %v", err)
	}
	t.Cleanup(func() {
		DB.Where("token_hash = ?", token).Delete(&model.Online{})
	})

	exists, err := OnlineExists(token)
	if err != nil {
		t.Fatalf("OnlineExists: %v", err)
	}
	if exists {
		t.Fatalf("过期会话不应判定为在线")
	}
}

// TestStartOnlineCleanup 验证后台清理会删除过期会话并能随 stop 退出。
func TestStartOnlineCleanup(t *testing.T) {
	token := fmt.Sprintf("cleanup-test-%d", time.Now().UnixNano())
	if err := DB.Create(&model.Online{
		TokenHash: token,
		UserID:    1,
		UserName:  "admin",
		LoginTime: time.Now().Add(-2 * time.Hour),
		ExpireAt:  time.Now().Add(-time.Hour),
	}).Error; err != nil {
		t.Fatalf("create expired online: %v", err)
	}
	t.Cleanup(func() {
		DB.Where("token_hash = ?", token).Delete(&model.Online{})
	})

	stop := make(chan struct{})
	StartOnlineCleanup(stop, 20*time.Millisecond)
	deadline := time.Now().Add(2 * time.Second)
	for {
		var cnt int64
		if err := DB.Model(&model.Online{}).Where("token_hash = ?", token).Count(&cnt).Error; err != nil {
			t.Fatalf("count: %v", err)
		}
		if cnt == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("过期会话未被后台清理")
		}
		time.Sleep(20 * time.Millisecond)
	}
	close(stop)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := WaitBackground(ctx); err != nil {
		t.Fatalf("WaitBackground: %v", err)
	}
}

// TestOnlineExistsActive 验证未过期会话仍可正常通过。
func TestOnlineExistsActive(t *testing.T) {
	token := fmt.Sprintf("active-test-%d", time.Now().UnixNano())
	if err := DB.Create(&model.Online{
		TokenHash: token,
		UserID:    1,
		UserName:  "admin",
		LoginTime: time.Now(),
		ExpireAt:  time.Now().Add(time.Hour),
	}).Error; err != nil {
		t.Fatalf("create active online: %v", err)
	}
	t.Cleanup(func() {
		DB.Where("token_hash = ?", token).Delete(&model.Online{})
	})

	exists, err := OnlineExists(token)
	if err != nil {
		t.Fatalf("OnlineExists: %v", err)
	}
	if !exists {
		t.Fatalf("未过期会话应判定为在线")
	}
}
