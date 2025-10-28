package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/alfredchaos/demo/pkg/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery 自定义的panic恢复中间件
// 捕获panic，记录错误日志，并返回500错误
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取请求ID
				requestID := GetRequestID(c)

				// 获取堆栈信息
				stackBytes := debug.Stack()
				stackStr := string(stackBytes)

				// 将堆栈按行分割，便于日志查看
				stackLines := strings.Split(stackStr, "\n")

				// 过滤空行
				var filteredStack []string
				for _, line := range stackLines {
					if strings.TrimSpace(line) != "" {
						filteredStack = append(filteredStack, line)
					}
				}

				// 记录错误日志
				log.Error("Panic recovered",
					zap.String("request_id", requestID),
					zap.String("panic_error", fmt.Sprintf("%v", err)),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("client_ip", c.ClientIP()),
					zap.String("user_agent", c.Request.UserAgent()),
					zap.Strings("stack_trace", filteredStack),
				)

				// 返回500错误
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":       500,
					"message":    "Internal server error",
					"request_id": requestID,
				})

				// 终止请求处理
				c.Abort()
			}
		}()

		c.Next()
	}
}
