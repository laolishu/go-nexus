package app

import (
	"github.com/google/wire"
)

// ProviderSet 应用程序层的 Wire 提供者集合
var ProviderSet = wire.NewSet(
	NewApp,
)
