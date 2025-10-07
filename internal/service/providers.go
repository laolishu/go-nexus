package service

import (
	"github.com/google/wire"

	"github.com/laolishu/go-nexus/internal/service/impl"
)

// ProviderSet 服务层的 Wire 提供者集合
var ProviderSet = wire.NewSet(
	// 绑定接口实现
	wire.Bind(new(RepositoryService), new(*impl.RepositoryServiceImpl)),
	wire.Bind(new(ArtifactService), new(*impl.ArtifactServiceImpl)),

	// 注册工厂函数
	impl.NewRepositoryService,
	impl.NewArtifactService,
)
