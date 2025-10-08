/*
 * @Descripttion:
 * @version:
 * @Author: lfzxs@qq.com
 * @Date: 2025-10-08 10:56:26
 * @LastEditors: lfzxs@qq.com
 * @LastEditTime: 2025-10-08 11:14:35
 */
package web

import (
	"github.com/gin-gonic/gin"
	"github.com/laolishu/go-nexus/core/global"
	"github.com/laolishu/go-nexus/internal/handler"
	"github.com/laolishu/go-nexus/pkg/sysinfo"
)

// SetupHealthCheck 设置健康检查路由
func SetupHealthCheck(router *gin.Engine) {
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		info, err := sysinfo.GetInfo()
		if err != nil {
			InternalServerError(c, "Failed to get system info: "+err.Error())
			return
		}

		healthData := gin.H{
			"status":       "ok",
			"cpu_usage":    info.CPUUsage,
			"memory_usage": info.MemoryUsage,
		}

		Success(c, healthData)
	})
}

// SetupRoutes 设置所有路由
func SetupRoutes(
	router *gin.Engine,
	repositoryHandler *handler.RepositoryHandler,
	artifactHandler *handler.ArtifactHandler,
) {
	// 先对全局变量赋值
	global.RootRouter = router
	global.APIv1Router = router.Group("/api/v1")

	// 设置全局中间件
	SetupMiddlewares(router)

	// 设置根路径路由
	SetupRootRoutes(router)

	// 设置健康检查路由
	SetupHealthCheck(router)

	// 设置API v1路由
	SetupAPIv1Routes(router, repositoryHandler, artifactHandler)
}
