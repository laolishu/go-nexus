package app

import (
	"github.com/google/wire"
	"github.com/laolishu/go-nexus/internal/config"
)

// ProviderSet 应用程序层的 Wire 提供者集合
var ProviderSet = wire.NewSet(
	NewApp,
	wire.FieldsOf(new(*config.Config), "Server", "Database", "Log"),
)
