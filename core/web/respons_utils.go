/*
 * @Descripttion:
 * @version:
 * @Author: lfzxs@qq.com
 * @Date: 2025-10-08 11:13:47
 * @LastEditors: lfzxs@qq.com
 * @LastEditTime: 2025-10-08 11:21:50
 */
package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// StandardResponse 标准响应体结构
type StandardResponse struct {
	Code      int         `json:"code"`      // 响应代码
	Msg       string      `json:"msg"`       // 响应信息
	Data      interface{} `json:"data"`      // 响应数据
	RequestID string      `json:"requestId"` // 请求ID
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, StandardResponse{
		Code:      http.StatusOK,
		Msg:       "success",
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// SuccessWithMsg 带自定义消息的成功响应
func SuccessWithMsg(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, StandardResponse{
		Code:      http.StatusOK,
		Msg:       msg,
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// Error 错误响应
func Error(c *gin.Context, httpStatus int, code int, msg string) {
	c.JSON(httpStatus, StandardResponse{
		Code:      code,
		Msg:       msg,
		Data:      nil,
		RequestID: getRequestID(c),
	})
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(c *gin.Context, httpStatus int, code int, msg string, data interface{}) {
	c.JSON(httpStatus, StandardResponse{
		Code:      code,
		Msg:       msg,
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// BadRequest 400错误响应
func BadRequest(c *gin.Context, msg string) {
	Error(c, http.StatusBadRequest, 400, msg)
}

// Unauthorized 401错误响应
func Unauthorized(c *gin.Context, msg string) {
	Error(c, http.StatusUnauthorized, 401, msg)
}

// Forbidden 403错误响应
func Forbidden(c *gin.Context, msg string) {
	Error(c, http.StatusForbidden, 403, msg)
}

// NotFound 404错误响应
func NotFound(c *gin.Context, msg string) {
	Error(c, http.StatusNotFound, 404, msg)
}

// InternalServerError 500错误响应
func InternalServerError(c *gin.Context, msg string) {
	Error(c, http.StatusInternalServerError, 500, msg)
}

// getRequestID 获取或生成请求ID
func getRequestID(c *gin.Context) string {
	// 先尝试从Header中获取请求ID
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}

	// 尝试从上下文中获取请求ID
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}

	// 如果没有找到，生成新的UUID作为请求ID
	return uuid.New().String()
}
