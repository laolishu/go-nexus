//go:build wireinject
// +build wireinject

/*
 * @Descripttion:
 * @version:
 * @Author: lfzxs@qq.com
 * @Date: 2025-09-28 23:40:30
 * @LastEditors: lfzxs@qq.com
 * @LastEditTime: 2025-10-07 22:58:40
 */

package main

import (
	"github.com/google/wire"

	"github.com/laolishu/go-nexus/internal/app"
	"github.com/laolishu/go-nexus/internal/handler"
	"github.com/laolishu/go-nexus/internal/repository"
	"github.com/laolishu/go-nexus/internal/service"
	"github.com/laolishu/go-nexus/pkg/config"
	"github.com/laolishu/go-nexus/pkg/logger"
)

// InitializeApp 初始化整个应用程序
func InitializeApp(configFile string) (*app.App, func(), error) {
	wire.Build(
		config.LoadConfig,
		logger.NewLogger,
		repository.NewDB,
		repository.ProviderSet,
		service.ProviderSet,
		handler.NewRepositoryHandler,
		handler.NewArtifactHandler,
		app.NewApp,
	)
	return nil, nil, nil
}
