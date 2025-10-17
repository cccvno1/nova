package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 系统总配置
// 包含服务器、日志、数据库、Redis、认证、限流、权限、上传、队列、审计日志等模块配置
// 支持通过 NOVA_ 前缀的环境变量覆盖配置项
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`    // 服务器配置
	Logger    LoggerConfig    `mapstructure:"logger"`    // 日志配置
	DB        DBConfig        `mapstructure:"database"`  // 数据库配置
	Redis     RedisConfig     `mapstructure:"redis"`     // Redis配置
	Auth      AuthConfig      `mapstructure:"auth"`      // 认证配置
	RateLimit RateLimitConfig `mapstructure:"ratelimit"` // 限流配置
	Casbin    CasbinConfig    `mapstructure:"casbin"`    // Casbin权限配置
	Upload    UploadConfig    `mapstructure:"upload"`    // 文件上传配置
	Queue     QueueConfig     `mapstructure:"queue"`     // 队列配置
	AuditLog  AuditLogConfig  `mapstructure:"audit_log"` // 审计日志配置
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `mapstructure:"host"` // 监听地址，如 "0.0.0.0" 或 "127.0.0.1"
	Port int    `mapstructure:"port"` // 监听端口，默认8080
	Mode string `mapstructure:"mode"` // 运行模式：debug/release/test
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      string `mapstructure:"level"`       // 日志级别：debug/info/warn/error
	Format     string `mapstructure:"format"`      // 日志格式：json/text
	Output     string `mapstructure:"output"`      // 输出目标：stdout/file/both
	FilePath   string `mapstructure:"file_path"`   // 日志文件路径
	MaxSize    int    `mapstructure:"max_size"`    // 单个日志文件最大大小（MB）
	MaxBackups int    `mapstructure:"max_backups"` // 保留的旧日志文件数量
	MaxAge     int    `mapstructure:"max_age"`     // 日志文件保留天数
}

// DBConfig 数据库配置
type DBConfig struct {
	Driver   string `mapstructure:"driver"`   // 数据库驱动：postgres/mysql
	Host     string `mapstructure:"host"`     // 数据库主机地址
	Port     int    `mapstructure:"port"`     // 数据库端口
	User     string `mapstructure:"user"`     // 数据库用户名
	Password string `mapstructure:"password"` // 数据库密码
	DBName   string `mapstructure:"dbname"`   // 数据库名称
	Charset  string `mapstructure:"charset"`  // 字符集
	MaxIdle  int    `mapstructure:"max_idle"` // 最大空闲连接数
	MaxOpen  int    `mapstructure:"max_open"` // 最大打开连接数
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret            string `mapstructure:"jwt_secret"`             // JWT密钥，生产环境必须修改
	AccessTokenDuration  int    `mapstructure:"access_token_duration"`  // 访问令牌有效期（秒），默认2小时
	RefreshTokenDuration int    `mapstructure:"refresh_token_duration"` // 刷新令牌有效期（秒），默认7天
	Issuer               string `mapstructure:"issuer"`                 // JWT签发者标识
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string `mapstructure:"host"`           // Redis主机地址
	Port         int    `mapstructure:"port"`           // Redis端口
	Password     string `mapstructure:"password"`       // Redis密码（可选）
	DB           int    `mapstructure:"db"`             // 使用的数据库编号（0-15）
	PoolSize     int    `mapstructure:"pool_size"`      // 连接池大小
	MinIdleConns int    `mapstructure:"min_idle_conns"` // 最小空闲连接数
	MaxRetries   int    `mapstructure:"max_retries"`    // 命令最大重试次数
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled    bool   `mapstructure:"enabled"`     // 是否启用限流
	Algorithm  string `mapstructure:"algorithm"`   // 限流算法：token_bucket（令牌桶）或 sliding_window（滑动窗口）
	IPLimit    int    `mapstructure:"ip_limit"`    // 每个IP的请求限制（次数）
	IPWindow   int    `mapstructure:"ip_window"`   // IP限流时间窗口（秒）
	UserLimit  int    `mapstructure:"user_limit"`  // 每个用户的请求限制（次数）
	UserWindow int    `mapstructure:"user_window"` // 用户限流时间窗口（秒）
}

// CasbinConfig Casbin权限配置
type CasbinConfig struct {
	ModelPath    string `mapstructure:"model_path"`     // RBAC 模型文件路径（rbac_model.conf）
	AutoSave     bool   `mapstructure:"auto_save"`      // 是否自动保存策略到数据库
	AutoLoad     bool   `mapstructure:"auto_load"`      // 是否定期从数据库重新加载策略（用于多实例同步）
	AutoLoadTick int    `mapstructure:"auto_load_tick"` // 自动加载策略的间隔时间（秒）
}

// UploadConfig 文件上传配置
type UploadConfig struct {
	// 基础配置
	StorageType  string   `mapstructure:"storage_type"`  // 存储类型：local（本地）、oss（阿里云OSS）、s3（AWS S3/MinIO）
	MaxSize      int64    `mapstructure:"max_size"`      // 单文件最大大小（MB）
	AllowedTypes []string `mapstructure:"allowed_types"` // 允许的 MIME 类型列表（如 image/jpeg）
	AllowedExts  []string `mapstructure:"allowed_exts"`  // 允许的文件扩展名列表（如 .jpg）

	// 本地存储配置
	LocalPath string `mapstructure:"local_path"` // 本地存储路径（相对于项目根目录）
	LocalURL  string `mapstructure:"local_url"`  // 本地访问 URL 前缀（用于生成下载链接）

	// 缩略图配置
	EnableThumbnail  bool `mapstructure:"enable_thumbnail"`  // 是否为图片生成缩略图
	ThumbnailWidth   int  `mapstructure:"thumbnail_width"`   // 缩略图宽度（像素）
	ThumbnailHeight  int  `mapstructure:"thumbnail_height"`  // 缩略图高度（像素）
	ThumbnailQuality int  `mapstructure:"thumbnail_quality"` // 缩略图质量（1-100）

	// OSS 配置（阿里云对象存储）
	OSSEndpoint        string `mapstructure:"oss_endpoint"`          // OSS访问端点（如 oss-cn-hangzhou.aliyuncs.com）
	OSSAccessKeyID     string `mapstructure:"oss_access_key_id"`     // OSS访问密钥ID
	OSSAccessKeySecret string `mapstructure:"oss_access_key_secret"` // OSS访问密钥Secret
	OSSBucketName      string `mapstructure:"oss_bucket_name"`       // OSS存储桶名称
	OSSBasePath        string `mapstructure:"oss_base_path"`         // OSS基础路径前缀

	// S3 配置（AWS S3 或 MinIO 兼容存储）
	S3Endpoint        string `mapstructure:"s3_endpoint"`          // S3访问端点
	S3AccessKeyID     string `mapstructure:"s3_access_key_id"`     // S3访问密钥ID
	S3AccessKeySecret string `mapstructure:"s3_access_key_secret"` // S3访问密钥Secret
	S3BucketName      string `mapstructure:"s3_bucket_name"`       // S3存储桶名称
	S3Region          string `mapstructure:"s3_region"`            // S3区域（如 us-east-1）
	S3UseSSL          bool   `mapstructure:"s3_use_ssl"`           // 是否使用SSL连接
}

// QueueConfig 队列配置
type QueueConfig struct {
	Enabled      bool   `mapstructure:"enabled"`       // 是否启用异步任务队列
	Workers      int    `mapstructure:"workers"`       // Worker 并发数量
	MaxRetry     int    `mapstructure:"max_retry"`     // 任务失败后的最大重试次数
	RetryDelay   int    `mapstructure:"retry_delay"`   // 重试延迟时间（秒）
	RedisPrefix  string `mapstructure:"redis_prefix"`  // Redis 键前缀（用于命名空间隔离）
	PollInterval int    `mapstructure:"poll_interval"` // 队列轮询间隔（秒）
}

// AuditLogConfig 审计日志配置
type AuditLogConfig struct {
	Enabled         bool     `mapstructure:"enabled"`          // 是否启用审计日志
	LogRequest      bool     `mapstructure:"log_request"`      // 是否记录请求体内容
	LogResponse     bool     `mapstructure:"log_response"`     // 是否记录响应体内容（通常较大，建议关闭）
	MaxBodySize     int      `mapstructure:"max_body_size"`    // 请求/响应体最大记录大小（字节）
	ExcludePaths    []string `mapstructure:"exclude_paths"`    // 排除的路径列表（这些路径不记录审计日志）
	IncludeActions  []string `mapstructure:"include_actions"`  // 只记录指定动作（为空则全部记录）【注：当前中间件暂未实现此过滤】
	SensitiveFields []string `mapstructure:"sensitive_fields"` // 敏感字段名称列表（需要脱敏处理，如 password、token）
}

var globalConfig *Config

// Load 加载配置文件
// 支持从指定路径加载配置文件，或自动搜索多个目录
// 支持通过 NOVA_ 前缀的环境变量覆盖配置项（如 NOVA_SERVER_PORT）
func Load(configPath string) (*Config, error) {
	v := viper.New()

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// 自动搜索配置文件
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
		v.AddConfigPath("../configs")
		v.AddConfigPath("../../configs")
	}

	// 环境变量支持：NOVA_SERVER_PORT 映射到 server.port
	v.SetEnvPrefix("NOVA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	globalConfig = &cfg
	return &cfg, nil
}

// Get 获取全局配置实例
func Get() *Config {
	return globalConfig
}

// GetServerAddr 获取服务器监听地址（host:port格式）
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetDSN 获取数据库连接字符串（DSN）
func (c *DBConfig) GetDSN() string {
	switch c.Driver {
	case "", "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			c.Host, c.Port, c.User, c.Password, c.DBName)
	default:
		return ""
	}
}
