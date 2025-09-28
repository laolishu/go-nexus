# 插件开发规范

## 插件架构设计

### 插件接口定义
```go
// pkg/plugin/interface.go
package plugin

import "context"

// Plugin 插件基础接口
type Plugin interface {
    // Name 返回插件名称
    Name() string
    
    // Version 返回插件版本
    Version() string
    
    // Initialize 初始化插件
    Initialize(ctx context.Context, config map[string]interface{}) error
    
    // Shutdown 关闭插件
    Shutdown(ctx context.Context) error
}

// FormatPlugin 格式处理插件接口
type FormatPlugin interface {
    Plugin
    
    // Format 返回支持的格式名称
    Format() string
    
    // ParseMetadata 解析artifact元数据
    ParseMetadata(ctx context.Context, data []byte) (*Metadata, error)
    
    // ValidatePath 验证路径格式
    ValidatePath(path string) error
    
    // GenerateMetadata 生成元数据文件
    GenerateMetadata(ctx context.Context, artifacts []*Artifact) ([]byte, error)
}

// StoragePlugin 存储插件接口
type StoragePlugin interface {
    Plugin
    
    // Upload 上传文件
    Upload(ctx context.Context, path string, data []byte) error
    
    // Download 下载文件
    Download(ctx context.Context, path string) ([]byte, error)
    
    // Delete 删除文件
    Delete(ctx context.Context, path string) error
    
    // List 列出文件
    List(ctx context.Context, prefix string) ([]string, error)
    
    // Exists 检查文件是否存在
    Exists(ctx context.Context, path string) (bool, error)
}

// IntegrationPlugin 集成插件接口
type IntegrationPlugin interface {
    Plugin
    
    // OnArtifactUploaded artifact上传后回调
    OnArtifactUploaded(ctx context.Context, event *ArtifactEvent) error
    
    // OnArtifactDeleted artifact删除后回调
    OnArtifactDeleted(ctx context.Context, event *ArtifactEvent) error
    
    // OnRepositoryCreated 仓库创建后回调
    OnRepositoryCreated(ctx context.Context, event *RepositoryEvent) error
}
```

### 插件元数据规范
```go
// plugin.yaml - 插件描述文件
name: maven-plugin
version: 1.0.0
type: format
description: Maven repository format support
author: go-nexus team
license: Apache-2.0
go_version: 1.21+
dependencies:
  - name: xml-parser
    version: ">=1.0.0"

entry_point: ./maven-plugin.so
config_schema:
  properties:
    checksum_policy:
      type: string
      enum: [strict, warn, ignore]
      default: strict
    metadata_update_policy:
      type: string
      enum: [always, daily, never]
      default: daily

permissions:
  - read_artifacts
  - write_metadata
  - network_access

compatibility:
  go_nexus_version: ">=0.1.0"
```

## 格式插件开发

### Maven插件示例
```go
// plugins/maven/plugin.go
package maven

import (
    "context"
    "encoding/xml"
    "fmt"
    "path/filepath"
    "regexp"
    "strings"
    
    "github.com/laolishu/go-nexus/pkg/plugin"
)

type MavenPlugin struct {
    config *MavenConfig
}

type MavenConfig struct {
    ChecksumPolicy        string `json:"checksum_policy"`
    MetadataUpdatePolicy  string `json:"metadata_update_policy"`
}

// 实现Plugin接口
func (p *MavenPlugin) Name() string {
    return "maven-plugin"
}

func (p *MavenPlugin) Version() string {
    return "1.0.0"
}

func (p *MavenPlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
    // 解析配置
    p.config = &MavenConfig{}
    return mapstructure.Decode(config, p.config)
}

func (p *MavenPlugin) Shutdown(ctx context.Context) error {
    return nil
}

// 实现FormatPlugin接口
func (p *MavenPlugin) Format() string {
    return "maven"
}

func (p *MavenPlugin) ValidatePath(path string) error {
    // Maven路径格式: /groupId/artifactId/version/artifactId-version[-classifier].extension
    pattern := `^/?([^/]+/)*[^/]+-[^/]+(\.[^/]+)?$`
    matched, err := regexp.MatchString(pattern, path)
    if err != nil {
        return err
    }
    if !matched {
        return fmt.Errorf("invalid maven path format: %s", path)
    }
    return nil
}

func (p *MavenPlugin) ParseMetadata(ctx context.Context, data []byte) (*plugin.Metadata, error) {
    if filepath.Ext(string(data)) == ".pom" {
        return p.parsePomMetadata(data)
    }
    return p.parseArtifactMetadata(data)
}

func (p *MavenPlugin) parsePomMetadata(data []byte) (*plugin.Metadata, error) {
    var pom MavenPom
    if err := xml.Unmarshal(data, &pom); err != nil {
        return nil, fmt.Errorf("failed to parse POM: %w", err)
    }
    
    return &plugin.Metadata{
        GroupId:    pom.GroupId,
        ArtifactId: pom.ArtifactId,
        Version:    pom.Version,
        Packaging:  pom.Packaging,
        Dependencies: pom.Dependencies,
    }, nil
}

// POM文件结构
type MavenPom struct {
    XMLName      xml.Name `xml:"project"`
    GroupId      string   `xml:"groupId"`
    ArtifactId   string   `xml:"artifactId"`
    Version      string   `xml:"version"`
    Packaging    string   `xml:"packaging"`
    Dependencies []Dependency `xml:"dependencies>dependency"`
}

type Dependency struct {
    GroupId    string `xml:"groupId"`
    ArtifactId string `xml:"artifactId"`
    Version    string `xml:"version"`
    Scope      string `xml:"scope"`
}

// 导出插件符号
var Plugin MavenPlugin

func init() {
    // 插件初始化逻辑
}
```

### NPM插件示例
```go
// plugins/npm/plugin.go
package npm

import (
    "context"
    "encoding/json"
    "fmt"
    "regexp"
    
    "github.com/laolishu/go-nexus/pkg/plugin"
)

type NPMPlugin struct {
    config *NPMConfig
}

type NPMConfig struct {
    RegistryURL string `json:"registry_url"`
    AuthToken   string `json:"auth_token"`
}

func (p *NPMPlugin) Name() string {
    return "npm-plugin"
}

func (p *NPMPlugin) Version() string {
    return "1.0.0"
}

func (p *NPMPlugin) Format() string {
    return "npm"
}

func (p *NPMPlugin) ValidatePath(path string) error {
    // NPM路径格式验证
    // /@scope/package-name/-/package-name-version.tgz
    // /package-name/-/package-name-version.tgz
    patterns := []string{
        `^/@[^/]+/[^/]+/-/[^/]+-[^/]+\.tgz$`,
        `^/[^/]+/-/[^/]+-[^/]+\.tgz$`,
        `^/@[^/]+/[^/]+$`,
        `^/[^/]+$`,
    }
    
    for _, pattern := range patterns {
        matched, err := regexp.MatchString(pattern, path)
        if err != nil {
            return err
        }
        if matched {
            return nil
        }
    }
    
    return fmt.Errorf("invalid npm path format: %s", path)
}

func (p *NPMPlugin) ParseMetadata(ctx context.Context, data []byte) (*plugin.Metadata, error) {
    var pkg NPMPackage
    if err := json.Unmarshal(data, &pkg); err != nil {
        return nil, fmt.Errorf("failed to parse package.json: %w", err)
    }
    
    return &plugin.Metadata{
        Name:         pkg.Name,
        Version:      pkg.Version,
        Description:  pkg.Description,
        Dependencies: pkg.Dependencies,
        Keywords:     pkg.Keywords,
    }, nil
}

type NPMPackage struct {
    Name         string            `json:"name"`
    Version      string            `json:"version"`
    Description  string            `json:"description"`
    Keywords     []string          `json:"keywords"`
    Dependencies map[string]string `json:"dependencies"`
    DevDependencies map[string]string `json:"devDependencies"`
}

var Plugin NPMPlugin
```

## 存储插件开发

### S3存储插件示例
```go
// plugins/s3/plugin.go
package s3

import (
    "bytes"
    "context"
    "fmt"
    "io"
    
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    
    "github.com/laolishu/go-nexus/pkg/plugin"
)

type S3Plugin struct {
    client *s3.S3
    config *S3Config
}

type S3Config struct {
    Endpoint        string `json:"endpoint"`
    Region          string `json:"region"`
    Bucket          string `json:"bucket"`
    AccessKeyID     string `json:"access_key_id"`
    SecretAccessKey string `json:"secret_access_key"`
    UseSSL          bool   `json:"use_ssl"`
}

func (p *S3Plugin) Name() string {
    return "s3-storage-plugin"
}

func (p *S3Plugin) Version() string {
    return "1.0.0"
}

func (p *S3Plugin) Initialize(ctx context.Context, config map[string]interface{}) error {
    p.config = &S3Config{}
    if err := mapstructure.Decode(config, p.config); err != nil {
        return err
    }
    
    // 创建S3客户端
    sess, err := session.NewSession(&aws.Config{
        Endpoint:         aws.String(p.config.Endpoint),
        Region:           aws.String(p.config.Region),
        Credentials:      credentials.NewStaticCredentials(p.config.AccessKeyID, p.config.SecretAccessKey, ""),
        S3ForcePathStyle: aws.Bool(true),
        DisableSSL:       aws.Bool(!p.config.UseSSL),
    })
    if err != nil {
        return fmt.Errorf("failed to create S3 session: %w", err)
    }
    
    p.client = s3.New(sess)
    return nil
}

func (p *S3Plugin) Upload(ctx context.Context, path string, data []byte) error {
    _, err := p.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
        Bucket: aws.String(p.config.Bucket),
        Key:    aws.String(path),
        Body:   bytes.NewReader(data),
    })
    return err
}

func (p *S3Plugin) Download(ctx context.Context, path string) ([]byte, error) {
    result, err := p.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
        Bucket: aws.String(p.config.Bucket),
        Key:    aws.String(path),
    })
    if err != nil {
        return nil, err
    }
    defer result.Body.Close()
    
    return io.ReadAll(result.Body)
}

func (p *S3Plugin) Delete(ctx context.Context, path string) error {
    _, err := p.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
        Bucket: aws.String(p.config.Bucket),
        Key:    aws.String(path),
    })
    return err
}

func (p *S3Plugin) List(ctx context.Context, prefix string) ([]string, error) {
    result, err := p.client.ListObjectsV2WithContext(ctx, &s3.ListObjectsV2Input{
        Bucket: aws.String(p.config.Bucket),
        Prefix: aws.String(prefix),
    })
    if err != nil {
        return nil, err
    }
    
    var keys []string
    for _, obj := range result.Contents {
        keys = append(keys, *obj.Key)
    }
    return keys, nil
}

func (p *S3Plugin) Exists(ctx context.Context, path string) (bool, error) {
    _, err := p.client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
        Bucket: aws.String(p.config.Bucket),
        Key:    aws.String(path),
    })
    if err != nil {
        if aerr, ok := err.(awserr.Error); ok {
            if aerr.Code() == "NotFound" {
                return false, nil
            }
        }
        return false, err
    }
    return true, nil
}

var Plugin S3Plugin
```

## 插件管理规范

### 插件加载器
```go
// internal/plugin/manager.go
package plugin

import (
    "context"
    "fmt"
    "plugin"
    "sync"
    
    "go.uber.org/zap"
)

type Manager struct {
    plugins map[string]*LoadedPlugin
    mutex   sync.RWMutex
    logger  *zap.Logger
}

type LoadedPlugin struct {
    Plugin   Plugin
    Metadata *Metadata
    Config   map[string]interface{}
}

func NewManager(logger *zap.Logger) *Manager {
    return &Manager{
        plugins: make(map[string]*LoadedPlugin),
        logger:  logger,
    }
}

func (m *Manager) LoadPlugin(ctx context.Context, path string, config map[string]interface{}) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    // 加载插件动态库
    plug, err := plugin.Open(path)
    if err != nil {
        return fmt.Errorf("failed to open plugin %s: %w", path, err)
    }
    
    // 获取插件符号
    sym, err := plug.Lookup("Plugin")
    if err != nil {
        return fmt.Errorf("plugin %s does not export Plugin symbol: %w", path, err)
    }
    
    // 类型断言
    pluginInstance, ok := sym.(Plugin)
    if !ok {
        return fmt.Errorf("plugin %s does not implement Plugin interface", path)
    }
    
    // 初始化插件
    if err := pluginInstance.Initialize(ctx, config); err != nil {
        return fmt.Errorf("failed to initialize plugin %s: %w", path, err)
    }
    
    // 保存插件实例
    m.plugins[pluginInstance.Name()] = &LoadedPlugin{
        Plugin: pluginInstance,
        Config: config,
    }
    
    m.logger.Info("plugin loaded successfully",
        zap.String("name", pluginInstance.Name()),
        zap.String("version", pluginInstance.Version()),
        zap.String("path", path),
    )
    
    return nil
}

func (m *Manager) GetFormatPlugin(format string) (FormatPlugin, error) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    
    for _, loaded := range m.plugins {
        if fp, ok := loaded.Plugin.(FormatPlugin); ok && fp.Format() == format {
            return fp, nil
        }
    }
    
    return nil, fmt.Errorf("format plugin not found: %s", format)
}
```

### 插件配置
```yaml
# config.yaml中的插件配置
plugins:
  enabled: ["maven", "npm", "s3-storage"]
  path: "/var/lib/go-nexus/plugins"
  
  maven:
    checksum_policy: "strict"
    metadata_update_policy: "daily"
  
  npm:
    registry_url: "https://registry.npmjs.org"
  
  s3-storage:
    endpoint: "https://s3.amazonaws.com"
    region: "us-east-1"
    bucket: "go-nexus-artifacts"
    access_key_id: "${AWS_ACCESS_KEY_ID}"
    secret_access_key: "${AWS_SECRET_ACCESS_KEY}"
```

## 插件开发工具

### 插件脚手架
```bash
#!/bin/bash
# create-plugin.sh

PLUGIN_NAME=$1
PLUGIN_TYPE=$2

if [ -z "$PLUGIN_NAME" ] || [ -z "$PLUGIN_TYPE" ]; then
    echo "Usage: $0 <plugin-name> <plugin-type>"
    echo "Plugin types: format, storage, integration"
    exit 1
fi

mkdir -p "plugins/$PLUGIN_NAME"
cd "plugins/$PLUGIN_NAME"

# 生成基础文件结构
cat > plugin.yaml << EOF
name: ${PLUGIN_NAME}-plugin
version: 1.0.0
type: $PLUGIN_TYPE
description: $PLUGIN_NAME plugin for go-nexus
author: Your Name
license: Apache-2.0
go_version: 1.21+

entry_point: ./${PLUGIN_NAME}-plugin.so
config_schema:
  properties: {}

permissions: []

compatibility:
  go_nexus_version: ">=0.1.0"
EOF

# 生成插件代码模板
case $PLUGIN_TYPE in
    "format")
        generate_format_plugin_template
        ;;
    "storage")
        generate_storage_plugin_template
        ;;
    "integration")
        generate_integration_plugin_template
        ;;
esac

echo "Plugin $PLUGIN_NAME created successfully in plugins/$PLUGIN_NAME/"
```

### 插件测试框架
```go
// pkg/plugin/testing/framework.go
package testing

import (
    "context"
    "testing"
    
    "github.com/laolishu/go-nexus/pkg/plugin"
)

// TestSuite 插件测试套件
type TestSuite struct {
    Plugin plugin.Plugin
    Config map[string]interface{}
}

func NewTestSuite(p plugin.Plugin, config map[string]interface{}) *TestSuite {
    return &TestSuite{
        Plugin: p,
        Config: config,
    }
}

func (ts *TestSuite) TestBasicInterface(t *testing.T) {
    // 测试基础接口
    name := ts.Plugin.Name()
    if name == "" {
        t.Error("Plugin name should not be empty")
    }
    
    version := ts.Plugin.Version()
    if version == "" {
        t.Error("Plugin version should not be empty")
    }
}

func (ts *TestSuite) TestInitialization(t *testing.T) {
    ctx := context.Background()
    err := ts.Plugin.Initialize(ctx, ts.Config)
    if err != nil {
        t.Errorf("Plugin initialization failed: %v", err)
    }
}

func (ts *TestSuite) RunAllTests(t *testing.T) {
    t.Run("BasicInterface", ts.TestBasicInterface)
    t.Run("Initialization", ts.TestInitialization)
}
```