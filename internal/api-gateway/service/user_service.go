package service

import (
	"context"
	"fmt"

	userv1 "github.com/alfredchaos/demo/api/user/v1"
	"github.com/alfredchaos/demo/internal/api-gateway/domain"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

// userService 用户服务实现
// 封装对 user-service 的 gRPC 调用
type userService struct {
	baseService
	userClient userv1.UserServiceClient
}

// NewUserService 创建用户服务实例
// 注入 gRPC 客户端依赖
func NewUserService(userClient userv1.UserServiceClient) domain.IUserService {
	return &userService{
		baseService: baseService{},
		userClient:  userClient,
	}
}

// SayHello 调用 user-service 的 SayHello 接口
func (s *userService) SayHello(ctx context.Context) (string, error) {
	// 传递 trace ID 到 gRPC metadata
	ctx = s.withTraceID(ctx)

	// 调用 user-service
	resp, err := s.userClient.SayHello(ctx, &userv1.HelloRequest{})
	if err != nil {
		log.WithContext(ctx).Error("failed to call user service", zap.Error(err))
		return "", fmt.Errorf("failed to call user service: %w", err)
	}

	log.WithContext(ctx).Info("user service SayHello success", zap.String("message", resp.Message))
	return resp.Message, nil
}
