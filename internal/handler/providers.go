/*
 * @Descripttion:
 * @version:
 * @Author: lfzxs@qq.com
 * @Date: 2025-09-28 23:43:23
 * @LastEditors: lfzxs@qq.com
 * @LastEditTime: 2025-10-07 22:02:18
 */
package handler

import (
	"github.com/google/wire"
)

// ProviderSet 处理器层的 Wire 提供者集合
var ProviderSet = wire.NewSet(
	NewRepositoryHandler,
	NewArtifactHandler,
)

var HandlerSet = wire.NewSet(
	NewRepositoryHandler,
	NewArtifactHandler,
)
