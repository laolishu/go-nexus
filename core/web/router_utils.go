/*
 * @Descripttion: 路由注册工具函数
 * @version:
 * @Author: lfzxs@qq.com
 * @Date: 2025-10-08 15:30:00
 * @LastEditors: lfzxs@qq.com
 * @LastEditTime: 2025-10-08 20:28:32
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

// RegisterMiddleware 向系统注册全局中间件
// middleware: 中间件函数链
func RegisterMiddleware(middleware ...gin.HandlerFunc) {
	if global.RootRouter == nil {
		panic("RootRouter is not initialized")
	}
	global.RootRouter.Use(middleware...)
}
