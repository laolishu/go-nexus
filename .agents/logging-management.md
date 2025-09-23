# 日志管理 (slog)

## 概述

go-nexus使用Go官方的slog包进行结构化日志记录，提供高性能、可扩展的日志功能。

## 日志配置

### 配置结构
```go
type LoggingConfig struct {
    Level  string `mapstructure:"level"`  // debug, info, warn, error
    Format string `mapstructure:"format"` // json, text
    Output string `mapstructure:"output"` // stdout, file
    File   string `mapstructure:"file"`   // 日志文件路径
}
```

### 日志初始化
```go
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
```

## 结构化日志记录器

### Agent专用日志记录器
```go
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

### HTTP请求日志记录器
```go
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

## 日志级别和用法

### Debug级别
```go
// 调试信息，开发阶段使用
logger.Debug("Processing package upload",
    slog.String("package", pkg.Name),
    slog.String("version", pkg.Version),
    slog.String("format", pkg.Format),
)
```

### Info级别
```go
// 一般信息，正常操作记录
logger.Info("Package uploaded successfully",
    slog.String("package", pkg.Name),
    slog.String("version", pkg.Version),
    slog.String("repository", repo.Name),
    slog.Int64("size", pkg.Size),
)
```

### Warn级别
```go
// 警告信息，潜在问题
logger.Warn("Large package detected",
    slog.String("package", pkg.Name),
    slog.Int64("size", pkg.Size),
    slog.Int64("threshold", maxSize),
)
```

### Error级别
```go
// 错误信息，需要关注的问题
logger.Error("Failed to upload package",
    slog.String("package", pkg.Name),
    slog.String("error", err.Error()),
    slog.String("repository", repo.Name),
)
```

## 日志格式

### JSON格式输出
```json
{
  "time": "2024-01-01T12:00:00Z",
  "level": "INFO",
  "msg": "Package uploaded successfully",
  "agent": "package_engine",
  "package": "example-lib",
  "version": "1.0.0",
  "repository": "maven-releases",
  "size": 1048576
}
```

### 文本格式输出
```
2024-01-01T12:00:00Z INFO Package uploaded successfully agent=package_engine package=example-lib version=1.0.0 repository=maven-releases size=1048576
```

## 日志轮转和管理

### 使用logrotate配置
```bash
# /etc/logrotate.d/nexus
/var/log/nexus/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 0644 nexus nexus
    postrotate
        systemctl reload nexus
    endscript
}
```

### 程序内日志轮转
```go
import "gopkg.in/natefinch/lumberjack.v2"

func initLoggerWithRotation(config LoggingConfig) (*slog.Logger, error) {
    var writer io.Writer
    
    if config.Output == "file" {
        writer = &lumberjack.Logger{
            Filename:   config.File,
            MaxSize:    100, // megabytes
            MaxBackups: 3,
            MaxAge:     28, // days
            Compress:   true,
        }
    } else {
        writer = os.Stdout
    }
    
    var handler slog.Handler
    opts := &slog.HandlerOptions{Level: parseLevel(config.Level)}
    
    if config.Format == "json" {
        handler = slog.NewJSONHandler(writer, opts)
    } else {
        handler = slog.NewTextHandler(writer, opts)
    }
    
    return slog.New(handler), nil
}
```

## 上下文日志

### 请求ID追踪
```go
func requestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := uuid.New().String()
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)
        
        // 将请求ID添加到日志上下文
        ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
        c.Request = c.Request.WithContext(ctx)
        
        c.Next()
    }
}
```

### 用户信息追踪
```go
func (a *PackageAgent) HandleUpload(ctx context.Context, pkg *Package) error {
    // 从上下文获取用户信息
    userID := ctx.Value("user_id")
    requestID := ctx.Value("request_id")
    
    logger := a.logger.With(
        slog.String("user_id", userID.(string)),
        slog.String("request_id", requestID.(string)),
        slog.String("operation", "upload"),
    )
    
    logger.Info("Starting package upload",
        slog.String("package", pkg.Name),
        slog.String("version", pkg.Version),
    )
    
    // 处理逻辑...
    
    logger.Info("Package upload completed")
    return nil
}
```

## 性能监控日志

### 操作耗时记录
```go
func LogDuration(logger *slog.Logger, operation string) func() {
    start := time.Now()
    return func() {
        duration := time.Since(start)
        logger.Info("Operation completed",
            slog.String("operation", operation),
            slog.Duration("duration", duration),
        )
        
        // 记录慢操作
        if duration > 5*time.Second {
            logger.Warn("Slow operation detected",
                slog.String("operation", operation),
                slog.Duration("duration", duration),
            )
        }
    }
}

// 使用示例
func (a *PackageAgent) ProcessPackage(pkg *Package) error {
    defer LogDuration(a.logger, "process_package")()
    
    // 处理逻辑...
    return nil
}
```

### 资源使用监控
```go
func (a *StorageAgent) logStorageMetrics() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            stats := a.getStorageStats()
            a.logger.Info("Storage metrics",
                slog.Int64("total_size", stats.TotalSize),
                slog.Int64("used_size", stats.UsedSize),
                slog.Int64("available_size", stats.AvailableSize),
                slog.Float64("usage_percent", stats.UsagePercent),
                slog.Int("file_count", stats.FileCount),
            )
        case <-a.stopCh:
            return
        }
    }
}
```

## 安全和审计日志

### 安全事件记录
```go
func (a *SecurityAgent) logSecurityEvent(event SecurityEvent) {
    a.logger.Warn("Security event detected",
        slog.String("event_type", event.Type),
        slog.String("source_ip", event.SourceIP),
        slog.String("user_id", event.UserID),
        slog.String("resource", event.Resource),
        slog.String("action", event.Action),
        slog.Time("timestamp", event.Timestamp),
        slog.String("severity", event.Severity),
    )
}
```

### 审计日志
```go
func (a *AuditLogger) LogOperation(ctx context.Context, op AuditOperation) {
    a.logger.Info("Audit log",
        slog.String("operation", op.Operation),
        slog.String("resource_type", op.ResourceType),
        slog.String("resource_id", op.ResourceID),
        slog.String("user_id", op.UserID),
        slog.String("source_ip", op.SourceIP),
        slog.Time("timestamp", op.Timestamp),
        slog.String("result", op.Result),
        slog.Any("metadata", op.Metadata),
    )
}
```

## 日志配置示例

### 开发环境配置
```yaml
logging:
  level: "debug"
  format: "text"
  output: "stdout"
```

### 生产环境配置
```yaml
logging:
  level: "info"
  format: "json"
  output: "file"
  file: "/var/log/nexus/nexus.log"
```

### 容器环境配置
```yaml
logging:
  level: "info"
  format: "json"
  output: "stdout"  # 容器环境建议输出到stdout
```

## 日志分析和监控

### ELK集成
```bash
# Filebeat配置示例
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /var/log/nexus/*.log
  fields:
    service: go-nexus
  json.keys_under_root: true

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "nexus-logs-%{+yyyy.MM.dd}"
```

### Prometheus指标
```go
var (
    logMessagesTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "log_messages_total",
            Help: "Total number of log messages",
        },
        []string{"level", "component"},
    )
)

// 在日志处理器中增加指标统计
func (h *MetricsHandler) Handle(ctx context.Context, r slog.Record) error {
    logMessagesTotal.WithLabelValues(r.Level.String(), "nexus").Inc()
    return h.next.Handle(ctx, r)
}
```

## 最佳实践

1. **结构化**: 始终使用结构化日志，便于查询和分析
2. **上下文**: 包含足够的上下文信息便于问题排查
3. **性能**: 避免在高频操作中记录过多调试日志
4. **敏感信息**: 避免记录密码、密钥等敏感信息
5. **一致性**: 保持日志格式和字段名称的一致性
