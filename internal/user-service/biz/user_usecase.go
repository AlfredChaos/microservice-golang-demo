package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	bookv1 "github.com/alfredchaos/demo/api/book/v1"
	"github.com/alfredchaos/demo/internal/user-service/cache"
	"github.com/alfredchaos/demo/internal/user-service/domain"
	"github.com/alfredchaos/demo/internal/user-service/messaging"
	"github.com/alfredchaos/demo/internal/user-service/repository"
	"github.com/alfredchaos/demo/pkg/log"
	"github.com/alfredchaos/demo/pkg/mq"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UserUseCase 用户业务逻辑用例接口
type IUserUseCase interface {
	SayHello(ctx context.Context, name string) (string, error)
}

// userUseCase 用户业务逻辑用例实现
type UserUseCase struct {
	bookClient  bookv1.BookServiceClient
	userRepo    repository.UserRepository
	userDocRepo repository.UserDocumentRepository
	userCache   cache.UserCache
	publisher   messaging.Publisher
}

// NewUserUseCase 创建新的用户业务逻辑用例
func NewUserUseCase(
	bookClient bookv1.BookServiceClient,
	userRepo repository.UserRepository,
	userDocRepo repository.UserDocumentRepository,
	userCache cache.UserCache,
	publisher messaging.Publisher,
) *UserUseCase {
	return &UserUseCase{
		bookClient:  bookClient,
		userRepo:    userRepo,
		userDocRepo: userDocRepo,
		userCache:   userCache,
		publisher:   publisher,
	}
}

func (uc *UserUseCase) SayHello(ctx context.Context, name string) (string, error) {
	log.WithContext(ctx).Info("processing SayHello request", zap.String("name", name))

	// 1. 生成user-service的消息
	userMessage := "Hello from user-service"
	if name != "" {
		userMessage = "Hello " + name
	}

	// 2. 同步调用book-service获取消息
	log.Info("calling book-service via gRPC")
	bookResp, err := uc.bookClient.SayHello(ctx, &bookv1.HelloRequest{})
	if err != nil {
		log.Error("failed to call book-service", zap.Error(err))
		return "", err
	}
	bookMessage := bookResp.Message
	log.Info("received message from book-service", zap.String("message", bookMessage))

	// 3. 组合User结构
	user := domain.User{
		ID:       uuid.New().String(),
		Username: userMessage,
		Email:    bookMessage,
	}

	// 5. 保存用户
	if err := uc.userRepo.Create(ctx, &user); err != nil {
		log.Error("failed to create user", zap.Error(err))
		return "", err
	}

	// 6. 保存用户文档
	if err := uc.userDocRepo.SaveDocument(ctx, user.ID, map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
	}); err != nil {
		log.Error("failed to save user document", zap.Error(err))
		return "", err
	}

	// 7. 缓存用户
	if err := uc.userCache.SetUser(ctx, &user, 60); err != nil {
		log.Error("failed to cache user", zap.Error(err))
		return "", err
	}

	// 8. 发送异步任务消息（使用 Topic Exchange）
	// 构建任务消息
	taskMsg := map[string]interface{}{
		"user_id":    user.ID,
		"username":   user.Username,
		"task_type":  "sayhello",
		"message":    userMessage,
		"created_at": time.Now().Format(time.RFC3339),
	}
	taskData, err := json.Marshal(taskMsg)
	if err != nil {
		log.Error("failed to marshal task message", zap.Error(err))
		// 消息序列化失败不影响主流程，继续执行
	} else {
		// 使用 PublishWithRouting 发送到指定的 routing key
		if err := uc.publisher.PublishWithRouting(ctx, mq.RoutingKeyTaskSayHelloCreate, taskData); err != nil {
			log.Error("failed to publish task message",
				zap.Error(err),
				zap.String("routing_key", mq.RoutingKeyTaskSayHelloCreate))
		} else {
			log.Info("task message published successfully",
				zap.String("routing_key", mq.RoutingKeyTaskSayHelloCreate),
				zap.String("user_id", user.ID))
		}
	}

	// 9. 转成字符串
	userString := fmt.Sprintf("User{ID: %s, Username: %s, Email: %s}", user.ID, user.Username, user.Email)

	return userString, nil
}
