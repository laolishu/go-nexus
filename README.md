# go-nexus：轻量云原生仓库管理工具

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg)](https://golang.org/)
[![Docker Image](https://img.shields.io/badge/docker-image-available-green.svg)](https://hub.docker.com/r/laolishu/go-nexus)


## 项目概述

`go-nexus` 是一款基于 Golang 开发的轻量云原生仓库管理工具，专为中小团队及云原生环境设计，旨在简化依赖管理流程。

它针对传统重量级工具（如 Nexus Repository Manager、Artifactory）的痛点，提供以下核心优势：
- **轻量部署**：单二进制文件运行，无额外运行时依赖（无需 JVM，无需复杂配置）；
- **云原生设计**：原生支持 Kubernetes、Docker 及对象存储（S3/MinIO）；
- **多格式兼容**：无缝支持 Maven、npm、Docker、Helm 等主流依赖格式；
- **高性能**：基于 Golang Goroutine 模型优化并发请求处理，资源占用低。


## 项目背景

传统仓库管理工具诞生于单体架构时代，在现代云原生场景中存在明显局限：
- **资源消耗高**：启动内存需数 GB，启动速度慢，不适合边缘环境或资源受限场景；
- **部署维护复杂**：需手动调优 JVM 参数，配置流程繁琐，运维成本高；
- **云原生适配差**：对 Kubernetes、CI/CD 流水线的集成支持滞后，容器镜像体积庞大。

`go-nexus` 正是为解决这些问题而生：
- 面向中小团队及边缘环境，满足「低资源占用、快速部署」需求；
- 贴合现代 DevOps 实践，支持「基础设施即代码（IaC）」与自动化流水线；
- 提供高性价比方案：开源核心功能免费，企业级功能按需扩展。


## 核心功能
| 功能模块                | 说明                                                                 |
|-------------------------|----------------------------------------------------------------------|
| **多类型仓库管理**      | 支持代理仓库（缓存公共仓库）、宿主仓库（存储私有资产）、仓库组（聚合多仓库） |
| **多格式依赖支持**      | 目前内置支持 Maven、npm 依赖管理；Docker、Helm 等格式将在后期通过插件系统扩展支持，同时预留接口支持用户自定义格式插件（如 PyPI、Cargo 等） |
| **云原生部署能力**      | 单二进制文件、Docker 镜像（体积 < 100MB）、K8s 部署专用 Helm Chart   |
| **高并发处理**          | 单实例支持 10k+ 并发请求，内存占用仅为传统工具的 1/5                  |
| **安全管控**            | 支持 HTTPS 加密传输、API Token 认证、基于角色的访问控制（RBAC）       |
| **可扩展性**            | 插件化架构，支持自定义依赖格式处理或第三方工具集成                   |


## 技术架构

`go-nexus` 采用「微内核 + 插件」架构，在保持核心轻量化的同时，兼顾功能扩展性。各层次的核心组件与功能如下：

| 架构层次   | 核心组件                                  | 功能说明                                                                 |
|------------|-------------------------------------------|--------------------------------------------------------------------------|
| 接入层     | Gin 框架、认证中间件、限流模块            | 提供高性能 HTTP/HTTPS 服务，处理 TLS 加密传输、请求校验、身份认证与流量控制；对接外部请求与内核层的交互。 |
| 核心内核   | 仓库引擎、缓存管理器、插件调度器          | 负责仓库全生命周期管理（创建/删除/配置）；内置 LRU 缓存（支持 Redis 分布式扩展）；管理插件加载、卸载与调用逻辑，保持核心体积轻量化（<20MB）。 |
| 插件层     | 格式插件、存储插件、集成插件              | 按需加载，不占用核心资源：<br>- 格式插件：当前支持 Maven、npm，后期将通过插件扩展 Docker、Helm 等；<br>- 存储插件：支持本地文件系统、S3/MinIO 等对象存储；<br>- 集成插件：对接 CI/CD 工具（如 GitLab CI）、监控工具（如 Prometheus）。 |
| 存储层     | 本地文件系统、S3/MinIO 客户端             | 负责依赖包文件的实际存储与读写，单节点默认用本地文件系统，大规模场景可通过插件切换至分布式对象存储。 |
| 元数据层   | SQLite（默认）、PostgreSQL 驱动           | 存储仓库配置、依赖版本信息、用户权限等元数据；单节点用 SQLite（零部署成本），集群场景支持 PostgreSQL 实现数据共享。 |

### 关键组件说明：
- **接入层**：基于 [Gin](https://gin-gonic.com/) 框架实现高性能 HTTP 服务，支持 TLS 终止、请求校验与限流；
- **核心内核**：负责仓库生命周期管理、LRU 缓存（可选集成 Redis）、插件生命周期调度；
- **存储层**：默认支持本地文件系统，可通过插件扩展至 S3 兼容对象存储（适合大规模场景）；
- **元数据存储**：单节点场景用 SQLite（零部署成本），集群场景支持 PostgreSQL；
- **插件系统**：基于 Golang `plugin` 包实现动态插件加载，支持非核心功能热重载。


## 快速开始

### 前置依赖
- 编译源码：Go 1.21+
- 容器部署：Docker
- K8s 部署：Kubernetes 集群（1.20+）、Helm 3.5+

### 1. 二进制方式运行
```bash
# 从 GitHub Releases 下载最新版本（替换为实际版本号）
wget https://github.com/laolishu/go-nexus/releases/latest/download/go-nexus-linux-amd64

# 赋予可执行权限
chmod +x go-nexus-linux-amd64

# 启动服务（默认端口 8081，使用默认配置）
./go-nexus-linux-amd64 server
```

### 2. Docker 方式运行
```bash
docker run -d -p 8081:8081 \
  -v ./go-nexus-data:/var/lib/go-nexus \  # 挂载数据目录，持久化存储
  laolishu/go-nexus:latest
```

### 3. Kubernetes 方式运行
```bash
# 添加 Helm 仓库
helm repo add go-nexus https://laolishu.github.io/go-nexus/charts

# 创建命名空间
kubectl create namespace go-nexus

# 安装 go-nexus（默认使用 SQLite，支持通过参数切换为 PostgreSQL）
helm install go-nexus go-nexus/go-nexus --namespace go-nexus
```

服务启动后，访问 `http://localhost:8081` 即可进入 Web 管理界面，或参考 [API 文档](docs/api.md) 通过接口管理仓库。


## 开发规范

### 代码风格
- 遵循 [Go 代码审查规范](https://github.com/golang/go/wiki/CodeReviewComments)；
- 使用 `gofmt` 格式化代码，`golint` 检查代码风格（已在 `Makefile` 中配置相关命令）；
- 公共函数、结构体需添加中文注释，说明功能与使用场景。

### 分支策略
- `main`：稳定发布分支（受保护，需 PR 审批才能合并）；
- `develop`：开发分支，用于集成新功能；
- `feature/xxx`：功能分支（从 `develop` 分支创建，完成后合并回 `develop`）；
- `hotfix/xxx`：紧急修复分支（从 `main` 分支创建，修复后同步合并到 `main` 和 `develop`）。

### 提交信息格式
遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范，格式如下：
```
<类型>(<范围>): <描述>

[可选的详细说明]

[可选的 Footer（如关闭 Issue）]
```
示例：`feat(repo): 新增 Docker 代理仓库功能`

### 测试要求
- 单元测试：覆盖核心逻辑（运行命令 `make test`）；
- 集成测试：验证仓库操作、依赖上传下载等流程（运行命令 `make test-integration`）；
- 代码覆盖率：目标 ≥ 80%（通过 `make coverage` 查看覆盖率报告）。


## 编译与发布

### 从源码编译
```bash
# 克隆仓库
git clone https://github.com/laolishu/go-nexus.git
cd go-nexus

# 编译二进制文件（输出到 ./bin/go-nexus）
make build

# 构建 Docker 镜像（指定版本标签）
make docker-build TAG=v0.1.0
```

### 发布流程
1. 在 `VERSION` 文件中更新版本号（遵循语义化版本：`vX.Y.Z`）；
2. 运行 `make release` 生成更新日志并创建 Git 标签；
3. 推送标签到 GitHub，触发 GitHub Actions 流水线：
   - 自动编译 Linux/Windows/macOS 多平台二进制文件；
   - 构建并推送 Docker 镜像到 Docker Hub；
   - 在 GitHub Releases 发布包含所有产物的版本包。


## 许可证

`go-nexus` 基于 [Apache License 2.0](LICENSE) 开源，核心功能免费使用。企业级功能（如高级 RBAC、集群模式、专属技术支持）需通过商业许可获取。


## 贡献指南

欢迎所有开发者参与贡献！具体流程请参考 [CONTRIBUTING.md](CONTRIBUTING.md)，包括：
- 提交 Issue（反馈 Bug 或提出功能建议）；
- 提交 Pull Request（修复 Bug 或开发新功能）；
- 开发自定义插件（扩展依赖格式或集成第三方工具）。


## 联系我们
- 项目维护者：可通过 GitHub Issues 或邮件联系（仓库 `README` 可补充具体邮箱）；
- GitHub 仓库：[https://github.com/laolishu/go-nexus](https://github.com/laolishu/go-nexus)；
- Issue 反馈：[https://github.com/laolishu/go-nexus/issues](https://github.com/laolishu/go-nexus/issues)。


---

*注：`go-nexus` 与 Sonatype 公司的 Nexus Repository Manager 无关联，是独立开发的开源项目。*


### 需补充确认的信息（完善文档时可补充）
1. 企业版具体包含哪些高级功能？是否需要明确收费模式（如按团队规模/存储量订阅）？
2. 测试环境是否有特定依赖（如 PostgreSQL 版本、Redis 版本）？
3. 是否需要添加第三方依赖列表（如使用的开源库及其用途）？
4. 技术架构图是否需要提供图片链接（如 GitHub 仓库中的架构图 PNG 文件）？
5. Web UI 的使用指南是否需要单独补充文档（如仓库创建、权限配置步骤）？这条消息已经在编辑器中准备就绪。你想如何调整这篇文档?请随时告诉我。