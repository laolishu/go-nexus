/*
 * @Descripttion:
 * @version:
 * @Author: lfzxs@qq.com
 * @Date: 2025-09-28 23:43:41
 * @LastEditors: lfzxs@qq.com
 * @LastEditTime: 2025-10-07 22:57:46
 */
package repository

import (
	"github.com/google/wire"
	"github.com/laolishu/go-nexus/internal/repository/dao"
	"github.com/laolishu/go-nexus/internal/repository/impl"
	"github.com/laolishu/go-nexus/pkg/config"
	"gorm.io/gorm"
)

// DBOut 用于 wire 注入数据库和 cleanup
type DBOut struct {
	DB      *gorm.DB
	Cleanup func()
}

// ProvideDB 适配 wire 的 cleanup 注入
func ProvideDB(cfg *config.Config) (DBOut, error) {
	db, cleanup, err := NewDB(cfg)
	return DBOut{DB: db, Cleanup: cleanup}, err
}

// ProviderSet is the Wire provider set for the repository layer.
var ProviderSet = wire.NewSet(
	ProvideDB,
	dao.NewRepositoryDAO,
	dao.NewArtifactDAO,
	impl.NewRepositoryRepository,
	impl.NewArtifactRepository,
	wire.Bind(new(RepositoryRepository), new(*impl.RepositoryRepositoryImpl)),
	wire.Bind(new(ArtifactRepository), new(*impl.ArtifactRepositoryImpl)),
)
