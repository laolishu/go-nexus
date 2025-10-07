package dao

import (
	"github.com/google/wire"
)

// ProviderSet is the Wire provider set for the DAO layer.
var ProviderSet = wire.NewSet(
	NewRepositoryDAO,
	NewArtifactDAO,
)
