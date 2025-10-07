package dao

import (
	"log/slog"

	"gorm.io/gorm"
)

// ArtifactDAO 制品数据访问对象
type ArtifactDAO struct {
	logger *slog.Logger
	db     *gorm.DB
}

// NewArtifactDAO 创建新的制品数据访问对象
func NewArtifactDAO(logger *slog.Logger, db *gorm.DB) *ArtifactDAO {
	return &ArtifactDAO{
		logger: logger,
		db:     db,
	}
}
