package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/alfredchaos/demo/internal/api-gateway/client"
	"github.com/alfredchaos/demo/internal/api-gateway/dto"
	"github.com/alfredchaos/demo/pkg/log"
	"github.com/alfredchaos/demo/pkg/mq"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HelloController 问候控制器
type HelloController struct {
	grpcClients *client.GRPCClients
	publisher   mq.Publisher
}

// NewHelloController 创建新的问候控制器
// 使用依赖注入,接收 gRPC 客户端和消息发布者
func NewHelloController(grpcClients *client.GRPCClients, publisher mq.Publisher) *HelloController {
	return &HelloController{
		grpcClients: grpcClients,
		publisher:   publisher,
	}
}

// SayHello 处理问候请求
// @Summary 问候接口
// @Description 调用后端服务并返回问候语
// @Tags Hello
// @Accept json
// @Produce json
// @Param request body dto.HelloRequest true "请求参数"
// @Success 200 {object} dto.Response{data=string} "成功响应"
// @Failure 500 {object} dto.Response "服务器错误"
// @Router /api/v1/hello [post]
func (h *HelloController) SayHello(c *gin.Context) {
	log.Info("received hello request")
	
	// 创建上下文,设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// 并发调用 user-service 和 book-service
	type result struct {
		message string
		err     error
	}
	
	userChan := make(chan result, 1)
	bookChan := make(chan result, 1)
	
	// 调用 user-service
	go func() {
		msg, err := h.grpcClients.CallUserService(ctx)
		userChan <- result{message: msg, err: err}
	}()
	
	// 调用 book-service
	go func() {
		msg, err := h.grpcClients.CallBookService(ctx)
		bookChan <- result{message: msg, err: err}
	}()
	
	// 等待结果
	userResult := <-userChan
	bookResult := <-bookChan
	
	// 检查错误
	if userResult.err != nil {
		log.Error("failed to call user service", zap.Error(userResult.err))
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(10001, "failed to call user service"))
		return
	}
	
	if bookResult.err != nil {
		log.Error("failed to call book service", zap.Error(bookResult.err))
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(10001, "failed to call book service"))
		return
	}
	
	// 组合响应
	response := userResult.message + " " + bookResult.message
	log.Info("combined response", zap.String("response", response))
	
	// 发送消息到 RabbitMQ
	go func() {
		msgCtx, msgCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer msgCancel()
		
		message := map[string]string{
			"type":    "hello",
			"message": response,
		}
		
		msgBytes, err := json.Marshal(message)
		if err != nil {
			log.Error("failed to marshal message", zap.Error(err))
			return
		}
		
		if err := h.publisher.Publish(msgCtx, msgBytes); err != nil {
			log.Error("failed to publish message", zap.Error(err))
		} else {
			log.Info("message published to rabbitmq")
		}
	}()
	
	// 返回响应
	c.JSON(http.StatusOK, dto.NewSuccessResponse(response))
}
