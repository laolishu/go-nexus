package logger

import (
	"github.com/google/wire"
)

// ProviderSet 日志层的 Wire 提供者集合
var ProviderSet = wire.NewSet(
	NewLogger,
)
