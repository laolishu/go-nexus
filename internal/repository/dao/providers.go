package dao

import (
	"github.com/google/wire"
)

// ProviderSet DAO层的 Wire 提供者集合
var ProviderSet = wire.NewSet(
	NewRepositoryDAO,
	NewArtifactDAO,
)
