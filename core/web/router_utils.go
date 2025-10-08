/*
 * @Descripttion: 路由注册工具函数
 * @version:
 * @Author: lfzxs@qq.com
 * @Date: 2025-10-08 15:30:00
 * @LastEditors: lfzxs@qq.com
 * @LastEditTime: 2025-10-08 15:30:00
 */
package web

import (
	"github.com/gin-gonic/gin"
	"github.com/laolishu/go-nexus/core/global"
)

// RegisterRootHandle 向根路由注册路径和处理方法
// httpMethod: HTTP方法 (GET, POST, PUT, DELETE, etc.)
// relativePath: 相对路径
// handlers: 处理函数链，最后一个应该是真正的处理函数，前面的是中间件
func RegisterRootHandle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) {
	if global.RootRouter == nil {
		panic("RootRouter is not initialized")
	}
	global.RootRouter.Handle(httpMethod, relativePath, handlers...)
}

// RegisterApiHandle 向API v1路由组注册路径和处理方法
// httpMethod: HTTP方法 (GET, POST, PUT, DELETE, etc.)
// relativePath: 相对路径 (相对于/api/v1)
// handlers: 处理函数链，最后一个应该是真正的处理函数，前面的是中间件
func RegisterApiHandle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) {
	if global.APIv1Router == nil {
		panic("APIv1Router is not initialized")
	}
	global.APIv1Router.Handle(httpMethod, relativePath, handlers...)
}

// SetupRootRoutes 设置根路径(/)的路由
func SetupRootRoutes(router *gin.Engine) {
	// 根路径重定向到健康检查或API文档
	router.GET("/", func(c *gin.Context) {
		Success(c, gin.H{
			"name":        "go-nexus",
			"description": "轻量云原生仓库管理工具",
			"version":     "v0.5.0",
			"api_docs":    "/api/v1/docs",
			"health":      "/health",
		})
	})

	// API根路径信息
	router.GET("/api", func(c *gin.Context) {
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
	router *gin.Engine,
	repositoryHandler interface{},
	artifactHandler interface{},
) *gin.RouterGroup {
	// 返回已经初始化的全局API v1路由组
	return global.APIv1Router
}

// SetupMiddlewares 设置全局中间件
func SetupMiddlewares(router *gin.Engine) {
	// CORS 中间件（如果需要）
	router.Use(func(c *gin.Context) {
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
	router.Use(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// 如果没有请求ID，会在响应时自动生成UUID
		} else {
			c.Set("request_id", requestID)
		}
		c.Next()
	})
}
