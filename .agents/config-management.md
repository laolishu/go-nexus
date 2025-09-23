# 配置管理 (Viper)

## 概述

go-nexus使用Viper进行配置管理，支持多种配置源、环境变量绑定和配置热重载。

## 配置结构

### 主配置结构
```go
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Storage  StorageConfig  `mapstructure:"storage"`
    Auth     AuthConfig     `mapstructure:"auth"`
    Agents   AgentsConfig   `mapstructure:"agents"`
    Logging  LoggingConfig  `mapstructure:"logging"`
    Maven    MavenConfig    `mapstructure:"maven"`
    NPM      NPMConfig      `mapstructure:"npm"`
    Plugins  PluginsConfig  `mapstructure:"plugins"`
    Monitor  MonitorConfig  `mapstructure:"monitoring"`
}
```

### 服务器配置
```go
type ServerConfig struct {
    Host         string `mapstructure:"host"`
    Port         int    `mapstructure:"port"`
    TLS          bool   `mapstructure:"tls"`
    CertFile     string `mapstructure:"cert_file"`
    KeyFile      string `mapstructure:"key_file"`
    ReadTimeout  string `mapstructure:"read_timeout"`
    WriteTimeout string `mapstructure:"write_timeout"`
}
```

### 存储配置
```go
type StorageConfig struct {
    Backend string       `mapstructure:"backend"` // "filesystem" or "s3"
    Local   LocalStorage `mapstructure:"local"`
    S3      S3Storage    `mapstructure:"s3"`
}

type LocalStorage struct {
    Path string `mapstructure:"path"`
}

type S3Storage struct {
    Endpoint        string `mapstructure:"endpoint"`
    Region          string `mapstructure:"region"`
    Bucket          string `mapstructure:"bucket"`
    AccessKey       string `mapstructure:"access_key"`
    SecretKey       string `mapstructure:"secret_key"`
    ForcePathStyle  bool   `mapstructure:"force_path_style"`
}
```

### 认证配置
```go
type AuthConfig struct {
    JWT    JWTConfig    `mapstructure:"jwt"`
    Casbin CasbinConfig `mapstructure:"casbin"`
}

type JWTConfig struct {
    Secret string `mapstructure:"secret"`
    Expire string `mapstructure:"expire"`
}

type CasbinConfig struct {
    ModelFile string `mapstructure:"model_file"`
    PolicyDB  string `mapstructure:"policy_db"`
}
```

## 配置初始化

### 基础初始化
```go
func InitConfig() (*viper.Viper, error) {
    v := viper.New()
    
    // 配置文件搜索路径
    v.SetConfigName("config")
    v.SetConfigType("yaml")
    v.AddConfigPath("./config")
    v.AddConfigPath("/etc/nexus")
    v.AddConfigPath(".")
    
    // 环境变量绑定
    v.SetEnvPrefix("NEXUS")
    v.AutomaticEnv()
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    
    // 默认值设置
    setDefaults(v)
    
    if err := v.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }
    
    return v, nil
}
```

### 默认值设置
```go
func setDefaults(v *viper.Viper) {
    // 服务器默认值
    v.SetDefault("server.host", "0.0.0.0")
    v.SetDefault("server.port", 8080)
    v.SetDefault("server.tls", false)
    v.SetDefault("server.read_timeout", "30s")
    v.SetDefault("server.write_timeout", "30s")
    
    // 存储默认值
    v.SetDefault("storage.backend", "filesystem")
    v.SetDefault("storage.local.path", "/var/lib/nexus/storage")
    
    // 日志默认值
    v.SetDefault("logging.level", "info")
    v.SetDefault("logging.format", "json")
    v.SetDefault("logging.output", "stdout")
    
    // Agent默认值
    v.SetDefault("agents.package_engine.enabled", true)
    v.SetDefault("agents.package_engine.workers", 10)
    v.SetDefault("agents.package_engine.max_package_size", "100MB")
    
    v.SetDefault("agents.security_agent.enabled", true)
    v.SetDefault("agents.security_agent.scan_on_upload", true)
    v.SetDefault("agents.security_agent.update_interval", "24h")
    
    v.SetDefault("agents.dependency_agent.enabled", true)
    v.SetDefault("agents.dependency_agent.max_depth", 10)
    v.SetDefault("agents.dependency_agent.cache_ttl", "1h")
    
    v.SetDefault("agents.storage_agent.enabled", true)
    v.SetDefault("agents.storage_agent.cleanup_policy.keep_last_versions", 5)
}
```

## 主配置文件

### config.yaml示例
```yaml
# 服务器配置
server:
  host: "0.0.0.0"
  port: 8080
  tls: false
  cert_file: ""
  key_file: ""
  read_timeout: "30s"
  write_timeout: "30s"
  
# 存储配置
storage:
  backend: "filesystem"  # filesystem 或 s3
  
  # 本地文件存储配置
  local:
    path: "/var/lib/nexus/storage"
    
  # S3兼容存储配置
  s3:
    endpoint: ""  # 留空使用AWS S3，或设置MinIO等兼容服务地址
    region: "us-east-1"
    bucket: "nexus-storage"
    access_key: "${NEXUS_S3_ACCESS_KEY}"
    secret_key: "${NEXUS_S3_SECRET_KEY}"
    force_path_style: false  # MinIO需要设置为true

# 认证授权配置
auth:
  jwt:
    secret: "${NEXUS_JWT_SECRET}"
    expire: "24h"
  casbin:
    model_file: "./config/rbac_model.conf"
    policy_db: "sqlite3://nexus.db"

# Agent配置
agents:
  package_engine:
    enabled: true
    supported_formats: ["maven", "npm"]
    max_package_size: "100MB"
    temp_dir: "/tmp/nexus-agents"
    workers: 10
    
  security_agent:
    enabled: true
    cve_database_url: "https://nvd.nist.gov/feeds/json/cve/1.1/"
    update_interval: "24h"
    scan_on_upload: true
    quarantine_vulnerable: true
    severity_threshold: "medium"
    
  dependency_agent:
    enabled: true
    max_depth: 10
    cache_ttl: "1h"
    resolve_transitive: true
    timeout: "30s"
    
  storage_agent:
    enabled: true
    cleanup_policy:
      keep_last_versions: 5
      cleanup_interval: "7d"
      min_age_for_cleanup: "30d"
    compression:
      enabled: true
      algorithm: "gzip"
      level: 6

# Maven特定配置
maven:
  repositories:
    - name: "maven-central"
      type: "proxy"
      remote_url: "https://repo1.maven.org/maven2/"
      layout: "default"
    - name: "maven-releases"
      type: "hosted"
      write_policy: "allow_once"
      layout: "default"
    - name: "maven-snapshots"
      type: "hosted"
      write_policy: "allow"
      layout: "default"
  metadata_update_policy: "always"
  snapshot_cleanup_days: 30
  checksum_validation: true

# NPM特定配置  
npm:
  registries:
    - name: "npm-proxy"
      type: "proxy"
      remote_url: "https://registry.npmjs.org"
      timeout: "30s"
    - name: "npm-hosted"
      type: "hosted"
      allow_redeploy: false
    - name: "npm-group"
      type: "group"
      members: ["npm-hosted", "npm-proxy"]
  cache_metadata: true
  support_scoped_packages: true
  tarball_cache_size: "1GB"

# 监控配置
monitoring:
  metrics:
    enabled: true
    path: "/metrics"
    port: 9090
  health:
    enabled: true
    path: "/health"
  jaeger:
    enabled: false
    endpoint: "http://localhost:14268/api/traces"
```

## 环境变量

### 环境变量命名规则
```bash
# 基础配置
export NEXUS_CONFIG_FILE="/etc/nexus/config.yaml"
export NEXUS_SERVER_HOST="0.0.0.0"
export NEXUS_SERVER_PORT="8080"

# 认证配置
export NEXUS_JWT_SECRET="your-super-secret-jwt-key-here"
export NEXUS_AUTH_CASBIN_MODEL_FILE="/etc/nexus/rbac_model.conf"

# 存储配置
export NEXUS_STORAGE_BACKEND="s3"
export NEXUS_S3_ACCESS_KEY="your-access-key"
export NEXUS_S3_SECRET_KEY="your-secret-key"
export NEXUS_S3_BUCKET="nexus-storage"
export NEXUS_S3_REGION="us-east-1"

# 日志配置
export NEXUS_LOGGING_LEVEL="info"
export NEXUS_LOGGING_FORMAT="json"
export NEXUS_LOGGING_OUTPUT="stdout"

# Agent配置
export NEXUS_AGENTS_PACKAGE_ENGINE_ENABLED="true"
export NEXUS_AGENTS_SECURITY_AGENT_ENABLED="true"
export NEXUS_AGENTS_SECURITY_AGENT_SCAN_ON_UPLOAD="true"
```

## 配置验证

### 验证函数
```go
func ValidateConfig(config *Config) error {
    // 服务器配置验证
    if config.Server.Port <= 0 || config.Server.Port > 65535 {
        return fmt.Errorf("invalid server port: %d", config.Server.Port)
    }
    
    // 存储配置验证
    if config.Storage.Backend == "s3" {
        if config.Storage.S3.Bucket == "" {
            return fmt.Errorf("S3 bucket name is required")
        }
        if config.Storage.S3.AccessKey == "" || config.Storage.S3.SecretKey == "" {
            return fmt.Errorf("S3 credentials are required")
        }
    } else if config.Storage.Backend == "filesystem" {
        if config.Storage.Local.Path == "" {
            return fmt.Errorf("local storage path is required")
        }
    }
    
    // JWT配置验证
    if config.Auth.JWT.Secret == "" {
        return fmt.Errorf("JWT secret is required")
    }
    
    // 日志配置验证
    validLevels := []string{"debug", "info", "warn", "error"}
    if !contains(validLevels, config.Logging.Level) {
        return fmt.Errorf("invalid log level: %s", config.Logging.Level)
    }
    
    return nil
}
```

## 配置热重载

### 热重载实现
```go
func (s *NexusServer) WatchConfig() {
    s.config.WatchConfig()
    s.config.OnConfigChange(func(e fsnotify.Event) {
        s.logger.Info("Config file changed", slog.String("file", e.Name))
        
        var newConfig Config
        if err := s.config.Unmarshal(&newConfig); err != nil {
            s.logger.Error("Failed to reload config", slog.Any("error", err))
            return
        }
        
        if err := ValidateConfig(&newConfig); err != nil {
            s.logger.Error("Invalid config", slog.Any("error", err))
            return
        }
        
        // 更新运行时配置
        s.updateRuntimeConfig(&newConfig)
        s.logger.Info("Config reloaded successfully")
    })
}

func (s *NexusServer) updateRuntimeConfig(newConfig *Config) {
    // 更新日志级别
    if newConfig.Logging.Level != s.currentConfig.Logging.Level {
        s.updateLogLevel(newConfig.Logging.Level)
    }
    
    // 更新Agent配置
    for name, agent := range s.agents.agents {
        if err := agent.Configure(s.config); err != nil {
            s.logger.Error("Failed to update agent config", 
                slog.String("agent", name), 
                slog.Any("error", err))
        }
    }
    
    s.currentConfig = newConfig
}
```

## 配置加载优先级

1. **命令行参数** (最高优先级)
2. **环境变量**
3. **配置文件**
4. **默认值** (最低优先级)

## 最佳实践

1. **敏感信息**: 使用环境变量存储密钥和密码
2. **默认值**: 为所有配置项设置合理的默认值
3. **验证**: 启动时验证配置的完整性和正确性
4. **文档**: 维护配置项的完整文档和示例
5. **版本控制**: 配置文件模板纳入版本控制，实际配置文件不纳入
