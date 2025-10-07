package impl

import (
	"log/slog"

	"github.com/laolishu/go-nexus/internal/repository"
)

// RepositoryServiceImpl 仓库服务实现
type RepositoryServiceImpl struct {
	logger     *slog.Logger
	repository repository.RepositoryRepository
}

// NewRepositoryService 创建新的仓库服务实现
func NewRepositoryService(logger *slog.Logger, repo repository.RepositoryRepository) *RepositoryServiceImpl {
	return &RepositoryServiceImpl{
		logger:     logger,
		repository: repo,
	}
}
