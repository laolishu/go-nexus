package handler

import (
	"github.com/google/wire"
)

// ProviderSet 处理器层的 Wire 提供者集合
var ProviderSet = wire.NewSet(
	NewRepositoryHandler,
	NewArtifactHandler,
)
