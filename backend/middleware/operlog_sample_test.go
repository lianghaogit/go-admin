package middleware

import (
	"net/http"
	"testing"
)

func TestShouldRecordGET(t *testing.T) {
	oldEvery, oldSampled := operLogGetSampleEvery, operLogGetSampled
	defer func() {
		operLogGetSampleEvery, operLogGetSampled = oldEvery, oldSampled
	}()

	if !shouldRecordGET(http.MethodPost, "/api/system/user", StatusSuccess, 0) {
		t.Fatal("非 GET 必须记录")
	}
	if !shouldRecordGET(http.MethodGet, "/api/system/user/list", StatusFail, 0) {
		t.Fatal("失败 GET 必须记录")
	}
	if !shouldRecordGET(http.MethodGet, "/api/system/user/list", StatusSuccess, operLogSlowGETMS) {
		t.Fatal("慢 GET 必须记录")
	}
	if !shouldRecordGET(http.MethodGet, "/api/system/user/export", StatusSuccess, 0) {
		t.Fatal("导出 GET 必须记录")
	}

	operLogGetSampleEvery, operLogGetSampled = 2, 0
	if shouldRecordGET(http.MethodGet, "/api/system/user/list", StatusSuccess, 0) {
		t.Fatal("1/2 采样第 1 次不应记录")
	}
	if !shouldRecordGET(http.MethodGet, "/api/system/user/list", StatusSuccess, 0) {
		t.Fatal("1/2 采样第 2 次应记录")
	}

	operLogGetSampleEvery, operLogGetSampled = 1, 0
	if !shouldRecordGET(http.MethodGet, "/api/system/user/list", StatusSuccess, 0) {
		t.Fatal("采样率 1 应记录全部 GET")
	}
}
