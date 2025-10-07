package logger

import (
	"github.com/google/wire"
)

// ProviderSet 日志层的 Wire 提供者集合
var ProviderSet = wire.NewSet(
	NewLogger, // 只保留 NewLogger 作为 *slog.Logger 的提供者
)
