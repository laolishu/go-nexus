/*
 * @Descripttion:
 * @version:
 * @Author: lfzxs@qq.com
 * @Date: 2025-10-08 10:56:26
 * @LastEditors: lfzxs@qq.com
 * @LastEditTime: 2025-10-08 20:48:17
 */
package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laolishu/go-nexus/core/global"
	"github.com/laolishu/go-nexus/internal/handler"
	"github.com/laolishu/go-nexus/pkg/sysinfo"
)

// SetupHealthCheck 设置健康检查路由
func SetupHealthCheck() {
	relativePath := "/health"
	//使用 RegisterRootHandle 注册路由
	RegisterRootHandle(http.MethodGet, relativePath, func(c *gin.Context) {
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
	SetupRootRoutes()

	// 设置健康检查路由
	SetupHealthCheck()

	// 设置API v1路由
	SetupAPIv1Routes(repositoryHandler, artifactHandler)
}

// SetupRootRoutes 设置根路径(/)的路由
func SetupRootRoutes() {
	// 根路径重定向到健康检查或API文档
	RegisterRootHandle(http.MethodGet, "/", func(c *gin.Context) {
		Success(c, gin.H{
			"name":        "go-nexus",
			"description": "轻量云原生仓库管理工具",
			"version":     "v0.5.0",
			"api_docs":    "/api/v1/docs",
			"health":      "/health",
		})
	})

	// API根路径信息
	RegisterRootHandle(http.MethodGet, "/api", func(c *gin.Context) {
		Success(c, gin.H{
			"name":     "go-nexus API",
			"version":  "v1",
			"base_url": "/api/v1",
			"endpoints": gin.H{
				"repositories": "/api/v1/repositories",
				"health":       "/health",
			},
		})
	})
}

// SetupAPIv1Routes 设置API v1路径(/api/v1)的路由
func SetupAPIv1Routes(
	repositoryHandler *handler.RepositoryHandler,
	artifactHandler *handler.ArtifactHandler,
) *gin.RouterGroup {
	// API信息接口
	RegisterApiHandle(http.MethodGet, "", func(c *gin.Context) {
		Success(c, gin.H{
			"name":    "go-nexus API v1",
			"version": "1.0.0",
			"endpoints": gin.H{
				"repositories": "/api/v1/repositories",
				"artifacts":    "/api/v1/repositories/{id}/artifacts",
			},
		})
	})

	// 仓库管理路由
	if repositoryHandler != nil {
		RegisterApiHandle(http.MethodGet, "/repositories", repositoryHandler.ListRepositories)
		RegisterApiHandle(http.MethodPost, "/repositories", repositoryHandler.CreateRepository)
		RegisterApiHandle(http.MethodGet, "/repositories/:id", repositoryHandler.GetRepository)
		RegisterApiHandle(http.MethodPut, "/repositories/:id", repositoryHandler.UpdateRepository)
		RegisterApiHandle(http.MethodDelete, "/repositories/:id", repositoryHandler.DeleteRepository)
	}

	// Artifact 管理路由（嵌套在仓库路由下）
	if artifactHandler != nil {
		RegisterApiHandle(http.MethodGet, "/repositories/:id/artifacts", artifactHandler.ListArtifacts)
		RegisterApiHandle(http.MethodPost, "/repositories/:id/artifacts", artifactHandler.UploadArtifact)
		RegisterApiHandle(http.MethodGet, "/repositories/:id/artifacts/*path", artifactHandler.DownloadArtifact)
		RegisterApiHandle(http.MethodDelete, "/repositories/:id/artifacts/*path", artifactHandler.DeleteArtifact)
	}

	return global.APIv1Router
}

// SetupMiddlewares 设置全局中间件
func SetupMiddlewares(router *gin.Engine) {
	// CORS 中间件（如果需要）
	RegisterMiddleware(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 请求ID中间件
	RegisterMiddleware(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// 如果没有请求ID，会在响应时自动生成UUID
		} else {
			c.Set("request_id", requestID)
		}
		c.Next()
	})
}
