package model

import (
	"time"

	"gorm.io/gorm"
)

// Repository 仓库模型
type Repository struct {
	ID          string            `gorm:"primaryKey;size:36" json:"id"`
	Name        string            `gorm:"uniqueIndex;not null;size:100" json:"name"`
	Type        string            `gorm:"not null;size:20" json:"type"`   // proxy, hosted, group
	Format      string            `gorm:"not null;size:20" json:"format"` // maven, npm, docker, etc.
	Description string            `gorm:"size:500" json:"description"`
	URL         string            `gorm:"size:500" json:"url"` // for proxy repositories
	Config      map[string]string `gorm:"serializer:json" json:"config"`
	Status      string            `gorm:"default:active;size:20" json:"status"` // active, inactive
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	DeletedAt   gorm.DeletedAt    `gorm:"index" json:"-"`

	// 关联
	Artifacts []Artifact `gorm:"foreignKey:RepositoryID" json:"-"`
}

// Artifact 制品模型
type Artifact struct {
	ID            string            `gorm:"primaryKey;size:36" json:"id"`
	RepositoryID  string            `gorm:"not null;size:36;index" json:"repository_id"`
	Path          string            `gorm:"not null;size:1000;index" json:"path"`
	Name          string            `gorm:"not null;size:200" json:"name"`
	Version       string            `gorm:"not null;size:50" json:"version"`
	Format        string            `gorm:"not null;size:20" json:"format"`
	Size          int64             `gorm:"not null" json:"size"`
	Checksum      string            `gorm:"size:64" json:"checksum"`
	ContentType   string            `gorm:"size:100" json:"content_type"`
	Metadata      map[string]string `gorm:"serializer:json" json:"metadata"`
	Properties    map[string]string `gorm:"serializer:json" json:"properties"`
	DownloadCount int64             `gorm:"default:0" json:"download_count"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	DeletedAt     gorm.DeletedAt    `gorm:"index" json:"-"`

	// 关联
	Repository Repository `gorm:"foreignKey:RepositoryID" json:"-"`
}

// User 用户模型
type User struct {
	ID        string         `gorm:"primaryKey;size:36" json:"id"`
	Username  string         `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Email     string         `gorm:"uniqueIndex;not null;size:100" json:"email"`
	Password  string         `gorm:"not null;size:255" json:"-"` // 密码哈希，不返回给客户端
	FullName  string         `gorm:"size:100" json:"full_name"`
	Status    string         `gorm:"default:active;size:20" json:"status"` // active, inactive, locked
	LastLogin *time.Time     `json:"last_login"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Roles []Role `gorm:"many2many:user_roles;" json:"roles"`
}

// Role 角色模型
type Role struct {
	ID          string         `gorm:"primaryKey;size:36" json:"id"`
	Name        string         `gorm:"uniqueIndex;not null;size:50" json:"name"`
	Description string         `gorm:"size:200" json:"description"`
	Permissions []string       `gorm:"serializer:json" json:"permissions"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Users []User `gorm:"many2many:user_roles;" json:"-"`
}

// AccessToken 访问令牌模型
type AccessToken struct {
	ID        string         `gorm:"primaryKey;size:36" json:"id"`
	UserID    string         `gorm:"not null;size:36;index" json:"user_id"`
	Name      string         `gorm:"not null;size:100" json:"name"`
	Token     string         `gorm:"uniqueIndex;not null;size:255" json:"token"` // 令牌哈希
	Scopes    []string       `gorm:"serializer:json" json:"scopes"`
	ExpiresAt *time.Time     `json:"expires_at"`
	LastUsed  *time.Time     `json:"last_used"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// AuditLog 审计日志模型
type AuditLog struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	UserID    string    `gorm:"size:36;index" json:"user_id"`
	Action    string    `gorm:"not null;size:50" json:"action"`
	Resource  string    `gorm:"not null;size:100" json:"resource"`
	Details   string    `gorm:"type:text" json:"details"`
	IP        string    `gorm:"size:45" json:"ip"`
	UserAgent string    `gorm:"size:500" json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`

	// 关联
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName 方法定义表名
func (Repository) TableName() string {
	return "repositories"
}

func (Artifact) TableName() string {
	return "artifacts"
}

func (User) TableName() string {
	return "users"
}

func (Role) TableName() string {
	return "roles"
}

func (AccessToken) TableName() string {
	return "access_tokens"
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
