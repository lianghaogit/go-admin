package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 全局配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	JWT      JWTConfig      `yaml:"jwt"`
	Captcha  CaptchaConfig  `yaml:"captcha"`
	Database DatabaseConfig `yaml:"database"`
	Logs     LogsConfig     `yaml:"logs"`
}

// LogsConfig 日志保留策略
type LogsConfig struct {
	// RetentionDays 操作日志/登录日志/任务日志保留天数；0 或负数表示不自动清理（默认）
	RetentionDays int `yaml:"retentionDays"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port      int    `yaml:"port"`
	Mode      string `yaml:"mode"`
	UploadDir string `yaml:"-"` // 上传目录（运行时解析，不读配置）
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret string `yaml:"secret"`
	Expire int    `yaml:"expire"` // 分钟
}

// CaptchaConfig 验证码配置
type CaptchaConfig struct {
	Enabled bool `yaml:"enabled"`
}

// DatabaseConfig 数据库配置
// Type 支持三种：sqlite（默认，纯 Go 驱动无需 CGO）/ mysql / postgres
type DatabaseConfig struct {
	Type     string `yaml:"type"`     // sqlite / mysql / postgres
	Path     string `yaml:"path"`     // SQLite 数据库文件路径（type=sqlite 时生效）
	Host     string `yaml:"host"`     // MySQL / PostgreSQL 主机
	Port     int    `yaml:"port"`     // MySQL / PostgreSQL 端口
	User     string `yaml:"user"`     // MySQL / PostgreSQL 用户名
	Password string `yaml:"password"` // MySQL / PostgreSQL 密码
	DBName   string `yaml:"dbname"`   // MySQL / PostgreSQL 数据库名
	// Timezone 业务时区（IANA 名称，默认 Asia/Shanghai）。
	// 同时用于连接串会话时区与按天/按小时统计的显式时区转换；
	// 不配置时会话时区取数据库默认，日期边界会偏移。
	Timezone string `yaml:"timezone"`
	// MaxOpenConns 连接池最大连接数（默认 25）
	MaxOpenConns int `yaml:"maxOpenConns"`
	// MaxIdleConns 连接池空闲连接数（默认 10）
	MaxIdleConns int `yaml:"maxIdleConns"`
	// ConnMaxLifetimeMinutes 单连接最长存活分钟数（默认 60）
	ConnMaxLifetimeMinutes int `yaml:"connMaxLifetimeMinutes"`
	// SSLMode PostgreSQL sslmode（disable/allow/prefer/require/verify-ca/verify-full，默认 disable）
	SSLMode string `yaml:"sslmode"`
	// AutoMigrate 启动时是否自动建表/迁移（nil 视为 true）。
	// 生产环境可设 false，改由 DBA 执行迁移脚本，避免启动即改表结构
	AutoMigrate *bool `yaml:"autoMigrate"`
	// SeedData 启动时是否写入/补齐种子数据与生成模块菜单（nil 视为 true）。
	// 与 AutoMigrate 解耦：生产关闭自动迁移后，仍可单独控制是否跑种子
	SeedData *bool `yaml:"seedData"`
	// DropForeignKeys 启动时是否删除老库中的主外键约束（nil 视为 false）。
	// 默认不删；需要清理历史外键时显式开启，或放到迁移脚本中执行
	DropForeignKeys *bool `yaml:"dropForeignKeys"`
}

// Cfg 全局配置实例
var Cfg Config

// Load 加载配置文件；JWT 密钥可用环境变量 GO_ADMIN_JWT_SECRET 覆盖（优先于配置文件）
func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, &Cfg); err != nil {
		return err
	}
	// 兼容旧配置：未声明 type 时默认 sqlite
	if Cfg.Database.Type == "" {
		Cfg.Database.Type = "sqlite"
	}
	// 兼容旧配置：未声明时区时默认 Asia/Shanghai
	if Cfg.Database.Timezone == "" {
		Cfg.Database.Timezone = "Asia/Shanghai"
	}
	// 连接池与 TLS 默认值
	if Cfg.Database.MaxOpenConns <= 0 {
		Cfg.Database.MaxOpenConns = 25
	}
	if Cfg.Database.MaxIdleConns <= 0 {
		Cfg.Database.MaxIdleConns = 10
	}
	if Cfg.Database.ConnMaxLifetimeMinutes <= 0 {
		Cfg.Database.ConnMaxLifetimeMinutes = 60
	}
	if Cfg.Database.SSLMode == "" {
		Cfg.Database.SSLMode = "disable"
	}
	if env := os.Getenv("GO_ADMIN_JWT_SECRET"); env != "" {
		Cfg.JWT.Secret = env
	}
	return nil
}
