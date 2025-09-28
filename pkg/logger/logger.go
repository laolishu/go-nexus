package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/laolishu/go-nexus/internal/config"
)

// NewLogger 创建新的日志记录器
func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	// 设置日志级别
	level, err := zapcore.ParseLevel(cfg.Log.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	// 设置编码器配置
	var encoderConfig zapcore.EncoderConfig
	if cfg.Log.Format == "json" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 设置编码器
	var encoder zapcore.Encoder
	if cfg.Log.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 设置输出
	var writeSyncer zapcore.WriteSyncer
	if cfg.Log.Output == "stdout" {
		writeSyncer = zapcore.AddSync(os.Stdout)
	} else {
		// 文件输出，使用 lumberjack 进行日志轮转
		lumberJackLogger := &lumberjack.Logger{
			Filename:   cfg.Log.FilePath,
			MaxSize:    cfg.Log.MaxSize,
			MaxBackups: cfg.Log.MaxBackups,
			MaxAge:     cfg.Log.MaxAge,
			Compress:   true,
		}
		writeSyncer = zapcore.AddSync(lumberJackLogger)
	}

	// 创建核心
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建日志记录器
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	// 添加全局字段
	logger = logger.With(
		zap.String("service", "go-nexus"),
		zap.String("version", "dev"), // 可以从构建时注入
	)

	return logger, nil
}

// NewDevelopmentLogger 创建开发环境日志记录器
func NewDevelopmentLogger() (*zap.Logger, error) {
	return zap.NewDevelopment()
}
