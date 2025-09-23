# 权限控制 (Casbin)

## 概述

go-nexus使用Casbin实现基于角色的访问控制(RBAC)，提供灵活强大的权限管理功能。

## RBAC模型定义

### 模型文件 (rbac_model.conf)
```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && r.act == p.act
```

### 模型解释
- **sub**: 主体(用户)
- **obj**: 对象(资源路径)
- **act**: 动作(操作类型)
- **keyMatch2**: 支持路径通配符匹配

## 权限系统初始化

### Casbin初始化
```go
func InitCasbin(modelFile, policyDB string) (*casbin.Enforcer, error) {
    // 创建模型
    m, err := model.NewModelFromFile(modelFile)
    if err != nil {
        return nil, fmt.Errorf("failed to load model: %w", err)
    }
    
    // 创建适配器
    adapter, err := gormadapter.NewAdapter("sqlite3", policyDB)
    if err != nil {
        return nil, fmt.Errorf("failed to create adapter: %w", err)
    }
    
    // 创建执行器
    enforcer, err := casbin.NewEnforcer(m, adapter)
    if err != nil {
        return nil, fmt.Errorf("failed to create enforcer: %w", err)
    }
    
    // 加载默认策略
    if err := loadDefaultPolicies(enforcer); err != nil {
        return nil, fmt.Errorf("failed to load default policies: %w", err)
    }
    
    return enforcer, nil
}
```

### 默认策略加载
```go
func loadDefaultPolicies(e *casbin.Enforcer) error {
    // 管理员权限
    adminPolicies := [][]string{
        {"admin", "repository:maven:*", "read"},
        {"admin", "repository:maven:*", "write"},
        {"admin", "repository:maven:*", "delete"},
        {"admin", "repository:npm:*", "read"},
        {"admin", "repository:npm:*", "write"},
        {"admin", "repository:npm:*", "delete"},
        {"admin", "agents:*", "manage"},
        {"admin", "users:*", "manage"},
        {"admin", "system:*", "manage"},
    }
    
    // 开发者权限
    developerPolicies := [][]string{
        {"developer", "repository:maven:*", "read"},
        {"developer", "repository:maven:releases", "write"},
        {"developer", "repository:maven:snapshots", "write"},
        {"developer", "repository:npm:*", "read"},
        {"developer", "repository:npm:hosted", "write"},
    }
    
    // 只读用户权限
    readonlyPolicies := [][]string{
        {"readonly", "repository:maven:*", "read"},
        {"readonly", "repository:npm:*", "read"},
    }
    
    // 添加策略
    for _, policy := range adminPolicies {
        if _, err := e.AddPolicy(policy); err != nil {
            return err
        }
    }
    
    for _, policy := range developerPolicies {
        if _, err := e.AddPolicy(policy); err != nil {
            return err
        }
    }
    
    for _, policy := range readonlyPolicies {
        if _, err := e.AddPolicy(policy); err != nil {
            return err
        }
    }
    
    return nil
}
```

## 中间件实现

### 认证中间件
```go
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(401, gin.H{"error": "missing authorization header"})
            c.Abort()
            return
        }
        
        tokenString = strings.TrimPrefix(tokenString, "Bearer ")
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return []byte(jwtSecret), nil
        })
        
        if err != nil || !token.Valid {
            c.JSON(401, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }
        
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.JSON(401, gin.H{"error": "invalid token claims"})
            c.Abort()
            return
        }
        
        c.Set("user_id", claims["sub"])
        c.Set("username", claims["username"])
        c.Set("roles", claims["roles"])
        c.Next()
    }
}
```

### 权限检查中间件
```go
func PermissionMiddleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, exists := c.Get("user_id")
        if !exists {
            c.JSON(401, gin.H{"error": "user not authenticated"})
            c.Abort()
            return
        }
        
        // 构建资源路径
        resource := buildResourcePath(c)
        action := mapHTTPMethodToAction(c.Request.Method)
        
        // 检查权限
        allowed, err := enforcer.Enforce(userID, resource, action)
        if err != nil {
            c.JSON(500, gin.H{"error": "permission check failed"})
            c.Abort()
            return
        }
        
        if !allowed {
            c.JSON(403, gin.H{"error": "insufficient permissions"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}

func buildResourcePath(c *gin.Context) string {
    path := c.Request.URL.Path
    
    // 映射URL路径到资源标识
    if strings.HasPrefix(path, "/api/v1/repository/maven/") {
        return "repository:maven:" + extractRepositoryName(path)
    } else if strings.HasPrefix(path, "/api/v1/repository/npm/") {
        return "repository:npm:" + extractRepositoryName(path)
    } else if strings.HasPrefix(path, "/api/v1/agents/") {
        return "agents:" + extractAgentName(path)
    } else if strings.HasPrefix(path, "/api/v1/users/") {
        return "users:*"
    } else if strings.HasPrefix(path, "/api/v1/system/") {
        return "system:*"
    }
    
    return path
}

func mapHTTPMethodToAction(method string) string {
    switch method {
    case "GET", "HEAD":
        return "read"
    case "POST", "PUT", "PATCH":
        return "write"
    case "DELETE":
        return "delete"
    default:
        return "unknown"
    }
}
```

## 角色和权限管理

### 用户角色管理
```go
type RoleManager struct {
    enforcer *casbin.Enforcer
    logger   *slog.Logger
}

func NewRoleManager(enforcer *casbin.Enforcer, logger *slog.Logger) *RoleManager {
    return &RoleManager{
        enforcer: enforcer,
        logger:   logger,
    }
}

// 为用户分配角色
func (rm *RoleManager) AssignRole(userID, role string) error {
    added, err := rm.enforcer.AddGroupingPolicy(userID, role)
    if err != nil {
        return fmt.Errorf("failed to assign role: %w", err)
    }
    
    if added {
        rm.logger.Info("Role assigned",
            slog.String("user_id", userID),
            slog.String("role", role),
        )
    }
    
    return nil
}

// 移除用户角色
func (rm *RoleManager) RemoveRole(userID, role string) error {
    removed, err := rm.enforcer.RemoveGroupingPolicy(userID, role)
    if err != nil {
        return fmt.Errorf("failed to remove role: %w", err)
    }
    
    if removed {
        rm.logger.Info("Role removed",
            slog.String("user_id", userID),
            slog.String("role", role),
        )
    }
    
    return nil
}

// 获取用户角色
func (rm *RoleManager) GetUserRoles(userID string) ([]string, error) {
    roles, err := rm.enforcer.GetRolesForUser(userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get user roles: %w", err)
    }
    return roles, nil
}

// 检查用户是否有特定角色
func (rm *RoleManager) HasRole(userID, role string) (bool, error) {
    roles, err := rm.GetUserRoles(userID)
    if err != nil {
        return false, err
    }
    
    for _, r := range roles {
        if r == role {
            return true, nil
        }
    }
    return false, nil
}
```

### 权限策略管理
```go
type PolicyManager struct {
    enforcer *casbin.Enforcer
    logger   *slog.Logger
}

// 添加权限策略
func (pm *PolicyManager) AddPolicy(subject, object, action string) error {
    added, err := pm.enforcer.AddPolicy(subject, object, action)
    if err != nil {
        return fmt.Errorf("failed to add policy: %w", err)
    }
    
    if added {
        pm.logger.Info("Policy added",
            slog.String("subject", subject),
            slog.String("object", object),
            slog.String("action", action),
        )
    }
    
    return nil
}

// 移除权限策略
func (pm *PolicyManager) RemovePolicy(subject, object, action string) error {
    removed, err := pm.enforcer.RemovePolicy(subject, object, action)
    if err != nil {
        return fmt.Errorf("failed to remove policy: %w", err)
    }
    
    if removed {
        pm.logger.Info("Policy removed",
            slog.String("subject", subject),
            slog.String("object", object),
            slog.String("action", action),
        )
    }
    
    return nil
}

// 获取角色的所有权限
func (pm *PolicyManager) GetRolePolicies(role string) ([][]string, error) {
    policies := pm.enforcer.GetFilteredPolicy(0, role)
    return policies, nil
}
```

## API接口

### 角色管理API
```go
// 获取用户角色
func (h *AuthHandler) GetUserRoles(c *gin.Context) {
    userID := c.Param("userId")
    
    roles, err := h.roleManager.GetUserRoles(userID)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"roles": roles})
}

// 分配角色
func (h *AuthHandler) AssignRole(c *gin.Context) {
    var req struct {
        UserID string `json:"user_id" binding:"required"`
        Role   string `json:"role" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    if err := h.roleManager.AssignRole(req.UserID, req.Role); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"message": "role assigned successfully"})
}

// 移除角色
func (h *AuthHandler) RemoveRole(c *gin.Context) {
    var req struct {
        UserID string `json:"user_id" binding:"required"`
        Role   string `json:"role" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    if err := h.roleManager.RemoveRole(req.UserID, req.Role); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"message": "role removed successfully"})
}
```

### 权限检查API
```go
// 检查权限
func (h *AuthHandler) CheckPermission(c *gin.Context) {
    var req struct {
        Subject string `json:"subject" binding:"required"`
        Object  string `json:"object" binding:"required"`
        Action  string `json:"action" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    allowed, err := h.enforcer.Enforce(req.Subject, req.Object, req.Action)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"allowed": allowed})
}
```

## 资源权限映射

### Maven仓库权限
```
repository:maven:central     -> 中央仓库代理
repository:maven:releases    -> 发布版本仓库
repository:maven:snapshots   -> 快照版本仓库
repository:maven:*           -> 所有Maven仓库
```

### NPM仓库权限
```
repository:npm:proxy         -> NPM代理仓库
repository:npm:hosted        -> NPM托管仓库
repository:npm:group         -> NPM组合仓库
repository:npm:*             -> 所有NPM仓库
```

### 系统管理权限
```
agents:package_engine        -> 包处理引擎
agents:security_agent        -> 安全扫描代理
agents:*                     -> 所有代理
users:*                      -> 用户管理
system:*                     -> 系统管理
```

## 权限配置示例

### 配置文件
```yaml
auth:
  casbin:
    model_file: "./config/rbac_model.conf"
    policy_db: "sqlite3://nexus.db"
    
    # 默认角色权限配置
    default_policies:
      admin:
        - "repository:maven:*:read"
        - "repository:maven:*:write"
        - "repository:maven:*:delete"
        - "repository:npm:*:read"
        - "repository:npm:*:write"
        - "repository:npm:*:delete"
        - "agents:*:manage"
        - "users:*:manage"
        - "system:*:manage"
      
      developer:
        - "repository:maven:*:read"
        - "repository:maven:releases:write"
        - "repository:maven:snapshots:write"
        - "repository:npm:*:read"
        - "repository:npm:hosted:write"
      
      readonly:
        - "repository:maven:*:read"
        - "repository:npm:*:read"
```

## 动态权限更新

### 权限缓存和更新
```go
func (rm *RoleManager) RefreshPolicies() error {
    // 重新加载策略
    if err := rm.enforcer.LoadPolicy(); err != nil {
        return fmt.Errorf("failed to reload policies: %w", err)
    }
    
    rm.logger.Info("Policies refreshed successfully")
    return nil
}

// 定期刷新权限
func (rm *RoleManager) StartPolicyRefresh(interval time.Duration) {
    ticker := time.NewTicker(interval)
    go func() {
        for range ticker.C {
            if err := rm.RefreshPolicies(); err != nil {
                rm.logger.Error("Failed to refresh policies", slog.Any("error", err))
            }
        }
    }()
}
```

## 最佳实践

1. **最小权限原则**: 只给用户必需的最小权限
2. **角色分离**: 合理设计角色层次，避免权限过于集中
3. **审计日志**: 记录所有权限变更操作
4. **定期审查**: 定期审查用户权限和角色分配
5. **动态更新**: 支持权限的动态更新，无需重启服务
