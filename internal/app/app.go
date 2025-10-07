package app

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/laolishu/go-nexus/internal/config"
	"github.com/laolishu/go-nexus/internal/handler"
	"github.com/laolishu/go-nexus/internal/service"
)

// slogWriter 实现 io.Writer，将 gin 日志重定向到 slog
type slogWriter struct {
	logger *slog.Logger
}

// Write 将 gin 的日志写入 slog
func (w slogWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))
	if msg != "" {
		w.logger.Info(msg, "source", "gin")
	}
	return len(p), nil
}

// App 应用程序主结构
type App struct {
	Config            *config.Config
	Logger            *slog.Logger
	Router            *gin.Engine
	RepositoryService service.RepositoryService
	ArtifactService   service.ArtifactService
}

// NewApp 创建新的应用程序实例
func NewApp(
	cfg *config.Config,
	logger *slog.Logger,
	repositoryHandler *handler.RepositoryHandler,
	artifactHandler *handler.ArtifactHandler,
	repositoryService service.RepositoryService,
	artifactService service.ArtifactService,
) *App {
	// 设置 Gin 模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := gin.New()

	// Gin日志重定向到slog
	router.Use(gin.LoggerWithWriter(slogWriter{logger: logger}))
	router.Use(gin.Recovery())

	// 设置路由
	setupRoutes(router, repositoryHandler, artifactHandler)

	return &App{
		Config:            cfg,
		Logger:            logger,
		Router:            router,
		RepositoryService: repositoryService,
		ArtifactService:   artifactService,
	}
}

// setupRoutes 设置路由
func setupRoutes(
	router *gin.Engine,
	repositoryHandler *handler.RepositoryHandler,
	artifactHandler *handler.ArtifactHandler,
) {
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/ready", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ready"})
	})

	// API 路由组
	v1 := router.Group("/api/v1")
	{
		// 仓库管理
		repositories := v1.Group("/repositories")
		{
			repositories.GET("", repositoryHandler.ListRepositories)
			repositories.POST("", repositoryHandler.CreateRepository)
			repositories.GET("/:id", repositoryHandler.GetRepository)
			repositories.PUT("/:id", repositoryHandler.UpdateRepository)
			repositories.DELETE("/:id", repositoryHandler.DeleteRepository)

			// Artifact 管理
			repositories.GET("/:id/artifacts", artifactHandler.ListArtifacts)
			repositories.POST("/:id/artifacts", artifactHandler.UploadArtifact)
			repositories.GET("/:id/artifacts/*path", artifactHandler.DownloadArtifact)
			repositories.DELETE("/:id/artifacts/*path", artifactHandler.DeleteArtifact)
		}
	}
}

// Run 运行应用程序
func (a *App) Run() error {
	addr := fmt.Sprintf("%s:%d", a.Config.Server.Host, a.Config.Server.Port)
	a.Logger.Info("Starting server", "address", addr)
	return a.Router.Run(addr)
}
