package service

import (
	"github.com/google/wire"

	"github.com/laolishu/go-nexus/internal/service/impl"
)

// ProviderSet 服务层的 Wire 提供者集�?
var ProviderSet = wire.NewSet(
	impl.NewRepositoryService,
	wire.Bind(new(RepositoryService), new(*impl.RepositoryServiceImpl)),
	impl.NewArtifactService,
	wire.Bind(new(ArtifactService), new(*impl.ArtifactServiceImpl)),
)
