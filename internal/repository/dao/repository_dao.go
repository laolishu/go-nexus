package dao

import (
	"log/slog"

	"gorm.io/gorm"
)

// RepositoryDAO 仓库数据访问对象
type RepositoryDAO struct {
	logger *slog.Logger
	db     *gorm.DB
}

// NewRepositoryDAO 创建新的仓库数据访问对象
func NewRepositoryDAO(logger *slog.Logger, db *gorm.DB) *RepositoryDAO {
	return &RepositoryDAO{
		logger: logger,
		db:     db,
	}
}
