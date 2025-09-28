# 代码开发规范

## Go代码风格规范

### 基本规范
- 遵循 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- 使用 `gofmt` 格式化代码
- 使用 `golint` 检查代码风格
- 使用 `go vet` 检查潜在问题
- 使用 `golangci-lint` 进行综合检查

### 命名规范
```go
// 包命名: 小写，简短，有意义
package repository

// 接口命名: 以 "er" 结尾
type RepositoryManager interface {}
type StorageProvider interface {}

// 结构体命名: 大驼峰
type RepositoryConfig struct {}

// 函数命名: 大驼峰（公开）、小驼峰（私有）
func CreateRepository() {}
func validateConfig() {}

// 常量命名: 大驼峰或全大写
const DefaultTimeout = 30 * time.Second
const (
    REPOSITORY_TYPE_PROXY = "proxy"
    REPOSITORY_TYPE_HOSTED = "hosted"
)

// 变量命名: 小驼峰
var defaultConfig Config
var repositoryCache map[string]*Repository
```

### 注释规范
```go
// Package repository 提供仓库管理功能
//
// 本包实现了多种类型仓库的创建、配置和管理，
// 支持Maven、npm等格式的依赖管理。
package repository

// RepositoryManager 仓库管理器接口
//
// 提供仓库的完整生命周期管理功能，包括：
// - 创建和删除仓库
// - 配置仓库参数
// - 管理仓库权限
type RepositoryManager interface {
    // CreateRepository 创建新仓库
    //
    // 参数：
    //   config: 仓库配置信息
    //
    // 返回：
    //   *Repository: 创建的仓库实例
    //   error: 创建过程中的错误
    CreateRepository(config *RepositoryConfig) (*Repository, error)
}
```

### 错误处理规范
```go
// 自定义错误类型
type RepositoryError struct {
    Code    string
    Message string
    Cause   error
}

func (e *RepositoryError) Error() string {
    return fmt.Sprintf("repository error [%s]: %s", e.Code, e.Message)
}

// 错误包装
func (r *repositoryManager) CreateRepository(config *RepositoryConfig) (*Repository, error) {
    if err := r.validateConfig(config); err != nil {
        return nil, fmt.Errorf("validate config failed: %w", err)
    }
    
    repo, err := r.storage.Create(config)
    if err != nil {
        return nil, &RepositoryError{
            Code:    "CREATE_FAILED",
            Message: "failed to create repository",
            Cause:   err,
        }
    }
    
    return repo, nil
}
```

## 项目结构规范（Wire DI）

### 依赖注入规范
```go
// wire.go - Wire依赖注入定义
//go:build wireinject
// +build wireinject

package main

import (
    "github.com/google/wire"
    "github.com/laolishu/go-nexus/internal/repository"
    "github.com/laolishu/go-nexus/internal/storage"
)

// 提供者集合
var repositoryProviderSet = wire.NewSet(
    repository.NewManager,
    repository.NewValidator,
)

var storageProviderSet = wire.NewSet(
    storage.NewFileSystemProvider,
    wire.Bind(new(storage.Provider), new(*storage.FileSystemProvider)),
)

// 应用程序Wire函数
func InitializeApp() (*App, error) {
    wire.Build(
        repositoryProviderSet,
        storageProviderSet,
        NewApp,
    )
    return nil, nil
}
```

### 接口定义规范
```go
// 接口应该在使用者包中定义，而不是实现者包中
// internal/service/repository.go
type RepositoryService interface {
    Create(ctx context.Context, req *CreateRepositoryRequest) (*Repository, error)
    Get(ctx context.Context, id string) (*Repository, error)
}

// internal/repository/manager.go 
type Manager struct {
    storage  storage.Provider
    cache    cache.Manager
    logger   *zap.Logger
}

// 实现接口
func (m *Manager) Create(ctx context.Context, req *CreateRepositoryRequest) (*Repository, error) {
    // 实现逻辑
}
```

### 配置注入规范
```go
// internal/config/config.go
type Config struct {
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database"`
    Storage  StorageConfig  `yaml:"storage"`
}

// wire_gen.go 生成的配置提供者
func provideConfig() (*Config, error) {
    return LoadConfig("config.yaml")
}

func provideServerConfig(cfg *Config) ServerConfig {
    return cfg.Server
}
```

## 测试规范

### 单元测试
```go
// repository_test.go
func TestRepositoryManager_CreateRepository(t *testing.T) {
    tests := []struct {
        name    string
        config  *RepositoryConfig
        want    *Repository
        wantErr bool
    }{
        {
            name: "valid maven repository",
            config: &RepositoryConfig{
                Name: "maven-public",
                Type: "proxy",
                Format: "maven",
            },
            wantErr: false,
        },
        {
            name: "invalid config - empty name",
            config: &RepositoryConfig{
                Name: "",
                Type: "proxy",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            manager := setupTestManager(t)
            got, err := manager.CreateRepository(tt.config)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateRepository() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !tt.wantErr && got == nil {
                t.Error("CreateRepository() returned nil repository")
            }
        })
    }
}

func setupTestManager(t *testing.T) *repositoryManager {
    // 使用Wire进行测试依赖注入
    return &repositoryManager{
        storage: &mockStorage{},
        logger:  zap.NewNop(),
    }
}
```

### 集成测试
```go
// integration_test.go
//go:build integration
// +build integration

func TestRepositoryIntegration(t *testing.T) {
    // 启动测试数据库
    db := setupTestDB(t)
    defer db.Close()
    
    // 使用真实依赖
    app, err := InitializeTestApp(db)
    require.NoError(t, err)
    
    // 执行集成测试
    ctx := context.Background()
    repo, err := app.RepositoryService.Create(ctx, &CreateRepositoryRequest{
        Name: "test-repo",
        Type: "hosted",
    })
    
    assert.NoError(t, err)
    assert.Equal(t, "test-repo", repo.Name)
}
```

## 代码组织规范

### 包导入顺序
```go
import (
    // 标准库
    "context"
    "fmt"
    "net/http"
    
    // 第三方库
    "github.com/gin-gonic/gin"
    "github.com/google/wire"
    "go.uber.org/zap"
    
    // 本项目包
    "github.com/laolishu/go-nexus/internal/config"
    "github.com/laolishu/go-nexus/internal/repository"
)
```

### 文件组织
- 每个文件聚焦单一职责
- 接口定义和实现分离
- 测试文件与源文件同目录
- Wire文件使用build tag隔离