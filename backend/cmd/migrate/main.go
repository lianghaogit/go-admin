package main

import (
	"flag"
	"log"
	"os"

	"go-admin/config"
	"go-admin/dao"
)

func boolPtr(v bool) *bool { return &v }

// 独立迁移入口：只跑表结构/索引，不启动 HTTP、不跑种子（可用 -seed 显式开启）。
// 用法：go run ./cmd/migrate -config config.yaml
func main() {
	configPath := flag.String("config", "", "配置文件路径，默认查找 config.yaml / backend/config.yaml")
	withSeed := flag.Bool("seed", false, "是否同时写入/补齐种子数据")
	dropFK := flag.Bool("drop-fk", false, "是否删除老库外键（危险，默认关闭）")
	flag.Parse()

	path := *configPath
	if path == "" {
		for _, p := range []string{"config.yaml", "backend/config.yaml"} {
			if _, err := os.Stat(p); err == nil {
				path = p
				break
			}
		}
	}
	if path == "" {
		log.Fatal("未找到配置文件，请用 -config 指定")
	}
	if err := config.Load(path); err != nil {
		log.Fatalf("读取配置失败: %v", err)
	}
	// 迁移入口显式打开 AutoMigrate，种子默认关闭；不删除外键。
	config.Cfg.Database.AutoMigrate = boolPtr(true)
	config.Cfg.Database.SeedData = boolPtr(*withSeed)
	config.Cfg.Database.DropForeignKeys = boolPtr(*dropFK)

	if err := dao.InitDB(config.Cfg.Database); err != nil {
		log.Fatalf("迁移失败: %v", err)
	}
	if sqlDB, err := dao.DB.DB(); err == nil {
		_ = sqlDB.Close()
	}
	log.Printf("迁移完成: %s（seed=%v, dropFK=%v）", path, *withSeed, *dropFK)
}
