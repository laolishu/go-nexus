package impl

import (
	"log/slog"

	"github.com/laolishu/go-nexus/internal/repository/dao"
)

// ArtifactRepositoryImpl 制品持久层实现
type ArtifactRepositoryImpl struct {
	logger *slog.Logger
	dao    *dao.ArtifactDAO
}

// NewArtifactRepository 创建新的制品持久层实现
func NewArtifactRepository(logger *slog.Logger, dao *dao.ArtifactDAO) *ArtifactRepositoryImpl {
	return &ArtifactRepositoryImpl{
		logger: logger,
		dao:    dao,
	}
}
