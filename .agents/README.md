# Agents 架构文档索引

## 概述

本目录包含go-nexus项目中智能代理系统的详细技术文档，按模块拆分为独立的说明文件。

## 文档结构

### 📋 核心架构文档

| 文档 | 描述 | 文件 |
|------|------|------|
| **Web框架层** | Gin框架配置、路由设计、中间件实现 | [web-framework.md](./web-framework.md) |
| **配置管理** | Viper配置系统、环境变量、热重载 | [config-management.md](./config-management.md) |
| **日志管理** | slog结构化日志、日志轮转、监控 | [logging-management.md](./logging-management.md) |
| **权限控制** | Casbin RBAC、角色管理、权限策略 | [permission-control.md](./permission-control.md) |
| **存储层** | 本地文件系统、S3兼容存储、接口抽象 | [storage-layer.md](./storage-layer.md) |

### 🎯 快速导航

#### 开发人员
- 如果你需要了解**API开发**，请查看 [Web框架层](./web-framework.md)
- 如果你需要了解**配置系统**，请查看 [配置管理](./config-management.md)
- 如果你需要了解**日志系统**，请查看 [日志管理](./logging-management.md)

#### 运维人员
- 如果你需要了解**权限配置**，请查看 [权限控制](./permission-control.md)
- 如果你需要了解**存储配置**，请查看 [存储层](./storage-layer.md)
- 如果你需要了解**系统配置**，请查看 [配置管理](./config-management.md)

#### 安全人员
- 如果你需要了解**访问控制**，请查看 [权限控制](./permission-control.md)
- 如果你需要了解**审计日志**，请查看 [日志管理](./logging-management.md)
- 如果你需要了解**安全中间件**，请查看 [Web框架层](./web-framework.md)

## 技术栈概览

### 🛠️ 核心技术
- **Web框架**: [Gin](https://gin-gonic.com/) - 高性能HTTP路由框架
- **配置管理**: [Viper](https://github.com/spf13/viper) - 灵活的配置解决方案
- **日志系统**: [slog](https://pkg.go.dev/log/slog) - Go官方结构化日志库
- **权限控制**: [Casbin](https://casbin.org/) - 强大的访问控制框架
- **存储层**: 本地文件系统 + Amazon S3兼容存储

### 🏗️ 架构特点
- **微服务架构**: Agent-based设计，模块化组件
- **前后端分离**: RESTful API + 独立前端应用
- **多存储支持**: 统一接口，支持多种存储后端
- **配置热重载**: 运行时配置更新，无需重启
- **结构化日志**: 便于监控和问题排查
- **细粒度权限**: 基于RBAC的访问控制

## 开发指南

### 🚀 新手入门

1. **环境搭建**
   ```bash
   # 阅读配置管理文档
   cat config-management.md
   
   # 设置开发环境
   cp config/config.example.yaml config/config.yaml
   ```

2. **本地运行**
   ```bash
   # 启动开发服务器
   go run cmd/server/main.go --config=config/config.yaml
   ```

3. **API测试**
   - 参考 [Web框架层文档](./web-framework.md) 中的API示例
   - 使用Postman或curl测试接口

### 🔧 开发流程

1. **添加新功能**
   - 在对应的Agent中实现业务逻辑
   - 在Web框架层添加路由和处理器
   - 更新权限控制策略

2. **配置更新**
   - 在配置结构中添加新字段
   - 更新默认值和验证逻辑
   - 测试配置热重载

3. **日志记录**
   - 使用结构化日志记录关键操作
   - 添加必要的监控指标
   - 确保敏感信息不被记录

### 📚 最佳实践

#### 代码规范
- 遵循Go语言标准规范
- 使用接口抽象外部依赖
- 编写单元测试和集成测试

#### 安全考虑
- 敏感配置使用环境变量
- 所有API都需要认证和授权
- 记录安全相关操作的审计日志

#### 性能优化
- 使用连接池和缓存
- 监控关键性能指标
- 实现优雅的错误处理

## 部署指南

### 📦 部署方式

1. **开发环境**: 直接运行Go二进制文件
2. **测试环境**: Docker容器部署
3. **生产环境**: Kubernetes集群部署

详细部署步骤请参考主文档 [../agents.md](../agents.md) 中的部署指南章节。

### 🔍 监控和运维

1. **健康检查**: `/health` 端点
2. **性能指标**: Prometheus `/metrics` 端点
3. **日志监控**: 结构化日志 + ELK栈
4. **配置管理**: 热重载 + 版本控制

## 故障排除

### 🚨 常见问题

| 问题类型 | 参考文档 | 关键检查点 |
|----------|----------|------------|
| 启动失败 | [配置管理](./config-management.md) | 配置文件格式、必需参数 |
| 认证问题 | [权限控制](./permission-control.md) | JWT密钥、用户角色 |
| 存储错误 | [存储层](./storage-layer.md) | 存储路径、权限、容量 |
| 性能问题 | [日志管理](./logging-management.md) | 日志级别、慢查询日志 |
| API错误 | [Web框架层](./web-framework.md) | 中间件配置、路由匹配 |

### 🔧 调试技巧

1. **日志分析**: 使用结构化日志快速定位问题
2. **配置验证**: 启动时检查配置文件的完整性
3. **权限测试**: 使用权限检查API验证用户权限
4. **性能监控**: 查看Prometheus指标识别瓶颈

## 贡献指南

### 🤝 如何贡献

1. **文档更新**: 发现文档不准确或过时时，请提交PR
2. **新功能**: 添加新功能时，请同步更新相关文档
3. **问题报告**: 发现问题时，请提供详细的复现步骤

### 📝 文档维护

- 保持文档与代码同步
- 添加清晰的示例代码
- 包含必要的配置说明
- 更新版本兼容性信息

---

## 联系方式

- **项目地址**: https://github.com/laolishu/go-nexus
- **问题反馈**: GitHub Issues
- **文档问题**: 请在对应文档中提出Issue

**最后更新**: 2025年9月23日
