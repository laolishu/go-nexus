package repository

import (
	"github.com/google/wire"

	"github.com/laolishu/go-nexus/internal/repository/dao"
	"github.com/laolishu/go-nexus/internal/repository/impl"
)

// ProviderSet 数据层的 Wire 提供者集合
var ProviderSet = wire.NewSet(
	// 数据库连接
	NewDB,

	// DAO层
	dao.NewRepositoryDAO,
	dao.NewArtifactDAO,

	// 绑定接口实现
	wire.Bind(new(RepositoryRepository), new(*impl.RepositoryRepositoryImpl)),
	wire.Bind(new(ArtifactRepository), new(*impl.ArtifactRepositoryImpl)),

	// 数据层实现
	impl.NewRepositoryRepository,
	impl.NewArtifactRepository,
)
