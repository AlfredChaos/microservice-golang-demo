package router

import (
	"github.com/alfredchaos/demo/internal/api-gateway/controller"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 设置路由
func SetupRouter(helloController *controller.HelloController) *gin.Engine {
	// 创建 Gin 引擎
	router := gin.Default()
	
	// API 路由组
	apiV1 := router.Group("/api/v1")
	{
		// 问候接口
		apiV1.POST("/hello", helloController.SayHello)
	}
	
	// Swagger 文档路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
	
	return router
}
