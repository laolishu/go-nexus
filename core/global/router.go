/*
 * @Descripttion: 全局路由变量定义
 * @version:
 * @Author: lfzxs@qq.com
 * @Date: 2025-10-08 15:30:00
 * @LastEditors: lfzxs@qq.com
 * @LastEditTime: 2025-10-08 18:29:08
 */
package global

import "github.com/gin-gonic/gin"

var (
	// RootRouter 根路由引擎实例
	RootRouter *gin.Engine
	
	// APIv1Router API v1路由组实例
	APIv1Router *gin.RouterGroup
)