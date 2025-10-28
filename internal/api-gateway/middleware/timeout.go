package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout 请求超时中间件
// 为每个请求设置超时时间，防止请求长时间占用资源
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建带超时的上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 替换请求的上下文
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// 检查是否超时
		if ctx.Err() == context.DeadlineExceeded {
			// 如果已经返回了响应，不再重复返回
			if !c.Writer.Written() {
				c.JSON(http.StatusRequestTimeout, gin.H{
					"code":       408,
					"message":    "Request timeout",
					"request_id": GetRequestID(c),
				})
				c.Abort()
			}
		}
	}
}
