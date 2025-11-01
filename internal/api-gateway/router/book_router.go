package router

import (
	"github.com/alfredchaos/demo/internal/api-gateway/controller"
	"github.com/gin-gonic/gin"
)

// BookRouter 图书路由组
func BookRouter(router *gin.RouterGroup, controller controller.IBookController) {
	bookGroup := router.Group("/book")
	{
		bookGroup.GET("", controller.GetBook)
		// 可以添加更多图书相关路由
		// bookGroup.GET("/:id", controller.GetBookByID)
		// bookGroup.POST("", controller.CreateBook)
		// bookGroup.PUT("/:id", controller.UpdateBook)
		// bookGroup.DELETE("/:id", controller.DeleteBook)
	}
}
