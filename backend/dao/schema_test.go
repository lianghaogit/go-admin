package dao

import (
	"strings"
	"testing"

	"go-admin/model"
)

// TestSchemaHasNoForeignKeyConstraints 校验表结构不建主外键关系：
// 建表 DDL 中不应出现 FOREIGN KEY 声明（关联由业务代码维护）。
func TestSchemaHasNoForeignKeyConstraints(t *testing.T) {
	tables, err := sqliteTables()
	if err != nil {
		t.Fatalf("读取 sqlite_master 失败: %v", err)
	}
	if len(tables) == 0 {
		t.Fatal("未读到任何业务表")
	}
	for _, tb := range tables {
		var ddl string
		if err := DB.Raw("SELECT sql FROM sqlite_master WHERE type='table' AND name = ?", tb).Scan(&ddl).Error; err != nil {
			t.Fatalf("读取 %s 建表语句失败: %v", tb, err)
		}
		if strings.Contains(strings.ToUpper(ddl), "FOREIGN KEY") {
			t.Fatalf("表 %s 仍含 FOREIGN KEY 约束: %s", tb, ddl)
		}
	}
}

// TestDropForeignKeysSQLiteNoop 验证 SQLite 分支不报错（约束无法原地删除，仅统计提示）
func TestDropForeignKeysSQLiteNoop(t *testing.T) {
	if err := dropForeignKeys(); err != nil {
		t.Fatalf("SQLite dropForeignKeys 不应返回错误: %v", err)
	}
}

// TestDeleteUserClearsAssociations 验证删除用户时应用层清理角色/岗位关联（不依赖数据库外键）
func TestDeleteUserClearsAssociations(t *testing.T) {
	post := model.Post{PostCode: "fk_probe_post", PostName: "外键探针岗位", Status: "0"}
	if err := DB.Create(&post).Error; err != nil {
		t.Fatalf("创建岗位失败: %v", err)
	}
	defer DB.Where("post_code = ?", "fk_probe_post").Delete(&model.Post{})

	user := model.User{Username: "fk_probe_user", Nickname: "外键探针", Password: "x", Status: "0", DeptID: 100}
	if err := DB.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	role := model.Role{ID: 1}
	if err := DB.Model(&user).Association("Roles").Append(&role); err != nil {
		t.Fatalf("绑定角色失败: %v", err)
	}
	if err := DB.Model(&user).Association("Posts").Append(&post); err != nil {
		t.Fatalf("绑定岗位失败: %v", err)
	}

	if err := DeleteUser(user.ID); err != nil {
		t.Fatalf("删除用户失败: %v", err)
	}

	var userCnt, postCnt int64
	DB.Table("sys_user_role").Where("user_id = ?", user.ID).Count(&userCnt)
	DB.Table("sys_user_post").Where("user_id = ?", user.ID).Count(&postCnt)
	if userCnt != 0 || postCnt != 0 {
		t.Fatalf("删除用户后仍有关联残留: user_role=%d user_post=%d", userCnt, postCnt)
	}
}

// TestDeleteDeptClearsRoleDept 验证删除部门时清理角色的部门数据权限关联
func TestDeleteDeptClearsRoleDept(t *testing.T) {
	dept := model.Dept{DeptName: "外键探针部门", ParentID: 100, Status: "0"}
	if err := DB.Create(&dept).Error; err != nil {
		t.Fatalf("创建部门失败: %v", err)
	}
	if err := DB.Exec("INSERT INTO sys_role_dept (role_id, dept_id) VALUES (?, ?)", 1, dept.ID).Error; err != nil {
		t.Fatalf("写入角色部门关联失败: %v", err)
	}
	if err := DeleteDept(dept.ID); err != nil {
		t.Fatalf("删除部门失败: %v", err)
	}
	var cnt int64
	DB.Table("sys_role_dept").Where("dept_id = ?", dept.ID).Count(&cnt)
	if cnt != 0 {
		t.Fatalf("删除部门后 sys_role_dept 仍有 %d 行残留", cnt)
	}
}
