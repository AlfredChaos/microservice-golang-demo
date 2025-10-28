package middleware

import (
	"time"

	"github.com/alfredchaos/demo/pkg/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger 请求日志中间件
// 记录每个HTTP请求的详细信息，包括请求方法、路径、状态码、耗时等
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		startTime := time.Now()

		// 获取请求路径
		path := c.Request.URL.Path

		// 获取请求ID
		requestID := GetRequestID(c)

		// 处理请求
		c.Next()

		// 计算请求耗时
		latency := time.Since(startTime)

		// 获取响应状态码
		statusCode := c.Writer.Status()

		// 获取客户端IP
		clientIP := c.ClientIP()

		// 获取请求方法
		method := c.Request.Method

		// 获取错误信息（如果有）
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// 记录日志
		fields := []zap.Field{
			zap.String("X-Request-ID", requestID),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.String("client_ip", clientIP),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		// 如果有错误，添加错误信息
		if errorMessage != "" {
			fields = append(fields, zap.String("error", errorMessage))
		}

		// 根据状态码选择日志级别
		if statusCode >= 500 {
			log.Error("HTTP request error", fields...)
		} else if statusCode >= 400 {
			log.Warn("HTTP request warning", fields...)
		} else {
			log.Info("HTTP request", fields...)
		}
	}
}
