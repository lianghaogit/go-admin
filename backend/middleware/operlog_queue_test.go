package middleware

import (
	"testing"

	"go-admin/model"
)

// TestOperLogQueueDrop 用局部队列稳定覆盖“队列满 → 丢弃并计数”路径。
func TestOperLogQueueDrop(t *testing.T) {
	queue := make(chan *model.OperLog, 1)
	queue <- &model.OperLog{ID: 1}
	var dropped uint64

	if tryEnqueueOperLog(queue, &model.OperLog{ID: 2}, &dropped) {
		t.Fatal("队列已满时不应入队成功")
	}
	if dropped != 1 {
		t.Fatalf("dropped = %d，期望 1", dropped)
	}

	drained := <-queue
	if drained.ID != 1 {
		t.Fatalf("队列中应是先入队的日志，实际 ID = %d", drained.ID)
	}
	if !tryEnqueueOperLog(queue, &model.OperLog{ID: 3}, &dropped) {
		t.Fatal("腾空后应能入队")
	}
	if dropped != 1 {
		t.Fatalf("成功入队不应增加 dropped，实际 %d", dropped)
	}
}
