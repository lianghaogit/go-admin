package dao

import "testing"

// TestIsSuperAdminName 验证超管判定只认内置账号名。
func TestIsSuperAdminName(t *testing.T) {
	if !IsSuperAdminName("admin") {
		t.Fatal("admin 应判定为超管")
	}
	if IsSuperAdminName("Admin") || IsSuperAdminName("admin2") || IsSuperAdminName("") {
		t.Fatal("非内置账号名不应判定为超管")
	}
}

// TestGetUserAuthInfoSkipsPreloads 防止鉴权查询退回“全量用户 + 三个 Preload”。
func TestGetUserAuthInfoSkipsPreloads(t *testing.T) {
	full, err := GetUserByID(flowTestAdmin.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if len(full.Roles) == 0 {
		t.Fatal("测试前提不成立：admin 应至少有一个角色")
	}

	lean, err := GetUserAuthInfo(flowTestAdmin.ID)
	if err != nil {
		t.Fatalf("GetUserAuthInfo: %v", err)
	}
	if lean.Username != full.Username || lean.Status != full.Status {
		t.Fatalf("鉴权字段不一致: got username=%q status=%q", lean.Username, lean.Status)
	}
	if lean.Dept != nil || len(lean.Roles) != 0 || len(lean.Posts) != 0 {
		t.Fatalf("鉴权查询不应预加载关联: dept=%v roles=%d posts=%d", lean.Dept, len(lean.Roles), len(lean.Posts))
	}
}
