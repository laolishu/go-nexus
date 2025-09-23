# Agents 智能代理

## 概述

本文档定义了go-nexus项目中的智能代理系统架构和规范。

## 项目背景

go-nexus是一个轻量级云原生仓库管理工具，专注于：
- **主要支持**: Maven包(.jar, .war, .pom等) 和 NPM包(.tgz)
- **开发语言**: Go语言，保证高性能和云原生特性
- **扩展性**: 采用插件架构，未来支持Docker镜像、Go Modules等格式
- **目标**: 为企业提供统一的私有仓库管理解决方案

## Agents 系统设计

### 核心理念
Agents是go-nexus的智能处理引擎，负责包的生命周期管理、安全检查和自动化运维。

### 支持的包格式

#### 当前支持
- **Maven Repository**: 
  - 支持Maven Central格式
  - 处理依赖解析和传递性依赖
  - 支持快照版本管理
- **NPM Registry**:
  - 兼容NPM registry API
  - 支持scoped packages
  - 处理语义化版本控制

#    return logger, nil
}

// 结构化日志记录器
type AgentLogger struct {
    *slog.Logger
    agentName string
}

func NewAgentLogger(base *slog.Logger, agentName string) *AgentLogger {
    return &AgentLogger{
        Logger:    base.With(slog.String("agent", agentName)),
        agentName: agentName,
    }
}

func (l *AgentLogger) WithContext(ctx context.Context) *slog.Logger {
    if reqID := ctx.Value("request_id"); reqID != nil {
        return l.Logger.With(slog.String("request_id", reqID.(string)))
    }
    return l.Logger
}
```

### 前后端分离架构
```go
// API响应标准格式
type APIResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
    Total  int `json:"total,omitempty"`
    Page   int `json:"page,omitempty"`
    Size   int `json:"size,omitempty"`
    Pages  int `json:"pages,omitempty"`
}

// 统一响应处理
func SuccessResponse(c *gin.Context, data interface{}) {
    c.JSON(200, APIResponse{
        Code:    0,
        Message: "success",
        Data:    data,
    })
}

func ErrorResponse(c *gin.Context, code int, message string) {
    c.JSON(code, APIResponse{
        Code:    code,
        Message: message,
    })
}

// 前端路由配置示例
type FrontendRoutes struct {
    // 管理界面路由
    Admin struct {
        Dashboard   string `json:"dashboard"`   // /admin/dashboard
        Repositories string `json:"repositories"` // /admin/repositories
        Users       string `json:"users"`       // /admin/users
        Agents      string `json:"agents"`      // /admin/agents
        Settings    string `json:"settings"`    // /admin/settings
    } `json:"admin"`
    
    // 仓库浏览路由
    Browse struct {
        Maven string `json:"maven"` // /browse/maven/*
        NPM   string `json:"npm"`   // /browse/npm/*
    } `json:"browse"`
}
```

### 插件系统设计
```go
type PackageHandler interface {
    // 包格式标识
    Format() string
    
    // 包上传处理
    Upload(ctx context.Context, pkg *Package) error
    
    // 包下载处理  
    Download(ctx context.Context, req *DownloadRequest) (*Package, error)
    
    // 包元数据解析
    ParseMetadata(data []byte) (*Metadata, error)
    
    // 依赖解析
    ResolveDependencies(pkg *Package) ([]*Dependency, error)
}
```

### 核心功能模块

### 核心功能模块

#### 1. 包处理引擎 (Package Engine)
```go
type PackageAgent struct {
    handlers map[string]PackageHandler
    storage  StorageBackend
    metadata MetadataService
}

// Maven包处理
func (a *PackageAgent) HandleMavenUpload(artifact *MavenArtifact) error {
    // 1. 验证包格式和完整性
    // 2. 解析POM文件，提取依赖信息
    // 3. 生成Maven元数据(maven-metadata.xml)
    // 4. 存储到仓库
}

// NPM包处理
func (a *PackageAgent) HandleNpmUpload(pkg *NpmPackage) error {
    // 1. 验证package.json格式
    // 2. 检查版本冲突
    // 3. 处理scoped packages路径
    // 4. 更新包索引
}
```

#### 2. 安全扫描代理 (Security Agent)
- **漏洞检测**: 集成CVE数据库，检查已知漏洞
- **许可证检查**: 分析包的许可证兼容性
- **恶意代码检测**: 静态分析可疑代码模式

#### 3. 依赖管理代理 (Dependency Agent)
- **传递依赖解析**: 自动计算完整依赖树
- **版本冲突检测**: 识别不兼容的版本组合
- **依赖更新通知**: 监控上游包更新

#### 4. 存储管理代理 (Storage Agent)
- **智能清理**: 自动清理过期和未使用的包版本
- **压缩优化**: 对历史版本进行压缩存储
- **备份同步**: 跨区域数据同步

## 技术架构

### 核心技术栈
- **Web框架**: Gin - 高性能HTTP路由框架
- **日志系统**: slog - Go官方结构化日志库
- **配置管理**: Viper - 灵活的配置解决方案
- **权限控制**: Casbin - 强大的访问控制框架
- **存储层**: 本地文件系统 + Amazon S3兼容存储
- **前后端分离**: RESTful API + 独立前端应用

### 整体架构设计
```go
// 主应用结构
type NexusServer struct {
    router      *gin.Engine
    logger      *slog.Logger
    config      *viper.Viper
    enforcer    *casbin.Enforcer
    storage     storage.Backend
    agents      *AgentManager
}

// Agent生命周期管理
type AgentManager struct {
    agents   map[string]Agent
    config   *viper.Viper
    logger   *slog.Logger
    registry *PluginRegistry
}

type Agent interface {
    Name() string
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() error
    Configure(config *viper.Viper) error
}
```

}
```

### Web框架层 (Gin)
```go
// 路由配置
func (s *NexusServer) setupRoutes() {
    // 中间件配置
    s.router.Use(
        middleware.Logger(s.logger),
        middleware.Recovery(),
        middleware.CORS(),
        middleware.RateLimit(),
    )
    
    // API路由组
    api := s.router.Group("/api/v1")
    {
        // 认证路由
        auth := api.Group("/auth")
        auth.POST("/login", s.handlers.Login)
        auth.POST("/logout", s.handlers.Logout)
        auth.GET("/profile", middleware.Auth(s.enforcer), s.handlers.GetProfile)
        
        // Maven仓库路由
        maven := api.Group("/repository/maven")
        maven.Use(middleware.Auth(s.enforcer))
        maven.PUT("/*path", middleware.Permission("maven:write"), s.handlers.MavenUpload)
        maven.GET("/*path", middleware.Permission("maven:read"), s.handlers.MavenDownload)
        
        // NPM仓库路由  
        npm := api.Group("/repository/npm")
        npm.Use(middleware.Auth(s.enforcer))
        npm.PUT("/*path", middleware.Permission("npm:write"), s.handlers.NpmPublish)
        npm.GET("/*path", middleware.Permission("npm:read"), s.handlers.NpmInstall)
        
        // Agent管理路由
        agents := api.Group("/agents")
        agents.Use(middleware.Auth(s.enforcer), middleware.Permission("admin:manage"))
        agents.GET("/status", s.handlers.GetAgentsStatus)
        agents.POST("/:name/restart", s.handlers.RestartAgent)
    }
    
    // 静态文件服务(前端资源)
    s.router.Static("/static", "./web/dist")
    s.router.NoRoute(func(c *gin.Context) {
        c.File("./web/dist/index.html")
    })
}

// 中间件实现
func Logger(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        
        logger.Info("HTTP Request",
            slog.String("method", c.Request.Method),
            slog.String("path", c.Request.URL.Path),
            slog.Int("status", c.Writer.Status()),
            slog.Duration("latency", time.Since(start)),
            slog.String("ip", c.ClientIP()),
        )
    }
}
```

### 配置管理 (Viper)
```go
// 配置结构定义
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Storage  StorageConfig  `mapstructure:"storage"`
    Auth     AuthConfig     `mapstructure:"auth"`
    Agents   AgentsConfig   `mapstructure:"agents"`
    Logging  LoggingConfig  `mapstructure:"logging"`
}

type ServerConfig struct {
    Host         string `mapstructure:"host"`
    Port         int    `mapstructure:"port"`
    TLS          bool   `mapstructure:"tls"`
    CertFile     string `mapstructure:"cert_file"`
    KeyFile      string `mapstructure:"key_file"`
    ReadTimeout  string `mapstructure:"read_timeout"`
    WriteTimeout string `mapstructure:"write_timeout"`
}

type StorageConfig struct {
    Backend string          `mapstructure:"backend"` // "filesystem" or "s3"
    Local   LocalStorage    `mapstructure:"local"`
    S3      S3Storage       `mapstructure:"s3"`
}

// 配置初始化
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
    v.SetDefault("server.host", "0.0.0.0")
    v.SetDefault("server.port", 8080)
    v.SetDefault("storage.backend", "filesystem")
    v.SetDefault("logging.level", "info")
    
    if err := v.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }
    
    return v, nil
}
```

### 权限控制 (Casbin)
```go
// RBAC模型定义 (model.conf)
const CasbinModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

// 权限初始化
func InitCasbin(configPath string) (*casbin.Enforcer, error) {
    // 创建模型
    m, err := model.NewModelFromString(CasbinModel)
    if err != nil {
        return nil, err
    }
    
    // 使用数据库适配器存储策略
    adapter, err := gormadapter.NewAdapter("sqlite3", "nexus.db")
    if err != nil {
        return nil, err
    }
    
    enforcer, err := casbin.NewEnforcer(m, adapter)
    if err != nil {
        return nil, err
    }
    
    // 加载默认策略
    if err := loadDefaultPolicies(enforcer); err != nil {
        return nil, err
    }
    
    return enforcer, nil
}

// 默认策略配置
func loadDefaultPolicies(e *casbin.Enforcer) error {
    // 角色定义
    roles := [][]string{
        {"admin", "repository:maven:*", "read"},
        {"admin", "repository:maven:*", "write"},
        {"admin", "repository:npm:*", "read"},
        {"admin", "repository:npm:*", "write"},
        {"admin", "agents:*", "manage"},
        
        {"developer", "repository:maven:*", "read"},
        {"developer", "repository:maven:releases", "write"},
        {"developer", "repository:npm:*", "read"},
        {"developer", "repository:npm:hosted", "write"},
        
        {"readonly", "repository:maven:*", "read"},
        {"readonly", "repository:npm:*", "read"},
    }
    
    for _, policy := range roles {
        if _, err := e.AddPolicy(policy); err != nil {
            return err
        }
    }
    
    return nil
}

// 权限检查中间件
func Permission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        user, exists := c.Get("user")
        if !exists {
            c.JSON(401, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }
        
        enforcer := c.MustGet("enforcer").(*casbin.Enforcer)
        allowed, err := enforcer.Enforce(user, c.Request.URL.Path, permission)
        if err != nil || !allowed {
            c.JSON(403, gin.H{"error": "forbidden"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### 存储层设计
```go
// 存储接口定义
type Backend interface {
    Store(ctx context.Context, path string, data io.Reader) error
    Retrieve(ctx context.Context, path string) (io.ReadCloser, error)
    Delete(ctx context.Context, path string) error
    Exists(ctx context.Context, path string) (bool, error)
    List(ctx context.Context, prefix string) ([]string, error)
    Stat(ctx context.Context, path string) (*ObjectInfo, error)
}

type ObjectInfo struct {
    Path         string
    Size         int64
    LastModified time.Time
    ETag         string
    ContentType  string
}

// 本地文件系统存储
type LocalStorage struct {
    basePath string
    logger   *slog.Logger
}

func NewLocalStorage(basePath string, logger *slog.Logger) *LocalStorage {
    return &LocalStorage{
        basePath: basePath,
        logger:   logger,
    }
}

func (s *LocalStorage) Store(ctx context.Context, path string, data io.Reader) error {
    fullPath := filepath.Join(s.basePath, path)
    
    // 创建目录
    if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
        return fmt.Errorf("failed to create directory: %w", err)
    }
    
    // 写入文件
    file, err := os.Create(fullPath)
    if err != nil {
        return fmt.Errorf("failed to create file: %w", err)
    }
    defer file.Close()
    
    if _, err := io.Copy(file, data); err != nil {
        return fmt.Errorf("failed to write file: %w", err)
    }
    
    s.logger.Info("File stored", slog.String("path", path))
    return nil
}

// S3兼容存储
type S3Storage struct {
    client *s3.Client
    bucket string
    logger *slog.Logger
}

func NewS3Storage(config S3Config, logger *slog.Logger) (*S3Storage, error) {
    cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
        awsconfig.WithRegion(config.Region),
        awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
            config.AccessKey, config.SecretKey, "",
        )),
    )
    if err != nil {
        return nil, err
    }
    
    // 自定义端点(用于MinIO等S3兼容存储)
    if config.Endpoint != "" {
        cfg.BaseEndpoint = aws.String(config.Endpoint)
    }
    
    return &S3Storage{
        client: s3.NewFromConfig(cfg),
        bucket: config.Bucket,
        logger: logger,
    }, nil
}

func (s *S3Storage) Store(ctx context.Context, path string, data io.Reader) error {
    _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(path),
        Body:   data,
    })
    
    if err != nil {
        return fmt.Errorf("failed to upload to S3: %w", err)
    }
    
    s.logger.Info("Object stored to S3", 
        slog.String("bucket", s.bucket),
        slog.String("key", path),
    )
    return nil
}

// 存储工厂
func NewStorageBackend(config StorageConfig, logger *slog.Logger) (Backend, error) {
    switch config.Backend {
    case "filesystem":
        return NewLocalStorage(config.Local.Path, logger), nil
    case "s3":
        return NewS3Storage(config.S3, logger)
    default:
        return nil, fmt.Errorf("unsupported storage backend: %s", config.Backend)
    }
}
```

### 日志系统 (slog)
```go
// 日志配置
type LoggingConfig struct {
    Level  string `mapstructure:"level"`  // debug, info, warn, error
    Format string `mapstructure:"format"` // json, text
    Output string `mapstructure:"output"` // stdout, file
    File   string `mapstructure:"file"`   // 日志文件路径
}

// 日志初始化
func InitLogger(config LoggingConfig) (*slog.Logger, error) {
    var level slog.Level
    switch strings.ToLower(config.Level) {
    case "debug":
        level = slog.LevelDebug
    case "info":
        level = slog.LevelInfo
    case "warn":
        level = slog.LevelWarn
    case "error":
        level = slog.LevelError
    default:
        level = slog.LevelInfo
    }
    
    var writer io.Writer = os.Stdout
    if config.Output == "file" && config.File != "" {
        file, err := os.OpenFile(config.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if err != nil {
            return nil, fmt.Errorf("failed to open log file: %w", err)
        }
        writer = file
    }
    
    var handler slog.Handler
    opts := &slog.HandlerOptions{Level: level}
    
    if config.Format == "json" {
        handler = slog.NewJSONHandler(writer, opts)
    } else {
        handler = slog.NewTextHandler(writer, opts)
    }
    
    logger := slog.New(handler)
    slog.SetDefault(logger)
    
    return logger, nil
}
```go
// 插件接口
type Plugin interface {
    Name() string
    Version() string
    Initialize(ctx context.Context, config Config) error
    Shutdown(ctx context.Context) error
}

// Docker插件示例（未来扩展）
type DockerPlugin struct {
    registry string
    auth     AuthConfig
}

func (p *DockerPlugin) HandleDockerPush(image *DockerImage) error {
    // Docker镜像处理逻辑
}
```

## 使用场景

### 场景1: Maven项目发布
```bash
# 开发者使用Maven发布
mvn deploy -DaltDeploymentRepository=nexus::default::http://nexus.company.com/repository/maven-releases/

# Agent处理流程:
# 1. 接收到jar/war/pom文件
# 2. 验证文件完整性(checksum)
# 3. 解析POM依赖关系
# 4. 安全扫描检查
# 5. 生成Maven元数据
# 6. 存储到仓库
```

### 场景2: NPM包发布和安装
```bash
# 发布NPM包
npm publish --registry http://nexus.company.com/repository/npm-hosted/

# 安装NPM包
npm install @company/utils --registry http://nexus.company.com/repository/npm-group/

# Agent处理:
# 1. 验证package.json格式
# 2. 检查版本语义化规则
# 3. 处理scoped package路径
# 4. 更新npm索引
# 5. 代理上游包下载
```

### 场景3: 安全漏洞响应
```yaml
# 当检测到安全漏洞时
security_agent:
  triggers:
    - cve_database_update
    - new_package_upload
  actions:
    - scan_all_packages
    - generate_vulnerability_report
    - notify_administrators
```

## 配置管理

> **详细文档**: 完整的配置管理文档请参考 [agents/config-management.md](./agents/config-management.md)

### 快速配置指南

#### 基础配置文件 (config.yaml)
```yaml
# 服务器配置
server:
  host: "0.0.0.0"
  port: 8080
  
# 存储配置  
storage:
  backend: "filesystem"  # 或 "s3"
  local:
    path: "/var/lib/nexus/storage"
    
# 认证配置
auth:
  jwt:
    secret: "${NEXUS_JWT_SECRET}"
    expire: "24h"
    
# 日志配置
logging:
  level: "info"
  format: "json"
  output: "stdout"
```

#### 环境变量配置
```bash
export NEXUS_JWT_SECRET="your-jwt-secret"
export NEXUS_S3_ACCESS_KEY="your-access-key"
export NEXUS_S3_SECRET_KEY="your-secret-key"
```

### 相关文档
- **详细配置**: [agents/config-management.md](./agents/config-management.md)
- **Web框架**: [agents/web-framework.md](./agents/web-framework.md)
- **权限控制**: [agents/permission-control.md](./agents/permission-control.md)
- **日志管理**: [agents/logging-management.md](./agents/logging-management.md)
- **存储层**: [agents/storage-layer.md](./agents/storage-layer.md)

## 部署指南

### 1. 本地开发环境
```bash
# 克隆代码
git clone https://github.com/laolishu/go-nexus.git
cd go-nexus

# 安装Go依赖
go mod tidy

# 创建配置文件
cp config/config.example.yaml config/config.yaml

# 初始化数据库
go run cmd/migrate/main.go

# 启动开发服务器
go run cmd/server/main.go --config=config/config.yaml

# 构建前端(如果有独立前端项目)
cd web
npm install
npm run build
```

### 2. 二进制部署
```bash
# 构建二进制文件
make build
# 或者
go build -o nexus-server ./cmd/server

# 创建系统用户
sudo useradd --system --no-create-home nexus

# 创建目录结构
sudo mkdir -p /etc/nexus /var/lib/nexus /var/log/nexus
sudo chown nexus:nexus /var/lib/nexus /var/log/nexus

# 复制配置文件
sudo cp config/config.yaml /etc/nexus/
sudo cp config/rbac_model.conf /etc/nexus/

# 创建systemd服务
sudo tee /etc/systemd/system/nexus.service > /dev/null <<EOF
[Unit]
Description=Go Nexus Repository Manager
After=network.target

[Service]
Type=simple
User=nexus
Group=nexus
ExecStart=/usr/local/bin/nexus-server --config=/etc/nexus/config.yaml
Restart=always
RestartSec=5
Environment=NEXUS_DATA_DIR=/var/lib/nexus
Environment=NEXUS_LOG_LEVEL=INFO

[Install]
WantedBy=multi-user.target
EOF

# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable nexus
sudo systemctl start nexus
```

### 3. Docker部署
```dockerfile
# 多阶段构建Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o nexus-server ./cmd/server

# 运行时镜像
FROM alpine:latest

# 安装必要的包
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1000 nexus && \
    adduser -D -s /bin/sh -u 1000 -G nexus nexus

WORKDIR /app

# 复制二进制文件和配置
COPY --from=builder /app/nexus-server .
COPY --from=builder /app/config/config.yaml ./config/
COPY --from=builder /app/config/rbac_model.conf ./config/

# 创建数据目录
RUN mkdir -p /data /logs && \
    chown -R nexus:nexus /app /data /logs

USER nexus

EXPOSE 8080 9090

VOLUME ["/data", "/logs"]

CMD ["./nexus-server", "--config=config/config.yaml"]
```

```bash
# 构建镜像
docker build -t go-nexus:latest .

# 使用Docker Compose部署
cat > docker-compose.yml <<EOF
version: '3.8'

services:
  nexus:
    image: go-nexus:latest
    ports:
      - "8080:8080"
      - "9090:9090"
    volumes:
      - nexus-data:/data
      - nexus-logs:/logs
      - ./config:/app/config
    environment:
      - NEXUS_JWT_SECRET=your-jwt-secret-here
      - NEXUS_S3_ACCESS_KEY=your-access-key
      - NEXUS_S3_SECRET_KEY=your-secret-key
    restart: unless-stopped
    
  # 可选：MinIO作为S3兼容存储
  minio:
    image: minio/minio:latest
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - minio-data:/data
    environment:
      - MINIO_ROOT_USER=admin
      - MINIO_ROOT_PASSWORD=admin123
    command: server /data --console-address ":9001"
    restart: unless-stopped

volumes:
  nexus-data:
  nexus-logs:
  minio-data:
EOF

# 启动服务
docker-compose up -d
```

### 4. Kubernetes部署
```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: nexus

---
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: nexus-config
  namespace: nexus
data:
  config.yaml: |
    server:
      host: "0.0.0.0"
      port: 8080
    storage:
      backend: "s3"
      s3:
        endpoint: "http://minio:9000"
        region: "us-east-1"
        bucket: "nexus"
        access_key: "${NEXUS_S3_ACCESS_KEY}"
        secret_key: "${NEXUS_S3_SECRET_KEY}"
        force_path_style: true
    logging:
      level: "info"
      format: "json"
    # ... 其他配置
  rbac_model.conf: |
    [request_definition]
    r = sub, obj, act
    # ... RBAC模型配置

---
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: nexus-secrets
  namespace: nexus
type: Opaque
data:
  jwt-secret: eW91ci1qd3Qtc2VjcmV0LWhlcmU=  # base64编码
  s3-access-key: YWRtaW4=                   # base64编码
  s3-secret-key: YWRtaW4xMjM=               # base64编码

---
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nexus
  namespace: nexus
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nexus
  template:
    metadata:
      labels:
        app: nexus
    spec:
      containers:
      - name: nexus
        image: go-nexus:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: NEXUS_JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: nexus-secrets
              key: jwt-secret
        - name: NEXUS_S3_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: nexus-secrets
              key: s3-access-key
        - name: NEXUS_S3_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: nexus-secrets
              key: s3-secret-key
        volumeMounts:
        - name: config
          mountPath: /app/config
        - name: data
          mountPath: /data
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: nexus-config
      - name: data
        persistentVolumeClaim:
          claimName: nexus-data

---
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: nexus-service
  namespace: nexus
spec:
  selector:
    app: nexus
  ports:
  - name: http
    port: 8080
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
  type: LoadBalancer

---
# pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: nexus-data
  namespace: nexus
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 50Gi

---
# ingress.yaml (可选)
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nexus-ingress
  namespace: nexus
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-body-size: "100m"
spec:
  tls:
  - hosts:
    - nexus.company.com
    secretName: nexus-tls
  rules:
  - host: nexus.company.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: nexus-service
            port:
              number: 8080
```

```bash
# 部署到Kubernetes
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml
kubectl apply -f pvc.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f ingress.yaml

# 检查部署状态
kubectl get pods -n nexus
kubectl logs -f deployment/nexus -n nexus
```

## API规范

### 认证API
```bash
# 用户登录
POST /api/v1/auth/login
Content-Type: application/json
{
  "username": "admin",
  "password": "admin123"
}

# 响应
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "admin",
      "roles": ["admin"]
    }
  }
}

# 获取用户信息
GET /api/v1/auth/profile
Authorization: Bearer <token>
```

### Maven Repository API
```bash
# 上传Maven artifact
PUT /api/v1/repository/maven/releases/com/company/artifact/1.0.0/artifact-1.0.0.jar
PUT /api/v1/repository/maven/releases/com/company/artifact/1.0.0/artifact-1.0.0.pom
Authorization: Bearer <token>
Content-Type: application/octet-stream

# 下载Maven artifact
GET /api/v1/repository/maven/public/com/company/artifact/1.0.0/artifact-1.0.0.jar
Authorization: Bearer <token>

# 搜索Maven artifacts
GET /api/v1/search/maven?q=artifact&g=com.company&v=1.0.0
Authorization: Bearer <token>

# 响应
{
  "code": 0,
  "message": "success",
  "data": {
    "artifacts": [
      {
        "groupId": "com.company",
        "artifactId": "artifact",
        "version": "1.0.0",
        "packaging": "jar",
        "lastModified": "2024-01-01T00:00:00Z"
      }
    ]
  },
  "meta": {
    "total": 1,
    "page": 1,
    "size": 20
  }
}

# 获取Maven元数据
GET /api/v1/repository/maven/public/com/company/artifact/maven-metadata.xml
Authorization: Bearer <token>
```

### NPM Registry API
```bash
# 发布NPM包
PUT /api/v1/repository/npm/hosted/@company%2futils
Authorization: Bearer <token>
Content-Type: application/json
{
  "name": "@company/utils",
  "version": "1.0.0",
  "description": "Utility functions",
  "_attachments": {
    "utils-1.0.0.tgz": {
      "content_type": "application/octet-stream",
      "data": "base64-encoded-tarball"
    }
  }
}

# 安装NPM包
GET /api/v1/repository/npm/group/@company/utils/1.0.0
Authorization: Bearer <token>

# NPM包搜索
GET /api/v1/repository/npm/group/-/v1/search?text=utils&size=20
Authorization: Bearer <token>

# 响应
{
  "objects": [
    {
      "package": {
        "name": "@company/utils",
        "version": "1.0.0",
        "description": "Utility functions",
        "keywords": ["utils", "helpers"],
        "date": "2024-01-01T00:00:00Z"
      },
      "score": {
        "final": 0.95,
        "detail": {
          "quality": 0.9,
          "popularity": 0.8,
          "maintenance": 1.0
        }
      }
    }
  ],
  "total": 1,
  "time": "2024-01-01T00:00:00Z"
}
```

### Agent管理API
```bash
# 获取所有Agent状态
GET /api/v1/agents/status
Authorization: Bearer <token>

# 响应
{
  "code": 0,
  "message": "success",
  "data": {
    "agents": [
      {
        "name": "package_engine",
        "status": "running",
        "health": "healthy",
        "last_updated": "2024-01-01T00:00:00Z",
        "metrics": {
          "processed_packages": 1234,
          "failed_packages": 5,
          "average_processing_time": "2.5s"
        }
      }
    ]
  }
}

# 重启特定Agent
POST /api/v1/agents/package_engine/restart
Authorization: Bearer <token>

# 获取Agent配置
GET /api/v1/agents/package_engine/config
Authorization: Bearer <token>

# 更新Agent配置
PUT /api/v1/agents/package_engine/config
Authorization: Bearer <token>
Content-Type: application/json
{
  "max_package_size": "200MB",
  "workers": 20
}
```

### 系统管理API
```bash
# 系统健康检查
GET /health

# 响应
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "1.0.0",
  "components": {
    "database": "healthy",
    "storage": "healthy",
    "agents": "healthy"
  }
}

# Prometheus指标
GET /metrics

# 系统信息
GET /api/v1/system/info
Authorization: Bearer <token>

# 响应
{
  "code": 0,
  "data": {
    "version": "1.0.0",
    "build_time": "2024-01-01T00:00:00Z",
    "go_version": "go1.21.0",
    "uptime": "72h30m15s",
    "storage_usage": {
      "total_space": "100GB",
      "used_space": "45GB",
      "available_space": "55GB"
    }
  }
}
```

## 最佳实践

### 1. 性能优化
```go
// 连接池配置
type ServerConfig struct {
    ReadTimeout    time.Duration `mapstructure:"read_timeout"`
    WriteTimeout   time.Duration `mapstructure:"write_timeout"`
    IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
    MaxHeaderBytes int           `mapstructure:"max_header_bytes"`
}

// Gin性能优化
func optimizeGin(router *gin.Engine) {
    // 禁用控制台颜色输出
    gin.DisableConsoleColor()
    
    // 设置最大多部分内存
    router.MaxMultipartMemory = 8 << 20 // 8 MiB
    
    // 启用gzip压缩
    router.Use(gzip.Gzip(gzip.DefaultCompression))
    
    // 缓存控制
    router.Use(func(c *gin.Context) {
        if strings.HasPrefix(c.Request.URL.Path, "/static/") {
            c.Header("Cache-Control", "public, max-age=31536000")
        }
        c.Next()
    })
}

// 并发限制
func rateLimitMiddleware() gin.HandlerFunc {
    limiter := rate.NewLimiter(rate.Limit(100), 200) // 100 req/s, burst 200
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(429, gin.H{"error": "rate limit exceeded"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### 2. 安全建议
```go
// CORS配置
func corsConfig() cors.Config {
    return cors.Config{
        AllowOrigins:     []string{"https://nexus.company.com"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }
}

// 安全头设置
func securityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Next()
    }
}

// JWT中间件
func jwtMiddleware(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(401, gin.H{"error": "missing authorization header"})
            c.Abort()
            return
        }
        
        tokenString = strings.TrimPrefix(tokenString, "Bearer ")
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(secret), nil
        })
        
        if err != nil || !token.Valid {
            c.JSON(401, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }
        
        claims := token.Claims.(jwt.MapClaims)
        c.Set("user", claims["sub"])
        c.Next()
    }
}
```

### 3. 监控与观测
```go
// Prometheus指标
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration",
        },
        []string{"method", "endpoint"},
    )
    
    packageUploads = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "package_uploads_total",
            Help: "Total number of package uploads",
        },
        []string{"format", "repository"},
    )
    
    storageUsage = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "storage_usage_bytes",
            Help: "Storage usage in bytes",
        },
        []string{"backend", "repository"},
    )
)

// 指标中间件
func prometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start).Seconds()
        httpRequestsTotal.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            strconv.Itoa(c.Writer.Status()),
        ).Inc()
        
        httpRequestDuration.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
        ).Observe(duration)
    }
}

// 结构化日志记录
func requestLogger(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        
        logger.Info("HTTP Request",
            slog.String("method", c.Request.Method),
            slog.String("path", c.Request.URL.Path),
            slog.String("query", c.Request.URL.RawQuery),
            slog.Int("status", c.Writer.Status()),
            slog.Duration("latency", time.Since(start)),
            slog.String("ip", c.ClientIP()),
            slog.String("user_agent", c.Request.UserAgent()),
            slog.Int("size", c.Writer.Size()),
        )
    }
}
```

### 4. 配置管理最佳实践
```go
// 配置验证
func validateConfig(config *Config) error {
    if config.Server.Port <= 0 || config.Server.Port > 65535 {
        return fmt.Errorf("invalid server port: %d", config.Server.Port)
    }
    
    if config.Storage.Backend == "s3" {
        if config.Storage.S3.Bucket == "" {
            return fmt.Errorf("S3 bucket name is required")
        }
        if config.Storage.S3.AccessKey == "" || config.Storage.S3.SecretKey == "" {
            return fmt.Errorf("S3 credentials are required")
        }
    }
    
    return nil
}

// 配置热重载
func (s *NexusServer) watchConfig() {
    s.config.WatchConfig()
    s.config.OnConfigChange(func(e fsnotify.Event) {
        s.logger.Info("Config file changed", slog.String("file", e.Name))
        
        var newConfig Config
        if err := s.config.Unmarshal(&newConfig); err != nil {
            s.logger.Error("Failed to reload config", slog.Any("error", err))
            return
        }
        
        if err := validateConfig(&newConfig); err != nil {
            s.logger.Error("Invalid config", slog.Any("error", err))
            return
        }
        
        // 更新运行时配置
        s.updateRuntimeConfig(&newConfig)
    })
}
```

### 4. 故障排除

#### 常见问题
1. **Maven上传失败**
   - 检查POM文件格式
   - 验证checksum是否正确
   - 确认仓库路径权限

2. **NPM安装缓慢**
   - 检查上游registry连接
   - 调整代理超时设置
   - 清理本地缓存

3. **Agent启动失败**
   - 检查配置文件语法
   - 验证存储路径权限
   - 查看系统资源限制

#### 日志分析
```bash
# 查看Agent日志
tail -f /var/log/nexus/agents.log | grep ERROR

# 监控包上传
grep "package_upload" /var/log/nexus/agents.log | tail -100

# 分析性能问题
grep "slow_request" /var/log/nexus/agents.log
```

## 开发计划

### Phase 1: 核心功能 (MVP)
- [x] 基础Agent框架设计
- [ ] Maven包处理Agent
- [ ] NPM包处理Agent  
- [ ] 基础存储管理
- [ ] 简单的Web UI

### Phase 2: 增强功能
- [ ] 安全扫描Agent
- [ ] 依赖管理Agent
- [ ] 性能监控
- [ ] 集群部署支持
- [ ] REST API完善

### Phase 3: 插件生态
- [ ] 插件SDK开发
- [ ] Docker插件
- [ ] Go Modules插件
- [ ] Python PyPI插件
- [ ] 社区插件支持

## 贡献指南

### 开发环境搭建
```bash
# 克隆代码
git clone https://github.com/laolishu/go-nexus.git
cd go-nexus

# 安装依赖
go mod tidy

# 运行测试
go test ./...

# 本地运行
go run cmd/server/main.go --config=config/agents.yaml
```

### Agent开发示例
```go
// 创建新的Agent
type CustomAgent struct {
    config Config
    logger *logrus.Logger
}

func (a *CustomAgent) Name() string {
    return "custom-agent"
}

func (a *CustomAgent) Start(ctx context.Context) error {
    a.logger.Info("Starting custom agent")
    // 初始化逻辑
    return nil
}

func (a *CustomAgent) Stop(ctx context.Context) error {
    a.logger.Info("Stopping custom agent")
    // 清理逻辑
    return nil
}

func (a *CustomAgent) Health() error {
    // 健康检查逻辑
    return nil
}

// 注册Agent
func init() {
    agent.Registry.Register("custom", func(config Config) Agent {
        return &CustomAgent{
            config: config,
            logger: logrus.WithField("agent", "custom"),
        }
    })
}
```

## 社区与支持

- **GitHub**: https://github.com/laolishu/go-nexus
- **文档**: https://go-nexus.readthedocs.io
- **问题反馈**: GitHub Issues
- **讨论区**: GitHub Discussions

---

**注意**: 这是一个活跃开发中的项目，API和配置可能会发生变化。请关注项目的CHANGELOG和版本发布说明。
```
