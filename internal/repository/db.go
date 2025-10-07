package repository

import (
	"fmt"

	"github.com/laolishu/go-nexus/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// NewDB 创建数据库连接
func NewDB(cfg *config.Config) (*gorm.DB, func(), error) {
	var dialector gorm.Dialector

	switch cfg.Database.Type {
	case "sqlite":
		dialector = sqlite.Open(cfg.Database.DSN)
	case "postgresql":
		dialector = postgres.Open(cfg.Database.DSN)
	default:
		return nil, nil, fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "gn_", // 表前缀
			SingularTable: true,  // 使用单数表名
		},
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层SQL连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get database connection pool: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// 返回清理函数
	cleanup := func() {
		sqlDB.Close()
	}

	return db, cleanup, nil
}
