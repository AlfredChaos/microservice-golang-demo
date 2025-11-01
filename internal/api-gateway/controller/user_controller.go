package controller

import (
	"net/http"

	"github.com/alfredchaos/demo/internal/api-gateway/domain"
	"github.com/alfredchaos/demo/internal/api-gateway/dto"
	"github.com/alfredchaos/demo/pkg/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// IUserController 用户控制器接口
type IUserController interface {
	SayHello(c *gin.Context)
}

// userController 用户控制器实现
type userController struct {
	userService domain.IUserService
}

// NewUserController 创建用户控制器
// 依赖领域服务接口
func NewUserController(userService domain.IUserService) IUserController {
	return &userController{
		userService: userService,
	}
}

// SayHello 处理问候请求
// @Summary 问候接口
// @Description 调用 user-service 并返回问候语
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=dto.HelloResponse} "成功响应"
// @Failure 500 {object} dto.Response "服务器错误"
// @Router /api/v1/user/hello [get]
func (ctrl *userController) SayHello(c *gin.Context) {
	ctx := c.Request.Context()

	// 使用 WithContext 自动附加请求上下文信息
	log.WithContext(ctx).Info("received user hello request")

	// 调用用户服务
	message, err := ctrl.userService.SayHello(ctx)
	if err != nil {
		log.WithContext(ctx).Error("failed to call user service", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(10001, "failed to call user service"))
		return
	}

	log.WithContext(ctx).Info("user hello request completed", zap.String("message", message))

	// 返回响应
	c.JSON(http.StatusOK, dto.NewSuccessResponse(dto.HelloResponse{
		Message: message,
	}))
}
