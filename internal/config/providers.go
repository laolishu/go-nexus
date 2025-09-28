package config

import (
	"github.com/google/wire"
)

// ProviderSet 配置层的 Wire 提供者集合
var ProviderSet = wire.NewSet(
	LoadConfig,
)
