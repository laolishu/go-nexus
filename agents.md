# go-nexus：轻量云原生仓库管理工具

## 项目概述

go-nexus是一款基于Golang开发的轻量云原生仓库管理工具，专为中小团队及云原生环境设计。项目采用微内核+插件架构，使用Wire依赖注入框架，遵循云原生设计理念。

### 核心特性
- **轻量部署**: 单二进制文件运行，无需JVM，启动内存占用仅为传统工具的1/5
- **云原生设计**: 原生支持Kubernetes、Docker及对象存储（S3/MinIO）
- **多格式兼容**: 目前支持Maven、npm；Docker、Helm等通过插件扩展
- **高性能**: 基于Goroutine模型，单实例支持10k+并发请求
- **插件化架构**: 核心体积<20MB，功能按需扩展

### 技术栈
- **语言**: Go 1.21+
- **依赖注入**: Google Wire
- **Web框架**: Gin
- **数据库**: SQLite（单节点）/ PostgreSQL（集群）
- **缓存**: 内存LRU / Redis
- **存储**: 本地文件系统 / S3兼容对象存储
- **监控**: Prometheus + Grafana
- **日志**: 结构化日志（JSON格式）

## 系统架构

### 整体架构层次

| 架构层次 | 核心组件 | Agents角色 | 技术实现 |
|----------|----------|------------|----------|
| **接入层** | Gin框架、认证中间件、限流模块 | HTTP请求代理、认证代理、限流代理 | Gin路由、JWT认证、令牌桶限流 |
| **核心内核** | 仓库引擎、缓存管理器、插件调度器 | 仓库管理代理、缓存代理、插件生命周期代理 | Wire DI、LRU缓存、Plugin系统 |
| **插件层** | 格式插件、存储插件、集成插件 | 格式处理代理、存储代理、集成代理 | Go Plugin、接口抽象 |
| **存储层** | 本地文件系统、S3/MinIO客户端 | 存储访问代理、对象存储代理 | 文件系统API、S3 SDK |

## 项目目录结构规范

go-nexus采用Go标准项目布局，结合Wire依赖注入框架的特殊需求：

```
go-nexus/
├── cmd/                        # 主程序入口
│   ├── server/                 # 服务端程序
│   │   ├── main.go            # 程序入口
│   │   ├── wire.go            # Wire依赖注入配置
│   │   └── wire_gen.go        # Wire生成的代码
│   └── cli/                   # 命令行工具
│       └── main.go
├── internal/                   # 私有应用代码
│   ├── app/                   # 应用程序层
│   │   ├── app.go            # 应用程序主结构
│   │   └── providers.go      # Wire提供者
│   ├── config/               # 配置管理
│   │   ├── config.go         # 配置结构定义
│   │   ├── loader.go         # 配置加载器
│   │   └── providers.go      # Wire配置提供者
│   ├── handler/              # HTTP处理器
│   │   ├── repository.go     # 仓库相关接口
│   │   ├── artifact.go       # 制品相关接口
│   │   └── middleware/       # 中间件
│   ├── service/              # 业务逻辑层
│   │   ├── repository.go     # 仓库服务接口
│   │   ├── artifact.go       # 制品服务接口
│   │   ├── impl/             # 服务实现
│   │   └── providers.go      # Wire服务提供者
│   ├── repository/           # 数据仓储层
│   │   ├── model/            # 数据模型
│   │   ├── dao/              # 数据访问对象
│   │   └── providers.go      # Wire数据层提供者
│   ├── plugin/               # 插件管理
│   │   ├── manager.go        # 插件管理器
│   │   ├── loader.go         # 插件加载器
│   │   └── providers.go      # Wire插件提供者
│   └── storage/              # 存储抽象层
│       ├── interface.go      # 存储接口定义
│       ├── filesystem/       # 文件系统实现
│       ├── s3/               # S3存储实现
│       └── providers.go      # Wire存储提供者
├── pkg/                       # 公共库代码
│   ├── logger/               # 日志工具
│   ├── cache/                # 缓存工具
│   ├── plugin/               # 插件框架
│   │   ├── interface.go      # 插件接口定义
│   │   └── registry.go       # 插件注册表
│   └── utils/                # 工具函数
├── api/                       # API定义
│   ├── openapi/              # OpenAPI规范
│   └── proto/                # Protocol Buffers定义
├── resource/                  # 运行资源目录
│   ├── data/                 # 运行数据（如数据库、持久化文件等）
│   │   └── go-nexus.db       # 默认数据库
│   ├── configs/              # 配置文件
│   │   ├── config.yaml       # 默认配置
│   │   ├── config-dev.yaml   # 开发环境配置
│   │   └── config-prod.yaml  # 生产环境配置
│   ├── plugins/              # 插件目录（如maven、npm等插件包）
│   └── web/                  # 前端静态资源与模板
├── scripts/                   # 构建和部署脚本
│   ├── build.sh              # 构建脚本
│   ├── deploy.sh             # 部署脚本
│   └── init.sql              # 数据库初始化脚本
├── deployments/               # 部署配置
│   ├── docker/               # Docker相关文件
│   │   └── Dockerfile
│   ├── kubernetes/           # K8s部署文件
│   └── helm/                 # Helm Chart
├── test/                      # 测试文件
│   ├── integration/          # 集成测试
│   ├── e2e/                  # 端到端测试
│   └── fixtures/             # 测试数据
├── docs/                      # 文档
├── .agents/                   # AI Agents规范文档
│   ├── config-standards.md   # 配置规范
│   ├── logging-standards.md  # 日志规范
│   ├── api-standards.md      # Web接口规范
│   ├── development-standards.md # 代码开发规范
│   ├── testing-standards.md  # 测试规范
│   ├── deployment-standards.md # 部署规范
│   └── plugin-standards.md   # 插件开发规范
├── go.mod                     # Go模块定义
├── go.sum                     # Go模块校验
├── Makefile                   # 构建脚本
├── README.md                  # 项目说明
├── CHANGELOG.md               # 更新日志
└── LICENSE                    # 开源许可证
```

### Wire依赖注入特殊要求

1. **Wire文件组织**：
   - `cmd/server/wire.go`: 主应用Wire配置，包含`//go:build wireinject`标签
   - `internal/*/providers.go`: 各模块的Wire提供者函数
   - `cmd/server/wire_gen.go`: Wire自动生成的文件，不可手动编辑

2. **提供者模式**：
   - 每个模块都有对应的`providers.go`文件
   - 使用`wire.NewSet`组织相关的提供者
   - 接口绑定使用`wire.Bind`

3. **构建标签**：
   - Wire配置文件使用`//go:build wireinject`标签
   - 生成的文件使用`//go:build !wireinject`标签

## 开发规范索引

为了方便AI理解和遵循项目规范，以下规范文档已分拆到`.agents`目录：

### 📋 [配置规范](.agents/config-standards.md)
- YAML配置文件格式和结构
- 环境变量覆盖规则
- 配置验证和加载顺序
- 数据库、缓存、存储等配置项

### 📝 [日志规范](.agents/logging-standards.md)  
- 结构化日志格式（JSON/文本）
- 日志级别定义和使用场景
- 业务日志和技术日志规范
- 错误码定义和敏感信息处理

### 🌐 [Web接口规范](.agents/api-standards.md)
- RESTful API设计规范
- 统一请求响应格式
- HTTP状态码使用规范
- OpenAPI文档和认证授权

### 💻 [代码开发规范](.agents/development-standards.md)
- Go代码风格和命名规范
- Wire依赖注入最佳实践
- 错误处理和接口设计
- 包导入和文件组织

### 🧪 [测试规范](.agents/testing-standards.md)
- 单元测试、集成测试、性能测试
- 测试文件组织和命名规范
- Mock和Stub使用规范
- 测试覆盖率要求

### 🚀 [部署规范](.agents/deployment-standards.md)
- Docker、Kubernetes、Helm部署
- 环境配置和资源限制
- 监控、日志收集配置
- 备份恢复策略

### 🔌 [插件开发规范](.agents/plugin-standards.md)
- 插件接口定义和实现规范
- 格式插件、存储插件开发
- 插件管理和配置规范
- 插件测试框架使用

## 快速上手指南

### 开发环境搭建
```bash
# 1. 克隆项目
git clone https://github.com/laolishu/go-nexus.git
cd go-nexus

# 2. 安装依赖
go mod download

# 3. 安装Wire工具
go install github.com/google/wire/cmd/wire@latest

# 4. 生成Wire代码
make wire-gen

# 5. 编译项目
make build

# 6. 运行测试
make test

# 7. 启动开发服务器
./bin/go-nexus server --config configs/config-dev.yaml
```

### 核心概念理解

1. **仓库类型**：
   - Proxy仓库：缓存远程仓库内容
   - Hosted仓库：存储私有制品
   - Group仓库：聚合多个仓库

2. **插件系统**：
   - Format插件：处理不同格式的制品（Maven、npm等）
   - Storage插件：支持不同存储后端（文件系统、S3等）
   - Integration插件：集成第三方工具（CI/CD、监控等）

3. **依赖注入**：
   - 使用Wire进行编译时依赖注入
   - 接口驱动设计，便于测试和扩展
   - 配置驱动的组件初始化