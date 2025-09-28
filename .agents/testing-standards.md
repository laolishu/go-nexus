# 测试规范

## 测试策略

### 测试金字塔
- **单元测试**: 70% - 测试单个函数/方法
- **集成测试**: 20% - 测试组件间交互
- **端到端测试**: 10% - 测试完整业务流程

### 测试覆盖率要求
- 核心业务逻辑: ≥ 90%
- 整体代码覆盖率: ≥ 80%
- 关键路径覆盖率: 100%

## 单元测试规范

### 测试文件组织
```
internal/
├── repository/
│   ├── manager.go
│   ├── manager_test.go          # 单元测试
│   ├── validator.go
│   └── validator_test.go
└── storage/
    ├── filesystem.go
    ├── filesystem_test.go
    ├── s3.go
    └── s3_test.go
```

### 测试命名规范
```go
// 函数测试: Test<FunctionName>
func TestCreateRepository(t *testing.T) {}
func TestValidateConfig(t *testing.T) {}

// 方法测试: Test<Type>_<Method>
func TestRepositoryManager_CreateRepository(t *testing.T) {}
func TestFileSystemStorage_Upload(t *testing.T) {}

// 表格驱动测试
func TestRepositoryManager_CreateRepository(t *testing.T) {
    tests := []struct {
        name    string
        input   *CreateRepositoryRequest
        want    *Repository
        wantErr bool
        setup   func(*testing.T) *RepositoryManager
    }{
        {
            name: "successful_creation_maven_proxy",
            input: &CreateRepositoryRequest{
                Name:   "maven-central",
                Type:   "proxy",
                Format: "maven",
                URL:    "https://repo1.maven.org/maven2/",
            },
            wantErr: false,
            setup:   setupValidManager,
        },
        {
            name: "error_duplicate_name",
            input: &CreateRepositoryRequest{
                Name: "existing-repo",
                Type: "proxy",
                Format: "maven",
            },
            wantErr: true,
            setup:   setupManagerWithExistingRepo,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            manager := tt.setup(t)
            got, err := manager.CreateRepository(context.Background(), tt.input)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateRepository() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !tt.wantErr {
                assert.Equal(t, tt.input.Name, got.Name)
                assert.Equal(t, tt.input.Type, got.Type)
            }
        })
    }
}
```

### Mock和Stub规范
```go
// 使用testify/mock
type MockStorageProvider struct {
    mock.Mock
}

func (m *MockStorageProvider) Upload(ctx context.Context, path string, data []byte) error {
    args := m.Called(ctx, path, data)
    return args.Error(0)
}

func (m *MockStorageProvider) Download(ctx context.Context, path string) ([]byte, error) {
    args := m.Called(ctx, path)
    return args.Get(0).([]byte), args.Error(1)
}

// 在测试中使用Mock
func TestRepositoryManager_UploadArtifact(t *testing.T) {
    mockStorage := new(MockStorageProvider)
    manager := &RepositoryManager{
        storage: mockStorage,
    }
    
    // 设置Mock期望
    mockStorage.On("Upload", mock.Anything, "test/path", mock.Anything).Return(nil)
    
    // 执行测试
    err := manager.UploadArtifact(context.Background(), "test/path", []byte("data"))
    
    // 验证结果
    assert.NoError(t, err)
    mockStorage.AssertExpectations(t)
}
```

### 测试辅助函数
```go
// test_helpers.go
func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    require.NoError(t, err)
    
    // 执行数据库迁移
    err = runMigrations(db)
    require.NoError(t, err)
    
    return db
}

func createTestRepository(t *testing.T, name string) *Repository {
    return &Repository{
        ID:     uuid.New().String(),
        Name:   name,
        Type:   "hosted",
        Format: "maven",
        CreatedAt: time.Now(),
    }
}

func assertRepositoryEqual(t *testing.T, expected, actual *Repository) {
    assert.Equal(t, expected.Name, actual.Name)
    assert.Equal(t, expected.Type, actual.Type)
    assert.Equal(t, expected.Format, actual.Format)
}
```

## 集成测试规范

### 测试环境
```go
//go:build integration
// +build integration

// integration_test.go
func TestRepositoryIntegration(t *testing.T) {
    // 启动测试依赖服务
    testContainer := setupTestContainer(t)
    defer testContainer.Cleanup()
    
    // 初始化应用
    app, cleanup := setupTestApp(t, testContainer)
    defer cleanup()
    
    // 执行集成测试
    t.Run("create_and_query_repository", func(t *testing.T) {
        testCreateAndQueryRepository(t, app)
    })
    
    t.Run("upload_and_download_artifact", func(t *testing.T) {
        testUploadAndDownloadArtifact(t, app)
    })
}

func setupTestContainer(t *testing.T) *TestContainer {
    // 使用testcontainers启动PostgreSQL
    ctx := context.Background()
    req := testcontainers.ContainerRequest{
        Image:        "postgres:13",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_DB":       "go_nexus_test",
            "POSTGRES_USER":     "test",
            "POSTGRES_PASSWORD": "test",
        },
        WaitingFor: wait.ForLog("database system is ready to accept connections"),
    }
    
    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    require.NoError(t, err)
    
    return &TestContainer{Container: container}
}
```

### 数据库测试
```go
func TestRepositoryDatabase(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := &Repository{
        Name:   "test-repo",
        Type:   "hosted",
        Format: "maven",
    }
    
    // 测试创建
    err := createRepository(db, repo)
    assert.NoError(t, err)
    
    // 测试查询
    found, err := getRepository(db, repo.Name)
    assert.NoError(t, err)
    assert.Equal(t, repo.Name, found.Name)
    
    // 测试更新
    repo.Description = "Updated description"
    err = updateRepository(db, repo)
    assert.NoError(t, err)
    
    // 测试删除
    err = deleteRepository(db, repo.Name)
    assert.NoError(t, err)
}
```

### HTTP API测试
```go
func TestRepositoryAPI(t *testing.T) {
    app := setupTestApp(t)
    server := httptest.NewServer(app.Router)
    defer server.Close()
    
    client := &http.Client{Timeout: 10 * time.Second}
    baseURL := server.URL + "/api/v1"
    
    t.Run("create_repository", func(t *testing.T) {
        payload := map[string]interface{}{
            "name":   "test-repo",
            "type":   "hosted",
            "format": "maven",
        }
        
        resp, err := postJSON(client, baseURL+"/repositories", payload)
        require.NoError(t, err)
        defer resp.Body.Close()
        
        assert.Equal(t, http.StatusCreated, resp.StatusCode)
        
        var result map[string]interface{}
        err = json.NewDecoder(resp.Body).Decode(&result)
        require.NoError(t, err)
        
        assert.True(t, result["success"].(bool))
    })
}

func postJSON(client *http.Client, url string, payload interface{}) (*http.Response, error) {
    data, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }
    
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    return client.Do(req)
}
```

## 性能测试规范

### 基准测试
```go
func BenchmarkRepositoryManager_CreateRepository(b *testing.B) {
    manager := setupBenchmarkManager(b)
    config := &RepositoryConfig{
        Name:   "bench-repo",
        Type:   "hosted",
        Format: "maven",
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        config.Name = fmt.Sprintf("bench-repo-%d", i)
        _, err := manager.CreateRepository(context.Background(), config)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkArtifactUpload(b *testing.B) {
    manager := setupBenchmarkManager(b)
    data := make([]byte, 1024*1024) // 1MB
    
    b.ResetTimer()
    b.SetBytes(int64(len(data)))
    
    for i := 0; i < b.N; i++ {
        path := fmt.Sprintf("test/artifact-%d.jar", i)
        err := manager.UploadArtifact(context.Background(), path, data)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### 负载测试
```go
func TestConcurrentRepositoryOperations(t *testing.T) {
    manager := setupTestManager(t)
    concurrency := 100
    operations := 1000
    
    var wg sync.WaitGroup
    errors := make(chan error, operations)
    
    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            
            for j := 0; j < operations/concurrency; j++ {
                config := &RepositoryConfig{
                    Name:   fmt.Sprintf("repo-%d-%d", workerID, j),
                    Type:   "hosted",
                    Format: "maven",
                }
                
                _, err := manager.CreateRepository(context.Background(), config)
                if err != nil {
                    errors <- err
                    return
                }
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    errorCount := 0
    for err := range errors {
        t.Logf("Error: %v", err)
        errorCount++
    }
    
    if errorCount > 0 {
        t.Errorf("Got %d errors out of %d operations", errorCount, operations)
    }
}
```

## 测试工具和命令

### Makefile测试命令
```makefile
# 运行所有测试
test:
	go test ./... -v

# 运行单元测试
test-unit:
	go test ./... -v -short

# 运行集成测试
test-integration:
	go test ./... -v -tags=integration

# 生成覆盖率报告
coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# 运行基准测试
benchmark:
	go test ./... -bench=. -benchmem

# 竞态检测
race:
	go test ./... -race
```

### 持续集成配置
```yaml
# .github/workflows/test.yml
name: Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Run unit tests
      run: make test-unit
    
    - name: Run integration tests
      run: make test-integration
    
    - name: Generate coverage
      run: make coverage
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```