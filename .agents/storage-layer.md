# 存储层设计

## 概述

go-nexus支持多种存储后端，包括本地文件系统和S3兼容存储，提供统一的存储抽象接口。

## 存储接口定义

### 核心接口
```go
type Backend interface {
    Store(ctx context.Context, path string, data io.Reader) error
    Retrieve(ctx context.Context, path string) (io.ReadCloser, error)
    Delete(ctx context.Context, path string) error
    Exists(ctx context.Context, path string) (bool, error)
    List(ctx context.Context, prefix string) ([]string, error)
    Stat(ctx context.Context, path string) (*ObjectInfo, error)
}

type ObjectInfo struct {
    Path         string
    Size         int64
    LastModified time.Time
    ETag         string
    ContentType  string
}
```

### 高级接口(可选)
```go
type AdvancedBackend interface {
    Backend
    
    // 批量操作
    StoreBatch(ctx context.Context, items []BatchItem) error
    DeleteBatch(ctx context.Context, paths []string) error
    
    // 流式操作
    StreamStore(ctx context.Context, path string) (io.WriteCloser, error)
    StreamRetrieve(ctx context.Context, path string) (io.ReadCloser, error)
    
    // 元数据操作
    SetMetadata(ctx context.Context, path string, metadata map[string]string) error
    GetMetadata(ctx context.Context, path string) (map[string]string, error)
}

type BatchItem struct {
    Path string
    Data io.Reader
    Size int64
}
```

## 本地文件系统存储

### 实现结构
```go
type LocalStorage struct {
    basePath string
    logger   *slog.Logger
    mu       sync.RWMutex
}

func NewLocalStorage(basePath string, logger *slog.Logger) (*LocalStorage, error) {
    // 确保基础路径存在
    if err := os.MkdirAll(basePath, 0755); err != nil {
        return nil, fmt.Errorf("failed to create base path: %w", err)
    }
    
    return &LocalStorage{
        basePath: basePath,
        logger:   logger,
    }, nil
}
```

### 核心方法实现
```go
func (s *LocalStorage) Store(ctx context.Context, path string, data io.Reader) error {
    fullPath := filepath.Join(s.basePath, path)
    
    // 创建目录
    if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
        return fmt.Errorf("failed to create directory: %w", err)
    }
    
    // 创建临时文件
    tempFile := fullPath + ".tmp"
    file, err := os.Create(tempFile)
    if err != nil {
        return fmt.Errorf("failed to create temp file: %w", err)
    }
    defer file.Close()
    
    // 写入数据
    size, err := io.Copy(file, data)
    if err != nil {
        os.Remove(tempFile)
        return fmt.Errorf("failed to write file: %w", err)
    }
    
    // 原子性重命名
    if err := os.Rename(tempFile, fullPath); err != nil {
        os.Remove(tempFile)
        return fmt.Errorf("failed to rename file: %w", err)
    }
    
    s.logger.Info("File stored",
        slog.String("path", path),
        slog.Int64("size", size),
    )
    
    return nil
}

func (s *LocalStorage) Retrieve(ctx context.Context, path string) (io.ReadCloser, error) {
    fullPath := filepath.Join(s.basePath, path)
    
    file, err := os.Open(fullPath)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("failed to open file: %w", err)
    }
    
    return file, nil
}

func (s *LocalStorage) Delete(ctx context.Context, path string) error {
    fullPath := filepath.Join(s.basePath, path)
    
    if err := os.Remove(fullPath); err != nil {
        if os.IsNotExist(err) {
            return ErrNotFound
        }
        return fmt.Errorf("failed to delete file: %w", err)
    }
    
    s.logger.Info("File deleted", slog.String("path", path))
    return nil
}

func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
    fullPath := filepath.Join(s.basePath, path)
    
    _, err := os.Stat(fullPath)
    if err != nil {
        if os.IsNotExist(err) {
            return false, nil
        }
        return false, fmt.Errorf("failed to stat file: %w", err)
    }
    
    return true, nil
}

func (s *LocalStorage) Stat(ctx context.Context, path string) (*ObjectInfo, error) {
    fullPath := filepath.Join(s.basePath, path)
    
    info, err := os.Stat(fullPath)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("failed to stat file: %w", err)
    }
    
    // 计算ETag (文件的MD5哈希)
    etag, err := s.calculateETag(fullPath)
    if err != nil {
        s.logger.Warn("Failed to calculate ETag", slog.String("path", path))
        etag = ""
    }
    
    return &ObjectInfo{
        Path:         path,
        Size:         info.Size(),
        LastModified: info.ModTime(),
        ETag:         etag,
        ContentType:  s.detectContentType(fullPath),
    }, nil
}
```

### 辅助方法
```go
func (s *LocalStorage) calculateETag(filePath string) (string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer file.Close()
    
    hash := md5.New()
    if _, err := io.Copy(hash, file); err != nil {
        return "", err
    }
    
    return hex.EncodeToString(hash.Sum(nil)), nil
}

func (s *LocalStorage) detectContentType(filePath string) string {
    ext := filepath.Ext(filePath)
    switch ext {
    case ".jar", ".war":
        return "application/java-archive"
    case ".pom":
        return "application/xml"
    case ".tgz", ".tar.gz":
        return "application/gzip"
    case ".json":
        return "application/json"
    default:
        return "application/octet-stream"
    }
}
```

## S3兼容存储

### 实现结构
```go
type S3Storage struct {
    client *s3.Client
    bucket string
    logger *slog.Logger
}

func NewS3Storage(config S3Config, logger *slog.Logger) (*S3Storage, error) {
    cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
        awsconfig.WithRegion(config.Region),
        awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
            config.AccessKey, config.SecretKey, "",
        )),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }
    
    // 自定义端点(用于MinIO等S3兼容存储)
    if config.Endpoint != "" {
        cfg.BaseEndpoint = aws.String(config.Endpoint)
    }
    
    // 强制路径样式(MinIO需要)
    if config.ForcePathStyle {
        cfg.S3ForcePathStyle = true
    }
    
    client := s3.NewFromConfig(cfg)
    
    return &S3Storage{
        client: client,
        bucket: config.Bucket,
        logger: logger,
    }, nil
}
```

### 核心方法实现
```go
func (s *S3Storage) Store(ctx context.Context, path string, data io.Reader) error {
    // 如果数据是文件，获取大小用于进度跟踪
    var contentLength *int64
    if seeker, ok := data.(io.Seeker); ok {
        if size, err := seeker.Seek(0, io.SeekEnd); err == nil {
            contentLength = &size
            seeker.Seek(0, io.SeekStart)
        }
    }
    
    _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
        Bucket:        aws.String(s.bucket),
        Key:           aws.String(path),
        Body:          data,
        ContentLength: contentLength,
        ContentType:   aws.String(s.detectContentType(path)),
    })
    
    if err != nil {
        return fmt.Errorf("failed to upload to S3: %w", err)
    }
    
    s.logger.Info("Object stored to S3",
        slog.String("bucket", s.bucket),
        slog.String("key", path),
    )
    
    return nil
}

func (s *S3Storage) Retrieve(ctx context.Context, path string) (io.ReadCloser, error) {
    result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(path),
    })
    
    if err != nil {
        var noSuchKey *types.NoSuchKey
        if errors.As(err, &noSuchKey) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("failed to get object from S3: %w", err)
    }
    
    return result.Body, nil
}

func (s *S3Storage) Delete(ctx context.Context, path string) error {
    _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(path),
    })
    
    if err != nil {
        return fmt.Errorf("failed to delete object from S3: %w", err)
    }
    
    s.logger.Info("Object deleted from S3",
        slog.String("bucket", s.bucket),
        slog.String("key", path),
    )
    
    return nil
}

func (s *S3Storage) Exists(ctx context.Context, path string) (bool, error) {
    _, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(path),
    })
    
    if err != nil {
        var noSuchKey *types.NoSuchKey
        if errors.As(err, &noSuchKey) {
            return false, nil
        }
        return false, fmt.Errorf("failed to check object existence: %w", err)
    }
    
    return true, nil
}

func (s *S3Storage) Stat(ctx context.Context, path string) (*ObjectInfo, error) {
    result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(path),
    })
    
    if err != nil {
        var noSuchKey *types.NoSuchKey
        if errors.As(err, &noSuchKey) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("failed to get object info: %w", err)
    }
    
    var etag string
    if result.ETag != nil {
        etag = strings.Trim(*result.ETag, "\"")
    }
    
    var contentType string
    if result.ContentType != nil {
        contentType = *result.ContentType
    }
    
    return &ObjectInfo{
        Path:         path,
        Size:         result.ContentLength,
        LastModified: *result.LastModified,
        ETag:         etag,
        ContentType:  contentType,
    }, nil
}
```

## 存储工厂

### 工厂函数
```go
func NewStorageBackend(config StorageConfig, logger *slog.Logger) (Backend, error) {
    switch strings.ToLower(config.Backend) {
    case "filesystem", "local":
        return NewLocalStorage(config.Local.Path, logger)
    case "s3":
        return NewS3Storage(config.S3, logger)
    default:
        return nil, fmt.Errorf("unsupported storage backend: %s", config.Backend)
    }
}
```

### 配置结构
```go
type StorageConfig struct {
    Backend string      `mapstructure:"backend"`
    Local   LocalConfig `mapstructure:"local"`
    S3      S3Config    `mapstructure:"s3"`
}

type LocalConfig struct {
    Path string `mapstructure:"path"`
}

type S3Config struct {
    Endpoint        string `mapstructure:"endpoint"`
    Region          string `mapstructure:"region"`
    Bucket          string `mapstructure:"bucket"`
    AccessKey       string `mapstructure:"access_key"`
    SecretKey       string `mapstructure:"secret_key"`
    ForcePathStyle  bool   `mapstructure:"force_path_style"`
}
```

## 存储路径规范

### Maven包路径
```
repository/maven/{repository}/{groupId}/{artifactId}/{version}/{filename}

示例:
repository/maven/releases/com/example/myapp/1.0.0/myapp-1.0.0.jar
repository/maven/releases/com/example/myapp/1.0.0/myapp-1.0.0.pom
repository/maven/releases/com/example/myapp/maven-metadata.xml
```

### NPM包路径
```
repository/npm/{repository}/{scope}/{package}/{version}/{filename}

示例:
repository/npm/hosted/@company/utils/1.0.0/utils-1.0.0.tgz
repository/npm/hosted/@company/utils/package.json
repository/npm/hosted/lodash/4.17.21/lodash-4.17.21.tgz
```

## 性能优化

### 并发控制
```go
type ConcurrentStorage struct {
    backend    Backend
    semaphore  chan struct{}
    logger     *slog.Logger
}

func NewConcurrentStorage(backend Backend, maxConcurrency int, logger *slog.Logger) *ConcurrentStorage {
    return &ConcurrentStorage{
        backend:   backend,
        semaphore: make(chan struct{}, maxConcurrency),
        logger:    logger,
    }
}

func (cs *ConcurrentStorage) Store(ctx context.Context, path string, data io.Reader) error {
    cs.semaphore <- struct{}{}
    defer func() { <-cs.semaphore }()
    
    return cs.backend.Store(ctx, path, data)
}
```

### 缓存层
```go
type CachedStorage struct {
    backend Backend
    cache   *lru.Cache
    logger  *slog.Logger
}

func NewCachedStorage(backend Backend, cacheSize int, logger *slog.Logger) (*CachedStorage, error) {
    cache, err := lru.New(cacheSize)
    if err != nil {
        return nil, err
    }
    
    return &CachedStorage{
        backend: backend,
        cache:   cache,
        logger:  logger,
    }, nil
}

func (cs *CachedStorage) Retrieve(ctx context.Context, path string) (io.ReadCloser, error) {
    // 检查缓存
    if cached, ok := cs.cache.Get(path); ok {
        cs.logger.Debug("Cache hit", slog.String("path", path))
        data := cached.([]byte)
        return io.NopCloser(bytes.NewReader(data)), nil
    }
    
    // 从后端获取
    reader, err := cs.backend.Retrieve(ctx, path)
    if err != nil {
        return nil, err
    }
    
    // 读取数据并缓存
    data, err := io.ReadAll(reader)
    reader.Close()
    if err != nil {
        return nil, err
    }
    
    cs.cache.Add(path, data)
    cs.logger.Debug("Cache miss", slog.String("path", path))
    
    return io.NopCloser(bytes.NewReader(data)), nil
}
```

## 错误处理

### 错误类型定义
```go
var (
    ErrNotFound      = errors.New("object not found")
    ErrAlreadyExists = errors.New("object already exists")
    ErrInvalidPath   = errors.New("invalid path")
    ErrStorageFull   = errors.New("storage full")
    ErrPermissionDenied = errors.New("permission denied")
)

type StorageError struct {
    Op   string
    Path string
    Err  error
}

func (e *StorageError) Error() string {
    return fmt.Sprintf("storage %s %s: %v", e.Op, e.Path, e.Err)
}

func (e *StorageError) Unwrap() error {
    return e.Err
}
```

## 监控和指标

### Prometheus指标
```go
var (
    storageOperationsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "storage_operations_total",
            Help: "Total number of storage operations",
        },
        []string{"backend", "operation", "status"},
    )
    
    storageOperationDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "storage_operation_duration_seconds",
            Help: "Duration of storage operations",
        },
        []string{"backend", "operation"},
    )
    
    storageUsageBytes = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "storage_usage_bytes",
            Help: "Storage usage in bytes",
        },
        []string{"backend", "repository"},
    )
)
```

### 带监控的存储包装器
```go
type MonitoredStorage struct {
    backend Backend
    name    string
    logger  *slog.Logger
}

func (ms *MonitoredStorage) Store(ctx context.Context, path string, data io.Reader) error {
    start := time.Now()
    err := ms.backend.Store(ctx, path, data)
    duration := time.Since(start)
    
    status := "success"
    if err != nil {
        status = "error"
    }
    
    storageOperationsTotal.WithLabelValues(ms.name, "store", status).Inc()
    storageOperationDuration.WithLabelValues(ms.name, "store").Observe(duration.Seconds())
    
    ms.logger.Info("Storage operation",
        slog.String("operation", "store"),
        slog.String("path", path),
        slog.Duration("duration", duration),
        slog.String("status", status),
    )
    
    return err
}
```

## 最佳实践

1. **接口设计**: 使用接口抽象不同的存储后端
2. **错误处理**: 统一错误类型，便于调用方处理
3. **并发安全**: 确保存储操作的并发安全性
4. **性能监控**: 监控存储操作的性能指标
5. **路径规范**: 统一存储路径命名规范
6. **数据完整性**: 使用校验和验证数据完整性
