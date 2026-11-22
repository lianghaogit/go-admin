package dao

import (
	"testing"

	"go-admin/model"
)

// TestAfterOpenSeedWithoutAutoMigrate 验证种子数据已与自动迁移解耦：
// 关闭 AutoMigrate 后，仍可单独执行 seed（生成模块菜单不会随之丢失）。
func TestAfterOpenSeedWithoutAutoMigrate(t *testing.T) {
	oldAuto, oldSeed, oldDrop := autoMigrateEnabled, seedDataEnabled, dropForeignKeysEnabled
	t.Cleanup(func() {
		autoMigrateEnabled, seedDataEnabled, dropForeignKeysEnabled = oldAuto, oldSeed, oldDrop
	})

	autoMigrateEnabled = false
	seedDataEnabled = true
	dropForeignKeysEnabled = false
	if err := afterOpen(); err != nil {
		t.Fatalf("autoMigrate=false + seedData=true 应可启动: %v", err)
	}
}

// TestAfterOpenWithoutAutoMigrateDoesNotCreateTable 验证关闭 AutoMigrate 后确实不建表，
// 而不只是 afterOpen 返回 nil。
func TestAfterOpenWithoutAutoMigrateDoesNotCreateTable(t *testing.T) {
	oldAuto, oldSeed, oldDrop := autoMigrateEnabled, seedDataEnabled, dropForeignKeysEnabled
	t.Cleanup(func() {
		// 恢复测试库结构，避免影响同包其他用例
		_ = DB.AutoMigrate(&model.Notice{})
		autoMigrateEnabled, seedDataEnabled, dropForeignKeysEnabled = oldAuto, oldSeed, oldDrop
	})

	if err := DB.Migrator().DropTable(&model.Notice{}); err != nil {
		t.Fatalf("drop sys_notice: %v", err)
	}
	autoMigrateEnabled = false
	seedDataEnabled = false
	dropForeignKeysEnabled = false
	if err := afterOpen(); err != nil {
		t.Fatalf("autoMigrate=false 应可启动: %v", err)
	}
	if DB.Migrator().HasTable(&model.Notice{}) {
		t.Fatalf("autoMigrate=false 时不应重新创建 sys_notice")
	}
}
