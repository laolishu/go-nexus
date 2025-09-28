package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 应用程序配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Cache    CacheConfig    `mapstructure:"cache"`
	Log      LogConfig      `mapstructure:"log"`
	Security SecurityConfig `mapstructure:"security"`
	Plugins  PluginsConfig  `mapstructure:"plugins"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	Mode           string        `mapstructure:"mode"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	MaxHeaderBytes int           `mapstructure:"max_header_bytes"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type            string        `mapstructure:"type"`
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type     string            `mapstructure:"type"`
	BasePath string            `mapstructure:"base_path"`
	S3       S3Config          `mapstructure:"s3"`
	Options  map[string]string `mapstructure:"options"`
}

// S3Config S3存储配置
type S3Config struct {
	Endpoint        string `mapstructure:"endpoint"`
	Region          string `mapstructure:"region"`
	Bucket          string `mapstructure:"bucket"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Type  string      `mapstructure:"type"`
	Redis RedisConfig `mapstructure:"redis"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWTSecret   string `mapstructure:"jwt_secret"`
	JWTExpire   string `mapstructure:"jwt_expire"`
	EnableHTTPS bool   `mapstructure:"enable_https"`
	CertFile    string `mapstructure:"cert_file"`
	KeyFile     string `mapstructure:"key_file"`
}

// PluginsConfig 插件配置
type PluginsConfig struct {
	Enabled []string          `mapstructure:"enabled"`
	Path    string            `mapstructure:"path"`
	Configs map[string]Config `mapstructure:"configs"`
}

// LoadConfig 加载配置文件
func LoadConfig(configFile string) (*Config, error) {
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")

	// 设置环境变量前缀
	viper.SetEnvPrefix("GO_NEXUS")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8081)
	viper.SetDefault("server.mode", "release")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.max_header_bytes", 1048576)

	// 数据库默认配置
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.dsn", "data/go-nexus.db")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.conn_max_lifetime", "5m")

	// 存储默认配置
	viper.SetDefault("storage.type", "filesystem")
	viper.SetDefault("storage.base_path", "/var/lib/go-nexus")

	// 缓存默认配置
	viper.SetDefault("cache.type", "memory")

	// 日志默认配置
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("log.max_size", 100)
	viper.SetDefault("log.max_backups", 3)
	viper.SetDefault("log.max_age", 28)

	// 插件默认配置
	viper.SetDefault("plugins.enabled", []string{"maven", "npm"})
	viper.SetDefault("plugins.path", "/var/lib/go-nexus/plugins")
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	// 验证服务器端口
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	// 验证服务器模式
	if config.Server.Mode != "debug" && config.Server.Mode != "release" && config.Server.Mode != "test" {
		return fmt.Errorf("invalid server mode: %s", config.Server.Mode)
	}

	// 验证数据库类型
	if config.Database.Type != "sqlite" && config.Database.Type != "postgresql" {
		return fmt.Errorf("unsupported database type: %s", config.Database.Type)
	}

	// 验证存储类型
	if config.Storage.Type != "filesystem" && config.Storage.Type != "s3" {
		return fmt.Errorf("unsupported storage type: %s", config.Storage.Type)
	}

	// 验证日志级别
	validLogLevels := []string{"debug", "info", "warn", "error"}
	isValidLogLevel := false
	for _, level := range validLogLevels {
		if config.Log.Level == level {
			isValidLogLevel = true
			break
		}
	}
	if !isValidLogLevel {
		return fmt.Errorf("invalid log level: %s", config.Log.Level)
	}

	return nil
}
