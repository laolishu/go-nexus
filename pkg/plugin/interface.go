package plugin

import (
	"context"
	"time"
)

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

// Metadata artifact元数据
type Metadata struct {
	GroupID      string            `json:"group_id"`
	ArtifactID   string            `json:"artifact_id"`
	Version      string            `json:"version"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Packaging    string            `json:"packaging"`
	Keywords     []string          `json:"keywords"`
	Dependencies map[string]string `json:"dependencies"`
	Properties   map[string]string `json:"properties"`
}

// Artifact artifact信息
type Artifact struct {
	ID           string            `json:"id"`
	RepositoryID string            `json:"repository_id"`
	Path         string            `json:"path"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Format       string            `json:"format"`
	Size         int64             `json:"size"`
	Checksum     string            `json:"checksum"`
	Metadata     *Metadata         `json:"metadata"`
	Properties   map[string]string `json:"properties"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// ArtifactEvent artifact事件
type ArtifactEvent struct {
	Type       string    `json:"type"` // uploaded, deleted, updated
	Artifact   *Artifact `json:"artifact"`
	Repository string    `json:"repository"`
	User       string    `json:"user"`
	Timestamp  time.Time `json:"timestamp"`
}

// RepositoryEvent 仓库事件
type RepositoryEvent struct {
	Type       string                 `json:"type"` // created, updated, deleted
	Repository *Repository            `json:"repository"`
	User       string                 `json:"user"`
	Changes    map[string]interface{} `json:"changes"`
	Timestamp  time.Time              `json:"timestamp"`
}

// Repository 仓库信息
type Repository struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`   // proxy, hosted, group
	Format      string            `json:"format"` // maven, npm, docker, etc.
	Description string            `json:"description"`
	URL         string            `json:"url"` // for proxy repositories
	Config      map[string]string `json:"config"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}
