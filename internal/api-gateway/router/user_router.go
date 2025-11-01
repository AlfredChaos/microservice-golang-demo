package router

import (
	"github.com/alfredchaos/demo/internal/api-gateway/controller"
	"github.com/gin-gonic/gin"
)

// UserRouter 用户路由组
func UserRouter(router *gin.RouterGroup, controller controller.IUserController) {
	userGroup := router.Group("/user")
	{
		userGroup.GET("/hello", controller.SayHello)
		// 可以添加更多用户相关路由
		// userGroup.GET("/:id", controller.GetUser)
		// userGroup.POST("", controller.CreateUser)
		// userGroup.PUT("/:id", controller.UpdateUser)
		// userGroup.DELETE("/:id", controller.DeleteUser)
	}
}
