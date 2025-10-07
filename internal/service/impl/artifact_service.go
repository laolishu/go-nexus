package impl

import (
	"log/slog"

	"github.com/laolishu/go-nexus/internal/repository"
)

// ArtifactServiceImpl 制品服务实现
type ArtifactServiceImpl struct {
	logger     *slog.Logger
	repository repository.ArtifactRepository
}

// NewArtifactService 创建新的制品服务实现
func NewArtifactService(logger *slog.Logger, repo repository.ArtifactRepository) *ArtifactServiceImpl {
	return &ArtifactServiceImpl{
		logger:     logger,
		repository: repo,
	}
}
