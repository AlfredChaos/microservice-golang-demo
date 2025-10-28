package router

import (
	"time"

	"github.com/alfredchaos/demo/internal/api-gateway/controller"
	"github.com/alfredchaos/demo/internal/api-gateway/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 设置路由
func SetupRouter(helloController *controller.HelloController) *gin.Engine {
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
		// 问候接口
		apiV1.POST("/hello", helloController.SayHello)
	}

	// Swagger 文档路由（不需要超时限制）
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查（不需要超时限制）
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	return router
}
