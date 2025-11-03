package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alfredchaos/demo/internal/nice-service/biz"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

// HandleService 消息处理服务
// 负责接收消息、解析消息、路由到具体的业务逻辑处理
type HandleService struct {
	taskUseCase *biz.TaskUseCase
}

// NewHandleService 创建新的消息处理服务
func NewHandleService(taskUseCase *biz.TaskUseCase) *HandleService {
	return &HandleService{
		taskUseCase: taskUseCase,
	}
}

// HandleMessage 处理接收到的消息
// 这是消息消费者的入口点
func (s *HandleService) HandleMessage(ctx context.Context, message []byte) error {
	log.WithContext(ctx).Info("received message from rabbitmq",
		zap.ByteString("raw_message", message))

	// 解析消息
	var taskMsg biz.TaskMessage
	if err := json.Unmarshal(message, &taskMsg); err != nil {
		log.WithContext(ctx).Error("failed to unmarshal message",
			zap.Error(err),
			zap.ByteString("message", message))
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	log.WithContext(ctx).Info("parsed task message",
		zap.String("user_id", taskMsg.UserID),
		zap.String("username", taskMsg.Username),
		zap.String("task_type", taskMsg.TaskType),
		zap.String("message", taskMsg.Message),
		zap.String("created_at", taskMsg.CreatedAt))

	// 根据任务类型路由到不同的业务逻辑处理器
	switch taskMsg.TaskType {
	case "sayhello":
		return s.taskUseCase.HandleSayHelloTask(ctx, &taskMsg)
	default:
		log.WithContext(ctx).Warn("unknown task type",
			zap.String("task_type", taskMsg.TaskType))
		return fmt.Errorf("unknown task type: %s", taskMsg.TaskType)
	}
}
