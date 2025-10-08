# go-nexus Web接口规范

## 项目概述
go-nexus是一款基于Golang开发的轻量云原生仓库管理工具，采用微内核+插件架构，使用Wire依赖注入框架和Gin Web框架，支持Maven、npm等多种制品格式。

## RESTful API设计规范

### URL设计规则
- 使用名词，避免动词
- 资源路径使用复数形式
- 使用连字符（-）分隔单词，避免下划线
- 版本号放在URL路径中：`/api/v1/`
- 支持通配符路径用于制品下载：`/api/v1/repositories/{id}/artifacts/*path`

### HTTP方法使用
- `GET`: 获取资源
- `POST`: 创建资源
- `PUT`: 完全更新资源
- `PATCH`: 部分更新资源
- `DELETE`: 删除资源
- `OPTIONS`: CORS预检请求

### 核心API路径规范

#### 系统信息
```
GET    /                              # 系统信息和API导航
GET    /api                          # API版本信息
GET    /health                       # 健康检查接口
```

#### 仓库管理
```
GET    /api/v1/repositories           # 获取仓库列表
POST   /api/v1/repositories           # 创建仓库
GET    /api/v1/repositories/{id}      # 获取指定仓库
PUT    /api/v1/repositories/{id}      # 更新仓库
DELETE /api/v1/repositories/{id}      # 删除仓库
```

#### Artifact制品管理
```
GET    /api/v1/repositories/{id}/artifacts        # 获取制品列表
POST   /api/v1/repositories/{id}/artifacts        # 上传制品
GET    /api/v1/repositories/{id}/artifacts/*path  # 下载制品（支持通配符路径）
DELETE /api/v1/repositories/{id}/artifacts/*path  # 删除制品（支持通配符路径）
```

#### 用户管理（规划中）
```
GET    /api/v1/users                  # 获取用户列表
POST   /api/v1/users                  # 创建用户
GET    /api/v1/users/{id}             # 获取用户信息
PUT    /api/v1/users/{id}             # 更新用户
DELETE /api/v1/users/{id}             # 删除用户
```

#### 认证相关（规划中）
```
POST   /api/v1/auth/login             # 用户登录
POST   /api/v1/auth/logout            # 用户登出
POST   /api/v1/auth/refresh           # 刷新Token
```

#### 插件管理（规划中）
```
GET    /api/v1/plugins                # 获取插件列表
POST   /api/v1/plugins                # 安装插件
GET    /api/v1/plugins/{id}           # 获取插件信息
DELETE /api/v1/plugins/{id}           # 卸载插件
```

## 路由注册架构

### 全局路由管理
go-nexus使用全局变量和统一注册函数管理路由，位于 `core/global/router.go` 和 `core/web/router_utils.go`：

```go
// 全局路由变量
var (
    RootRouter  *gin.Engine      // 根路由引擎实例
    APIv1Router *gin.RouterGroup // API v1路由组实例
)

// 路由注册函数
RegisterRootHandle(httpMethod, relativePath, handlers...)    // 注册根路径路由
RegisterApiHandle(httpMethod, relativePath, handlers...)     // 注册API路由
RegisterMiddleware(middleware...)                            // 注册全局中间件
```

### 路由初始化流程
1. `SetupRoutes()` 初始化全局路由变量
2. `SetupMiddlewares()` 注册CORS、请求ID等全局中间件
3. `SetupRootRoutes()` 注册根路径和API导航路由
4. `SetupHealthCheck()` 注册健康检查路由
5. `SetupAPIv1Routes()` 注册业务API路由

## 请求响应规范

### 请求头规范
```http
Content-Type: application/json
Authorization: Bearer {jwt_token}
X-Request-ID: {request_id}
User-Agent: go-nexus-client/1.0.0
Accept: application/json
```

### CORS支持
系统自动支持跨域请求，响应头包含：
```http
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Origin, Content-Type, Accept, Authorization, X-Request-ID
```

## 标准响应体规范

### 标准响应格式

项目使用统一的标准响应体格式，确保所有API接口返回数据的一致性和可追踪性：

```go
type StandardResponse struct {
    Code      int         `json:"code"`      // 响应代码：0表示成功，非0表示错误
    Msg       string      `json:"msg"`       // 响应信息：成功或错误描述
    Data      interface{} `json:"data"`      // 响应数据：具体的业务数据
    RequestID string      `json:"requestId"` // 请求ID：用于追踪和调试
}
```

### 响应工具函数

位于 `core/web/respons_utils.go`，提供统一的响应处理：

**成功响应**：
- `Success(c *gin.Context, data interface{})` - 标准成功响应（HTTP 200）
- `SuccessWithMsg(c *gin.Context, msg string, data interface{})` - 带自定义消息的成功响应

**错误响应**：
- `Error(c *gin.Context, httpStatus, code int, msg string)` - 通用错误响应
- `BadRequest(c *gin.Context, msg string)` - 400 错误
- `Unauthorized(c *gin.Context, msg string)` - 401 错误
- `Forbidden(c *gin.Context, msg string)` - 403 错误
- `NotFound(c *gin.Context, msg string)` - 404 错误
- `InternalServerError(c *gin.Context, msg string)` - 500 错误

**使用示例**：
```go
// 成功响应
Success(c, gin.H{"status": "ok"})

// 错误响应
InternalServerError(c, "Failed to get system info: "+err.Error())
```

### 请求ID处理

系统自动处理请求ID，支持以下方式：
1. 从 `X-Request-ID` Header 获取客户端传入的请求ID
2. 从 Gin Context 中获取中间件设置的请求ID
3. 自动生成 UUID 作为请求ID（兜底机制）

### 响应示例

**系统信息响应**（GET /）：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "name": "go-nexus",
    "description": "轻量云原生仓库管理工具",
    "version": "v0.5.0",
    "api_docs": "/api/v1/docs",
    "health": "/health"
  },
  "requestId": "550e8400-e29b-41d4-a716-446655440000"
}
```

**健康检查响应**（GET /health）：
```json
{
  "code": 0,
  "msg": "success", 
  "data": {
    "status": "ok",
    "cpu_usage": 12.5,
    "memory_usage": 45.2
  },
  "requestId": "550e8400-e29b-41d4-a716-446655440001"
}
```

**API信息响应**（GET /api/v1）：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "name": "go-nexus API v1",
    "version": "1.0.0",
    "endpoints": {
      "repositories": "/api/v1/repositories",
      "artifacts": "/api/v1/repositories/{id}/artifacts"
    }
  },
  "requestId": "550e8400-e29b-41d4-a716-446655440002"
}
```

**错误响应示例**：
```json
{
  "code": 500,
  "msg": "Failed to get system info: connection failed",
  "data": null,
  "requestId": "550e8400-e29b-41d4-a716-446655440003"
}
```

### 仓库管理响应格式

**仓库列表响应**（GET /api/v1/repositories）：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "items": [
      {
        "id": "repo-001",
        "name": "maven-central",
        "type": "proxy",
        "format": "maven",
        "url": "https://repo1.maven.org/maven2/",
        "storage": "filesystem",
        "created_at": "2025-10-08T10:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 1,
      "total_pages": 1
    }
  },
  "requestId": "550e8400-e29b-41d4-a716-446655440004"
}
```

### 制品管理响应格式

**制品列表响应**（GET /api/v1/repositories/{id}/artifacts）：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "items": [
      {
        "path": "com/example/demo/1.0.0/demo-1.0.0.jar",
        "size": 2048576,
        "checksum": "sha1:abc123...",
        "last_modified": "2025-10-08T15:30:00Z"
      }
    ]
  },
  "requestId": "550e8400-e29b-41d4-a716-446655440005"
}
```

## HTTP状态码规范

### 成功响应
- `200 OK`: 请求成功
- `201 Created`: 资源创建成功
- `204 No Content`: 删除成功，无返回内容

### 客户端错误
- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 未认证
- `403 Forbidden`: 无权限
- `404 Not Found`: 资源不存在
- `409 Conflict`: 资源冲突
- `422 Unprocessable Entity`: 请求格式正确但语义错误

### 服务端错误
- `500 Internal Server Error`: 服务器内部错误
- `502 Bad Gateway`: 网关错误
- `503 Service Unavailable`: 服务不可用
- `504 Gateway Timeout`: 网关超时

## 仓库类型规范

### 支持的仓库类型
- **Proxy仓库**：缓存远程仓库内容，如maven-central、npm-registry
- **Hosted仓库**：存储私有制品，如private-maven、company-npm
- **Group仓库**：聚合多个仓库，提供统一访问入口

### 支持的制品格式
- **Maven**: Java项目依赖管理（通过格式插件）
- **npm**: Node.js包管理（通过格式插件）
- **Docker**: 容器镜像（规划中，通过插件扩展）
- **Helm**: Kubernetes包管理（规划中，通过插件扩展）

## 存储后端规范

### 支持的存储类型
- **本地文件系统**: 适用于单节点部署
- **S3兼容存储**: 支持AWS S3、MinIO等对象存储，适用于云原生部署

### 存储路径规范
```
# Maven制品路径
/repositories/{repo-id}/com/example/demo/1.0.0/demo-1.0.0.jar
/repositories/{repo-id}/com/example/demo/1.0.0/demo-1.0.0.pom

# npm制品路径
/repositories/{repo-id}/@scope/package-name/-/package-name-1.0.0.tgz
/repositories/{repo-id}/package-name/-/package-name-1.0.0.tgz
```

## 接口文档规范

### OpenAPI 3.0规范
- API文档访问地址：`/api/v1/docs`
- 使用Swagger注解生成API文档
- 包含完整的请求/响应示例
- 标注必填参数和可选参数
- 提供错误码说明

### 接口注释示例
```go
// ListRepositories 获取仓库列表
// @Summary 获取仓库列表
// @Description 获取系统中所有仓库的列表，支持分页和过滤
// @Tags repositories
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页大小" default(20)
// @Param type query string false "仓库类型" Enums(proxy,hosted,group)
// @Success 200 {object} StandardResponse{data=RepositoryListResponse} "成功"
// @Failure 400 {object} StandardResponse "参数错误"
// @Failure 500 {object} StandardResponse "服务器错误"
// @Router /api/v1/repositories [get]
func (h *RepositoryHandler) ListRepositories(c *gin.Context) {
    // 实现代码
}
```

## 中间件规范

### 全局中间件
使用 `RegisterMiddleware()` 函数注册，包含：

1. **CORS中间件**: 自动处理跨域请求
2. **请求ID中间件**: 自动生成或提取请求追踪ID
3. **日志中间件**: 结构化请求日志记录（规划中）
4. **限流中间件**: 请求频率限制（规划中）
5. **认证中间件**: JWT Token验证（规划中）

### 请求追踪
每个请求都包含唯一的RequestID，支持：
- 客户端传入：`X-Request-ID` Header
- 系统生成：UUID格式
- 响应返回：`requestId` 字段

## 认证授权规范（规划中）

### JWT Token规范
- Header: `Authorization: Bearer {jwt_token}`
- Token过期时间: 24小时
- Refresh Token过期时间: 7天
- Token包含用户ID、角色、权限等信息

### API权限控制
- **管理员**: 所有操作权限
- **开发者**: 仓库读写权限
- **只读用户**: 仓库只读权限
- **匿名用户**: 公共仓库只读权限

### 限流规范（规划中）
- 全局限流: 10k req/min（基于Go高并发特性）
- 用户限流: 1k req/min
- IP限流: 2k req/min
- 上传限流: 100 upload/min

## 插件扩展规范

### 插件类型
- **格式插件**: 处理不同制品格式（Maven、npm、Docker等）
- **存储插件**: 支持不同存储后端（文件系统、S3等）
- **集成插件**: 集成第三方工具（CI/CD、监控等）

### 插件接口
位于 `pkg/plugin/interface.go`，定义标准的插件接口规范。

## 性能规范

### 并发性能
- 基于Goroutine模型，单实例支持10k+并发请求
- 内存占用仅为传统工具的1/5
- 支持水平扩展和负载均衡

### 缓存策略
- 内存LRU缓存：热点数据缓存
- Redis缓存：分布式缓存（可选）
- 制品缓存：代理仓库内容缓存

## 监控和日志规范

### 健康检查
- 端点：`GET /health`
- 返回系统状态、CPU使用率、内存使用率
- 支持Kubernetes健康检查

### 监控指标
- Prometheus指标暴露
- Grafana仪表板
- 请求量、响应时间、错误率等关键指标

### 日志规范
- 结构化日志（JSON格式）
- 请求追踪ID关联
- 分级日志：DEBUG、INFO、WARN、ERROR
- 详见 `.agents/logging-standards.md`