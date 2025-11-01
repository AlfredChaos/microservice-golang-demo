package controller

import (
	"net/http"

	"github.com/alfredchaos/demo/internal/api-gateway/domain"
	"github.com/alfredchaos/demo/internal/api-gateway/dto"
	"github.com/alfredchaos/demo/pkg/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// IBookController 图书控制器接口
type IBookController interface {
	GetBook(c *gin.Context)
}

// bookController 图书控制器实现
type bookController struct {
	bookService domain.IBookService
}

// NewBookController 创建图书控制器
// 依赖领域服务接口
func NewBookController(bookService domain.IBookService) IBookController {
	return &bookController{
		bookService: bookService,
	}
}

// GetBook 处理获取图书请求
// @Summary 获取图书
// @Description 调用 book-service 获取图书信息
// @Tags Book
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=string} "成功响应"
// @Failure 500 {object} dto.Response "服务器错误"
// @Router /api/v1/book [get]
func (ctrl *bookController) GetBook(c *gin.Context) {
	ctx := c.Request.Context()

	log.WithContext(ctx).Info("received get book request")

	// 调用图书服务
	message, err := ctrl.bookService.GetBook(ctx)
	if err != nil {
		log.WithContext(ctx).Error("failed to call book service", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(10002, "failed to call book service"))
		return
	}

	log.WithContext(ctx).Info("get book request completed", zap.String("message", message))

	// 返回响应
	c.JSON(http.StatusOK, dto.NewSuccessResponse(message))
}
