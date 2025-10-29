package middleware

import (
	"github.com/alfredchaos/demo/pkg/reqctx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDKey 请求ID在上下文中的键名
	RequestIDKey = "X-Request-ID"
)

// RequestID 请求ID中间件
// 为每个请求生成唯一ID，用于日志追踪和问题排查
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头获取请求ID
		requestID := c.GetHeader(RequestIDKey)

		// 如果没有，则生成新的UUID
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 将请求ID设置到 gin.Context 中
		c.Set(RequestIDKey, requestID)

		// 将请求ID添加到 request.Context 中
		ctx := reqctx.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		// 将请求ID设置到响应头中
		c.Writer.Header().Set(RequestIDKey, requestID)

		c.Next()
	}
}

// GetRequestID 从上下文中获取请求ID
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
