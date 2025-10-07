package impl

import (
	"log/slog"

	"github.com/laolishu/go-nexus/internal/repository/dao"
)

// RepositoryRepositoryImpl 仓库持久层实现
type RepositoryRepositoryImpl struct {
	logger *slog.Logger
	dao    *dao.RepositoryDAO
}

// NewRepositoryRepository 创建新的仓库持久层实现
func NewRepositoryRepository(logger *slog.Logger, dao *dao.RepositoryDAO) *RepositoryRepositoryImpl {
	return &RepositoryRepositoryImpl{
		logger: logger,
		dao:    dao,
	}
}
