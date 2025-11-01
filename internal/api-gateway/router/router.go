package router

import (
	"time"

	"github.com/alfredchaos/demo/internal/api-gateway/dependencies"
	"github.com/alfredchaos/demo/internal/api-gateway/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter(appCtx *dependencies.AppContext) *gin.Engine {
	// 创建 Gin 引擎（不使用默认中间件）
	router := gin.New()

	// 应用全局中间件（顺序很重要）
	router.Use(
		middleware.Recovery(),              // 1. Panic恢复（最先执行，确保能捕获所有panic）
		middleware.RequestID(),             // 2. 请求ID生成（用于后续日志追踪）
		middleware.Logger(),                // 3. 请求日志记录
		middleware.CORS(),                  // 4. 跨域处理
		middleware.Timeout(30*time.Second), // 5. 请求超时（30秒）
	)

	// API 路由组
	apiV1 := router.Group("/api/v1")
	{
		// 用户路由
		UserRouter(apiV1, appCtx.UserController)
		// 图书路由
		BookRouter(apiV1, appCtx.BookController)
		// 可以继续添加更多路由
		// OrderRouter(apiV1, appCtx.OrderController)
	}

	// 系统路由组
	SystemRouter(router)

	return router
}
