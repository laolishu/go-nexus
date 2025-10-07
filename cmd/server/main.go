package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/laolishu/go-nexus/internal/constant"
	"github.com/spf13/cobra"
)

var (
	configFile string
	logLevel   string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "go-nexus",
		Short: "轻量云原生仓库管理工具",
		Long: `go-nexus 是一款基于 Golang 开发的轻量云原生仓库管理工具，
专为中小团队及云原生环境设计，旨在简化依赖管理流程。`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", constant.Version, constant.GitCommit, constant.BuildTime),
	}

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "resource/configs/config.yaml", "配置文件路径")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "日志级别 (debug, info, warn, error)")

	// 服务器命令
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "启动 go-nexus 服务器",
		RunE:  runServer,
	}

	// 迁移命令
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "运行数据库迁移",
		RunE:  runMigrate,
	}

	migrateCmd.Flags().String("direction", "up", "迁移方向 (up, down)")

	rootCmd.AddCommand(serverCmd, migrateCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runServer(cmd *cobra.Command, args []string) error {
	// 重写配置中的日志级别
	os.Setenv("GO_NEXUS_LOG_LEVEL", logLevel)

	// 初始化应用
	app, cleanup, err := InitializeApp(configFile)
	if err != nil {
		return fmt.Errorf("failed to initialize app: %w", err)
	}
	defer cleanup()

	// 启动服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.Config.Server.Port),
		Handler: app.Router,
	}

	// 优雅关闭
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		app.Logger.Info("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			app.Logger.Error("Server forced to shutdown", "error", err)
		}
	}()

	app.Logger.Info("Starting server",
		"version", constant.Version,
		"buildTime", constant.BuildTime,
		"gitCommit", constant.GitCommit,
		"port", app.Config.Server.Port,
	)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func runMigrate(cmd *cobra.Command, args []string) error {
	direction, _ := cmd.Flags().GetString("direction")

	// 重写配置中的日志级别
	os.Setenv("GO_NEXUS_LOG_LEVEL", logLevel)

	app, cleanup, err := InitializeApp(configFile)
	if err != nil {
		return fmt.Errorf("failed to initialize app: %w", err)
	}
	defer cleanup()

	app.Logger.Info("Running database migration", "direction", direction)

	// TODO: 实现数据库迁移逻辑

	return nil
}
