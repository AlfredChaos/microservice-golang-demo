package service

import (
	"context"

	userv1 "github.com/alfredchaos/demo/api/user/v1"
	"github.com/alfredchaos/demo/internal/user-service/biz"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

// UserService gRPC服务实现
type UserService struct {
	userv1.UnimplementedUserServiceServer
	useCase *biz.UserUseCase
}

// NewUserService 创建新的用户服务
func NewUserService(useCase *biz.UserUseCase) *UserService {
	return &UserService{
		useCase: useCase,
	}
}

// SayHello 实现UserService.SayHello方法
func (s *UserService) SayHello(ctx context.Context, req *userv1.HelloRequest) (*userv1.HelloResponse, error) {
	log.WithContext(ctx).Info("received SayHello request")

	// 调用业务逻辑层
	message, err := s.useCase.SayHello(ctx, "")
	if err != nil {
		log.WithContext(ctx).Error("failed to say hello", zap.Error(err))
		return nil, err
	}

	log.WithContext(ctx).Info("SayHello completed", zap.String("message", message))

	// 构造gRPC响应
	return &userv1.HelloResponse{
		Message: message,
	}, nil
}
