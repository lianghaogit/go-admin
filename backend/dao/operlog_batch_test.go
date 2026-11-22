package dao

import (
	"fmt"
	"testing"
	"time"

	"go-admin/model"
)

// TestInsertOperLogs 验证中间件队列使用的批量写入路径能落库。
func TestInsertOperLogs(t *testing.T) {
	title := fmt.Sprintf("batch-test-%d", time.Now().UnixNano())
	rid := fmt.Sprintf("rid-%d", time.Now().UnixNano())
	logs := []*model.OperLog{
		{Title: title, RequestID: rid, RequestMethod: "POST", OperName: "batch-test", Status: "0", OperTime: time.Now()},
		{Title: title, RequestID: rid, RequestMethod: "POST", OperName: "batch-test", Status: "0", OperTime: time.Now()},
	}
	if err := InsertOperLogs(logs); err != nil {
		t.Fatalf("InsertOperLogs: %v", err)
	}
	t.Cleanup(func() {
		DB.Where("title = ?", title).Delete(&model.OperLog{})
	})

	var count int64
	if err := DB.Model(&model.OperLog{}).Where("title = ?", title).Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != int64(len(logs)) {
		t.Fatalf("批量写入条数 = %d，期望 %d", count, len(logs))
	}
	var one model.OperLog
	if err := DB.Where("title = ?", title).First(&one).Error; err != nil {
		t.Fatalf("查询 request_id: %v", err)
	}
	if one.RequestID != rid {
		t.Fatalf("request_id = %q，期望 %q", one.RequestID, rid)
	}
}
