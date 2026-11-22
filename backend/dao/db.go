package dao

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-admin/config"
	"go-admin/model"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库实例
var DB *gorm.DB

// DbPath 数据库标识（SQLite 为文件路径，MySQL/PG 为 host:port/dbname，供数据监控展示）
var DbPath string

// autoMigrateEnabled 是否允许启动时自动建表/迁移（配置 database.autoMigrate，默认 true）
var autoMigrateEnabled = true

// seedDataEnabled 是否允许启动时写入/补齐种子数据（配置 database.seedData，默认 true）
var seedDataEnabled = true

// dropForeignKeysEnabled 是否允许启动时删除老库外键（配置 database.dropForeignKeys，默认 false）
var dropForeignKeysEnabled = false

// pgSSLMode 允许的 PostgreSQL sslmode 取值（白名单，避免拼接 DSN 时被注入）
var pgSSLMode = map[string]bool{
	"disable": true, "allow": true, "prefer": true,
	"require": true, "verify-ca": true, "verify-full": true,
}

// Dialect 返回当前数据库方言：sqlite / mysql / postgres
func Dialect() string {
	return DB.Dialector.Name()
}

// InitDB 初始化数据库连接、建表并写入种子数据
// 支持三种方言：
//   - sqlite：glebarez/sqlite 纯 Go 驱动，无需 CGO，Windows 无 gcc 也可编译运行
//   - mysql / postgres：标准 gorm 驱动，需目标数据库服务可用
func InitDB(cfg config.DatabaseConfig) error {
	if err := initBizTimezone(cfg.Timezone); err != nil {
		return err
	}
	// 每次初始化都从默认值开始，避免同进程内多次 InitDB（测试/切库）沿用上一次开关。
	autoMigrateEnabled = true
	seedDataEnabled = true
	dropForeignKeysEnabled = false
	if cfg.AutoMigrate != nil {
		autoMigrateEnabled = *cfg.AutoMigrate
	}
	if cfg.SeedData != nil {
		seedDataEnabled = *cfg.SeedData
	}
	if cfg.DropForeignKeys != nil {
		dropForeignKeysEnabled = *cfg.DropForeignKeys
	}
	switch cfg.Type {
	case "mysql", "postgres":
		return initNetworkDB(cfg)
	default: // sqlite
		return initSQLite(cfg.Path)
	}
}

// initSQLite 初始化 SQLite（保留原有逻辑，追加生产化配置）
func initSQLite(path string) error {
	DbPath = path
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		// 表结构只保留普通列与索引，不创建主外键约束；关联由业务代码维护（老库外键由 database.dropForeignKeys 控制是否清理）
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return err
	}
	DB = db
	// SQLite 生产化：WAL 日志模式（读写不互斥，显著提升并发写入吞吐）+ busy_timeout（写锁等待 5s，避免 SQLITE_BUSY 直接报错）
	var wal string
	if err := DB.Raw("PRAGMA journal_mode=WAL").Scan(&wal).Error; err != nil {
		return fmt.Errorf("开启 WAL 失败: %w", err)
	}
	if err := DB.Exec("PRAGMA busy_timeout=5000").Error; err != nil {
		return fmt.Errorf("设置 busy_timeout 失败: %w", err)
	}
	// SQLite 单文件写锁受并发连接数影响，限制连接池规模减少锁竞争
	if sqlDB, err := DB.DB(); err == nil {
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetMaxIdleConns(5)
	}
	return afterOpen()
}

// initNetworkDB 初始化 MySQL / PostgreSQL
func initNetworkDB(cfg config.DatabaseConfig) error {
	host := cfg.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := cfg.Port
	if port == 0 {
		if cfg.Type == "postgres" {
			port = 5432
		} else {
			port = 3306
		}
	}
	DbPath = fmt.Sprintf("%s:%d/%s", host, port, cfg.DBName)
	var dialector gorm.Dialector
	switch cfg.Type {
	case "postgres":
		sslMode := strings.ToLower(strings.TrimSpace(cfg.SSLMode))
		if sslMode == "" {
			sslMode = "disable"
		}
		if !pgSSLMode[sslMode] {
			return fmt.Errorf("database.sslmode 取值非法: %q（可选 disable/allow/prefer/require/verify-ca/verify-full）", cfg.SSLMode)
		}
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
			host, port, cfg.User, cfg.Password, cfg.DBName, sslMode, tzName())
		dialector = postgres.Open(dsn)
	default: // mysql
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=%s",
			cfg.User, cfg.Password, host, port, cfg.DBName, url.QueryEscape(tzName()))
		dialector = mysql.Open(dsn)
	}
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		// 表结构只保留普通列与索引，不创建主外键约束；关联由业务代码维护（老库外键由 database.dropForeignKeys 控制是否清理）
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	// 连接池参数来自配置（默认 25/10/60 分钟），避免默认的无上限连接在高并发下打满数据库
	maxOpen := cfg.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 25
	}
	maxIdle := cfg.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 10
	}
	if maxIdle > maxOpen {
		maxIdle = maxOpen
	}
	lifetime := cfg.ConnMaxLifetimeMinutes
	if lifetime <= 0 {
		lifetime = 60
	}
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(time.Duration(lifetime) * time.Minute)
	DB = db
	return afterOpen()
}

// afterOpen 建表、种子数据、方言化数据回填与接口规则加载
func afterOpen() error {
	// database.autoMigrate=false 时跳过建表/迁移；种子与索引拆成独立开关，
	// 避免关闭自动迁移后连生成模块菜单也不创建。
	if autoMigrateEnabled {
		// 先把历史 varchar 时间列转成时间类型（PG 需要 USING，AutoMigrate 直接改会失败）
		if err := migrateLegacyTimeColumns(); err != nil {
			return err
		}
		if err := DB.AutoMigrate(&model.User{}, &model.Role{}, &model.Menu{}, &model.LoginLog{}, &model.Dept{}, &model.DictType{}, &model.DictData{}, &model.OperLog{}, &model.Post{}, &model.Config{}, &model.Notice{}, &model.Online{}, &model.Job{}, &model.JobLog{}, &model.Api{}, &model.GenTable{}, &model.GenColumn{}, &model.FlowDef{}, &model.FlowNode{}, &model.FlowInstance{}, &model.FlowTask{}, &model.FlowRecord{}); err != nil {
			return err
		}
	}
	if seedDataEnabled {
		if err := seed(); err != nil {
			return err
		}
	}
	if autoMigrateEnabled {
		// 性能相关索引（幂等）：补齐高频过滤列索引，PostgreSQL 另建三元组索引并 ANALYZE
		if err := ensureIndexes(); err != nil {
			return err
		}
	}
	// PostgreSQL：种子数据用固定 ID 插入，bigserial 序列不会随之推进，需与表内最大主键对齐，
	// 否则后续新增记录 nextval 会与既有主键冲突（duplicate key ... _pkey）
	if Dialect() == "postgres" {
		if err := syncPgSequences(); err != nil {
			return err
		}
		// PostgreSQL：补齐表与字段的中文注释（注释源 scripts/pg_comments.sql，幂等）
		syncPgComments()
	}
	// 表结构不保留主外键约束：即使模型里声明了关联，也不让 AutoMigrate 建 FK；
	// 老库中已有的 FK 默认保留，只有显式开启 database.dropForeignKeys 才删除。
	if dropForeignKeysEnabled {
		if err := dropForeignKeys(); err != nil {
			return err
		}
	}
	// 回填接口规则的创建时间（老数据迁移后为空），按方言使用当前时间函数
	if err := fixApiCreateTimes(); err != nil {
		return err
	}
	return ReloadApiRules()
}

// syncPgSequences 把 PostgreSQL 每张业务表的自增序列对齐到表内实际最大主键
// 背景：种子数据以固定 ID 写入（如菜单 1/100/101、岗位 1/2/3），bigserial 序列仍从 1 开始，
// 新增记录会与既有主键冲突，表现为"新增岗位失败"等 500 错误（SQLSTATE 23505）
// 实现说明：序列名只从 information_schema.columns.column_default（nextval(...)）解析，
// 不使用 pg_get_serial_sequence —— 后者在列不存在时会直接报错（如关联表 sys_role_menu 无 id 列）
func syncPgSequences() error {
	type colItem struct {
		TableName string `gorm:"column:table_name"`
		Default   string `gorm:"column:column_default"`
	}
	var items []colItem
	err := DB.Raw(`SELECT table_name, column_default FROM information_schema.columns
		WHERE table_schema = 'public' AND column_name = 'id' AND column_default LIKE 'nextval(%'`).Scan(&items).Error
	if err != nil {
		return fmt.Errorf("查询 PG 序列失败: %w", err)
	}
	for _, it := range items {
		seq := seqNameFromDefault(it.Default)
		if seq == "" {
			continue
		}
		var maxID int64
		if err := DB.Raw("SELECT COALESCE(MAX(id), 0) FROM " + it.TableName).Scan(&maxID).Error; err != nil {
			return fmt.Errorf("查询 %s 最大主键失败: %w", it.TableName, err)
		}
		if maxID <= 0 {
			continue
		}
		if err := DB.Exec("SELECT setval('"+seq+"', ?, true)", maxID).Error; err != nil {
			return fmt.Errorf("重置 %s 序列失败: %w", it.TableName, err)
		}
	}
	return nil
}

// syncPgComments 为 PostgreSQL 的表与字段补充中文注释
// 注释定义在 scripts/pg_comments.sql（每条语句一行，以 COMMENT ON 开头），逐条执行、幂等；
// 文件不存在或单条失败都不阻断启动（注释属于元数据增强，不应影响服务可用性）
// 背景：SQLite 没有表/列注释属性，迁移到 PG 后 information_schema 与客户端里全部为空
func syncPgComments() {
	var candidates []string
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "scripts", "pg_comments.sql"))
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "scripts", "pg_comments.sql"))
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err != nil {
			continue
		}
		f, err := os.Open(p)
		if err != nil {
			fmt.Println("[db] 打开注释文件失败:", err)
			return
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 1024*1024), 1024*1024)
		ok, failed := 0, 0
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || !strings.HasPrefix(line, "COMMENT ON") {
				continue
			}
			if err := DB.Exec(strings.TrimSuffix(line, ";")).Error; err != nil {
				failed++
				continue
			}
			ok++
		}
		if failed > 0 {
			fmt.Printf("[db] PG 注释同步完成：成功 %d 条，失败 %d 条（%s）\n", ok, failed, p)
		}
		return
	}
}

// seqNameFromDefault 从 nextval('sys_post_id_seq'::regclass) 中取出序列名
func seqNameFromDefault(def string) string {
	i := strings.Index(def, "'")
	if i < 0 {
		return ""
	}
	j := strings.Index(def[i+1:], "'")
	if j < 0 {
		return ""
	}
	return def[i+1 : i+1+j]
}

// migrateLegacyTimeColumns 把历史遗留的 varchar 时间列转成原生时间类型（幂等）
// 背景：早期版本 sys_api.create_time、sys_gen_table.create_time/update_time 用 varchar(20)
// 存 "YYYY-MM-DD HH:MM:SS"，在 PostgreSQL 中无法参与时间比较/排序，也无法排序稳定。
// PostgreSQL 直接 ALTER TYPE 需要 USING 子句，AutoMigrate 不提供，故在此显式转换。
func migrateLegacyTimeColumns() error {
	targets := []struct {
		Table  string
		Column string
	}{
		{"sys_api", "create_time"},
		{"sys_gen_table", "create_time"},
		{"sys_gen_table", "update_time"},
	}
	switch Dialect() {
	case "postgres":
		for _, t := range targets {
			var dataType string
			err := DB.Raw(`SELECT data_type FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = ? AND column_name = ?`,
				t.Table, t.Column).Scan(&dataType).Error
			if err != nil || dataType == "" {
				continue
			}
			if dataType != "character varying" && dataType != "text" {
				continue // 已是时间类型
			}
			stmt := fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN %s TYPE timestamptz USING NULLIF(btrim(%s), '')::timestamptz`,
				t.Table, t.Column, t.Column)
			if err := DB.Exec(stmt).Error; err != nil {
				return fmt.Errorf("转换 %s.%s 为 timestamptz 失败: %w", t.Table, t.Column, err)
			}
		}
	case "mysql":
		for _, t := range targets {
			var dataType string
			err := DB.Raw(`SELECT data_type FROM information_schema.columns
				WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?`,
				t.Table, t.Column).Scan(&dataType).Error
			if err != nil || dataType == "" || dataType == "datetime" || dataType == "timestamp" {
				continue
			}
			stmt := fmt.Sprintf(`ALTER TABLE %s MODIFY COLUMN %s DATETIME NULL`, t.Table, t.Column)
			if err := DB.Exec(stmt).Error; err != nil {
				return fmt.Errorf("转换 %s.%s 为 DATETIME 失败: %w", t.Table, t.Column, err)
			}
		}
	}
	return nil
}

// fixApiCreateTimes 回填 sys_api 创建时间（history 迁移后为空的行）
// 列已是时间类型，直接使用方言当前时间函数，不再格式化成字符串
func fixApiCreateTimes() error {
	var expr string
	switch Dialect() {
	case "mysql":
		expr = "NOW()"
	case "postgres":
		expr = "now()"
	default:
		expr = "datetime('now')"
	}
	return DB.Exec("UPDATE sys_api SET create_time = " + expr + " WHERE create_time IS NULL").Error
}
