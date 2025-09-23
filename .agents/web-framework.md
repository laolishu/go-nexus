# Web框架层 (Gin)

## 概述

go-nexus使用Gin作为Web框架，提供高性能的HTTP服务和RESTful API。

## 核心架构

### 主应用结构
```go
type NexusServer struct {
    router      *gin.Engine
    logger      *slog.Logger
    config      *viper.Viper
    enforcer    *casbin.Enforcer
    storage     storage.Backend
    agents      *AgentManager
}
```

## 路由配置

### 路由设置
```go
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
```

## 中间件

### 日志中间件
```go
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

### JWT认证中间件
```go
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

### 速率限制中间件
```go
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

### 安全头中间件
```go
func securityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Next()
    }
}
```

### CORS配置
```go
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
```

## 响应标准化

### API响应格式
```go
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
```

## 性能优化

### Gin优化配置
```go
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
```

### 服务器配置
```go
type ServerConfig struct {
    ReadTimeout    time.Duration `mapstructure:"read_timeout"`
    WriteTimeout   time.Duration `mapstructure:"write_timeout"`
    IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
    MaxHeaderBytes int           `mapstructure:"max_header_bytes"`
}
```

## 前端路由配置

### 路由结构
```go
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

## 最佳实践

1. **中间件顺序**: 日志 -> 恢复 -> CORS -> 认证 -> 权限检查
2. **错误处理**: 统一使用panic recovery中间件处理异常
3. **请求验证**: 使用gin的binding功能进行参数验证
4. **性能监控**: 集成Prometheus指标收集
5. **安全考虑**: 始终设置安全头，启用HTTPS
