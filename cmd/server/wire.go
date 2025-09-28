//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/laolishu/go-nexus/internal/app"
	"github.com/laolishu/go-nexus/internal/config"
	"github.com/laolishu/go-nexus/internal/handler"
	"github.com/laolishu/go-nexus/internal/plugin"
	"github.com/laolishu/go-nexus/internal/repository"
	"github.com/laolishu/go-nexus/internal/service"
	"github.com/laolishu/go-nexus/internal/storage"
	"github.com/laolishu/go-nexus/pkg/logger"
)

// InitializeApp 初始化整个应用程序
func InitializeApp(configFile string, logLevel string) (*app.App, func(), error) {
	wire.Build(
		// 配置层
		config.ProviderSet,

		// 日志层
		logger.ProviderSet,

		// 存储层
		storage.ProviderSet,

		// 数据层
		repository.ProviderSet,

		// 服务层
		service.ProviderSet,

		// 插件层
		plugin.ProviderSet,

		// 处理层
		handler.ProviderSet,

		// 应用层
		app.ProviderSet,
	)
	return nil, nil, nil
}
