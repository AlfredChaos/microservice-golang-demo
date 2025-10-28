package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// CORS 跨域资源共享中间件
// 处理跨域请求，允许前端应用访问API
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 允许的请求来源
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		
		// 允许的请求方法
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		
		// 允许的请求头
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, X-Trace-ID")
		
		// 允许暴露的响应头
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, X-Request-ID")
		
		// 预检请求缓存时间（秒）
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		
		// 允许携带凭证
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		
		// 处理 OPTIONS 预检请求
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}
