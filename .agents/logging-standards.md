# 日志规范

## 日志级别定义

### 级别说明
- **DEBUG**: 调试信息，仅开发环境使用
- **INFO**: 一般信息，业务流程关键节点
- **WARN**: 警告信息，潜在问题但不影响功能
- **ERROR**: 错误信息，功能异常需要关注
- **FATAL**: 致命错误，程序无法继续运行

## 日志格式规范

### 结构化日志格式（JSON）
```json
{
  "timestamp": "2025-09-28T10:30:00Z",
  "level": "INFO",
  "service": "go-nexus",
  "version": "v0.1.0",
  "trace_id": "abc123def456",
  "module": "repository",
  "operation": "upload_artifact",
  "message": "artifact uploaded successfully",
  "fields": {
    "repository": "maven-central-proxy",
    "artifact_path": "/com/example/demo/1.0.0/demo-1.0.0.jar",
    "size_bytes": 1048576,
    "duration_ms": 150
  },
  "error": {
    "code": "STORAGE_ERROR",
    "message": "failed to write file",
    "stack_trace": "..."
  }
}
```

### 文本格式（开发环境）
```
2025-09-28 10:30:00 [INFO] [repository] upload_artifact: artifact uploaded successfully trace_id=abc123def456 repository=maven-central-proxy size=1MB duration=150ms
```

## 日志内容规范

### 业务日志
- **仓库操作**: 创建、删除、配置变更
- **artifact操作**: 上传、下载、删除
- **用户认证**: 登录、权限检查
- **插件管理**: 加载、卸载、调用

### 技术日志
- **HTTP请求**: 记录关键API调用
- **数据库操作**: 慢查询、连接异常
- **缓存操作**: 命中率、失效清理
- **存储操作**: 文件读写、空间使用

### 必须记录的字段
- `timestamp`: 时间戳
- `level`: 日志级别
- `module`: 模块名称
- `operation`: 操作名称
- `trace_id`: 请求追踪ID
- `user_id`: 用户ID（如有）

### 敏感信息处理
- 密码、Token等敏感信息必须脱敏
- 用户数据需要hash处理
- API密钥只记录前4位和后4位

## 日志输出配置

### 开发环境
- 级别: DEBUG
- 格式: 文本格式
- 输出: 控制台

### 生产环境
- 级别: INFO
- 格式: JSON格式
- 输出: 文件 + 日志收集系统

### 日志轮转
- 单文件最大: 100MB
- 保留备份: 7个
- 保留天数: 30天
- 压缩存档: 启用

## 错误日志规范

### 错误码定义
- `AUTH_*`: 认证相关错误 (1000-1999)
- `REPO_*`: 仓库相关错误 (2000-2999)  
- `STORAGE_*`: 存储相关错误 (3000-3999)
- `PLUGIN_*`: 插件相关错误 (4000-4999)
- `SYS_*`: 系统相关错误 (9000-9999)

### 错误日志示例
```go
logger.Error("failed to upload artifact",
    zap.String("error_code", "STORAGE_WRITE_FAILED"),
    zap.String("repository", repoName),
    zap.String("artifact_path", artifactPath),
    zap.Error(err),
    zap.String("trace_id", traceID),
)
```