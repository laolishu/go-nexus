/*
 * @Descripttion:
 * @version:
 * @Author: lfzxs@qq.com
 * @Date: 2025-09-28 23:41:33
 * @LastEditors: lfzxs@qq.com
 * @LastEditTime: 2025-10-07 00:01:52
 */
package logger

import (
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/laolishu/go-nexus/pkg/config"
)

// NewLogger 创建新的日志记录器
func NewLogger(cfg *config.Config) *slog.Logger {
	// 设置日志级别
	var level slog.Level
	switch cfg.Log.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		// 默认使用 INFO 级别
		level = slog.LevelInfo
	}

	// 设置输出
	var output io.Writer
	if cfg.Log.Output == "stdout" {
		output = os.Stdout
	} else {
		// 文件输出，使用 lumberjack 进行日志轮转
		lumberJackLogger := &lumberjack.Logger{
			Filename:   cfg.Log.FilePath,
			MaxSize:    cfg.Log.MaxSize,
			MaxBackups: cfg.Log.MaxBackups,
			MaxAge:     cfg.Log.MaxAge,
			Compress:   true,
		}
		output = lumberJackLogger
	}

	// 设置处理选项
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}

	// 创建日志处理器
	var handler slog.Handler
	if cfg.Log.Format == "json" {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	// 创建日志记录器
	logger := slog.New(handler)

	// 添加全局属性
	logger = logger.With(
		"service", "go-nexus",
		"version", "dev", // 可以从构建时注入
	)

	return logger
}

// NewDevelopmentLogger 创建开发环境日志记录器
func NewDevelopmentLogger() *slog.Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})
	return slog.New(handler)
}
