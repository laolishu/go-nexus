# Web接口规范

## RESTful API设计规范

### URL设计规则
- 使用名词，避免动词
- 资源路径使用复数形式
- 使用连字符（-）分隔单词，避免下划线
- 版本号放在URL路径中：`/api/v1/`

### HTTP方法使用
- `GET`: 获取资源
- `POST`: 创建资源
- `PUT`: 完全更新资源
- `PATCH`: 部分更新资源
- `DELETE`: 删除资源

### API路径规范
```
# 仓库管理
GET    /api/v1/repositories           # 获取仓库列表
POST   /api/v1/repositories           # 创建仓库
GET    /api/v1/repositories/{id}      # 获取指定仓库
PUT    /api/v1/repositories/{id}      # 更新仓库
DELETE /api/v1/repositories/{id}      # 删除仓库

# Artifact管理
GET    /api/v1/repositories/{id}/artifacts      # 获取artifact列表
POST   /api/v1/repositories/{id}/artifacts      # 上传artifact
GET    /api/v1/repositories/{id}/artifacts/{path} # 下载artifact
DELETE /api/v1/repositories/{id}/artifacts/{path} # 删除artifact

# 用户管理
GET    /api/v1/users                  # 获取用户列表
POST   /api/v1/users                  # 创建用户
GET    /api/v1/users/{id}             # 获取用户信息
PUT    /api/v1/users/{id}             # 更新用户
DELETE /api/v1/users/{id}             # 删除用户

# 认证相关
POST   /api/v1/auth/login             # 用户登录
POST   /api/v1/auth/logout            # 用户登出
POST   /api/v1/auth/refresh           # 刷新Token
```

## 请求响应规范

### 请求头规范
```http
Content-Type: application/json
Authorization: Bearer {jwt_token}
X-Trace-ID: {trace_id}
User-Agent: go-nexus-client/1.0.0
```

### 统一响应格式
```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    // 具体数据
  },
  "meta": {
    "trace_id": "abc123def456",
    "timestamp": "2025-09-28T10:30:00Z",
    "version": "v0.1.0"
  }
}
```

### 错误响应格式
```json
{
  "success": false,
  "code": 4001,
  "message": "Repository not found",
  "error": {
    "type": "REPOSITORY_NOT_FOUND",
    "details": "Repository with ID 'maven-public' does not exist",
    "field": "repository_id"
  },
  "meta": {
    "trace_id": "abc123def456",
    "timestamp": "2025-09-28T10:30:00Z",
    "version": "v0.1.0"
  }
}
```

### 分页响应格式
```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "items": [...],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 100,
      "total_pages": 5
    }
  }
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

## 接口文档规范

### OpenAPI 3.0规范
- 使用Swagger注解生成API文档
- 包含完整的请求/响应示例
- 标注必填参数和可选参数
- 提供错误码说明

### 接口注释示例
```go
// CreateRepository 创建仓库
// @Summary 创建新仓库
// @Description 创建一个新的Maven或npm仓库
// @Tags repositories
// @Accept json
// @Produce json
// @Param repository body dto.CreateRepositoryRequest true "仓库信息"
// @Success 201 {object} dto.RepositoryResponse "仓库创建成功"
// @Failure 400 {object} dto.ErrorResponse "参数错误"
// @Failure 409 {object} dto.ErrorResponse "仓库已存在"
// @Router /api/v1/repositories [post]
func (c *RepositoryController) CreateRepository(ctx *gin.Context) {
    // 实现代码
}
```

## 认证授权规范

### JWT Token规范
- Header: `Authorization: Bearer {jwt_token}`
- Token过期时间: 24小时
- Refresh Token过期时间: 7天
- Token包含用户ID、角色、权限等信息

### API权限控制
- 管理员: 所有操作权限
- 开发者: 仓库读写权限
- 只读用户: 仓库只读权限
- 匿名用户: 公共仓库只读权限

### 限流规范
- 全局限流: 1000 req/min
- 用户限流: 100 req/min
- IP限流: 200 req/min
- 上传限流: 10 upload/min