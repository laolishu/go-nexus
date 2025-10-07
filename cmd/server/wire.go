//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/laolishu/go-nexus/internal/app"
	"github.com/laolishu/go-nexus/internal/config"
	"github.com/laolishu/go-nexus/internal/handler"
	"github.com/laolishu/go-nexus/internal/repository"
	"github.com/laolishu/go-nexus/internal/repository/dao"
	repoimpl "github.com/laolishu/go-nexus/internal/repository/impl"
	"github.com/laolishu/go-nexus/internal/service"
	svcimpl "github.com/laolishu/go-nexus/internal/service/impl"
	"github.com/laolishu/go-nexus/pkg/logger"
)

// InitializeApp 初始化整个应用程序
func InitializeApp(configFile string) (*app.App, func(), error) {
	wire.Build(
		// 配置层
		config.LoadConfig,

		// 日志层
		logger.NewLogger,

		// 数据层
		repository.NewDB,
		dao.NewRepositoryDAO,
		dao.NewArtifactDAO,
		repoimpl.NewRepositoryRepository,
		repoimpl.NewArtifactRepository,
		wire.Bind(new(repository.RepositoryRepository), new(*repoimpl.RepositoryRepositoryImpl)),
		wire.Bind(new(repository.ArtifactRepository), new(*repoimpl.ArtifactRepositoryImpl)),

		// 服务层
		svcimpl.NewRepositoryService,
		svcimpl.NewArtifactService,
		wire.Bind(new(service.RepositoryService), new(*svcimpl.RepositoryServiceImpl)),
		wire.Bind(new(service.ArtifactService), new(*svcimpl.ArtifactServiceImpl)),

		// 处理层
		handler.NewRepositoryHandler,
		handler.NewArtifactHandler,

		// 应用层
		app.NewApp,
	)
	return nil, nil, nil
}
