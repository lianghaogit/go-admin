package dao

import (
	"path/filepath"
	"testing"
	"time"

	"go-admin/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// TestSQLiteTimeColumnsRoundTrip 验证时间列改成 time.Time 后 SQLite 侧仍可读写
// （SQLite 用 TEXT 存时间，模型换成 time.Time 后必须确认驱动能正确往返）
func TestSQLiteTimeColumnsRoundTrip(t *testing.T) {
	old := DB
	defer func() { DB = old }()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "time.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开 sqlite 失败: %v", err)
	}
	DB = db
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close() // Windows 下不关闭会锁住临时库文件，导致 TempDir 清理失败
		}
	}()
	if err := DB.AutoMigrate(&model.Api{}, &model.GenTable{}); err != nil {
		t.Fatalf("建表失败: %v", err)
	}

	now := time.Date(2026, 11, 14, 12, 30, 0, 0, time.Local)
	api := model.Api{APIName: "探针", Path: "/api/probe", Method: "GET", CreateTime: model.LocalTime(now)}
	if err := DB.Create(&api).Error; err != nil {
		t.Fatalf("写入 sys_api 失败: %v", err)
	}
	var got model.Api
	if err := DB.First(&got, api.APIID).Error; err != nil {
		t.Fatalf("读取 sys_api 失败: %v", err)
	}
	if got.CreateTime.Time().IsZero() || got.CreateTime.Time().Unix() != now.Unix() {
		t.Fatalf("sys_api.create_time 往返不一致: %v", got.CreateTime.Time())
	}

	gt := model.GenTable{TblName: "probe_table", TableComment: "探针", CreateTime: model.LocalTime(now), UpdateTime: model.LocalTime(now)}
	if err := DB.Create(&gt).Error; err != nil {
		t.Fatalf("写入 sys_gen_table 失败: %v", err)
	}
	var gotGt model.GenTable
	if err := DB.First(&gotGt, gt.TableID).Error; err != nil {
		t.Fatalf("读取 sys_gen_table 失败: %v", err)
	}
	if gotGt.UpdateTime.Time().Unix() != now.Unix() {
		t.Fatalf("sys_gen_table.update_time 往返不一致: %v", gotGt.UpdateTime.Time())
	}
}

// TestSQLiteLegacyTextTimeReadable 历史 SQLite 库里时间是 TEXT('YYYY-MM-DD HH:MM:SS')，
// 模型改成 time.Time 后仍需能读出旧数据（不能因类型变更导致老库读不出）
func TestSQLiteLegacyTextTimeReadable(t *testing.T) {
	old := DB
	defer func() { DB = old }()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "legacy.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开 sqlite 失败: %v", err)
	}
	DB = db
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}()
	// 模拟旧库：TEXT 列 + 字符串时间
	if err := DB.Exec(`CREATE TABLE sys_api (
		api_id integer primary key autoincrement,
		api_name text, path text, method text, perms text, status text, remark text,
		create_time text)`).Error; err != nil {
		t.Fatalf("建旧表失败: %v", err)
	}
	if err := DB.Exec(`INSERT INTO sys_api (api_name, path, method, create_time) VALUES ('旧接口', '/api/old', 'GET', '2026-11-05 12:00:00')`).Error; err != nil {
		t.Fatalf("插入旧数据失败: %v", err)
	}
	var got model.Api
	if err := DB.First(&got, 1).Error; err != nil {
		t.Fatalf("读取旧数据失败: %v", err)
	}
	if got.CreateTime.Time().Year() != 2026 || got.CreateTime.Time().Month() != 11 || got.CreateTime.Time().Day() != 5 {
		t.Fatalf("旧 TEXT 时间未正确解析: %v", got.CreateTime.Time())
	}
}
