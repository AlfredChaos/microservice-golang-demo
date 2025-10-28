package service

import (
	"context"

	userv1 "github.com/alfredchaos/demo/api/user/v1"
	"github.com/alfredchaos/demo/internal/user-service/biz"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

// UserService gRPC 服务实现
// 作为胶水层,连接 gRPC 接口和业务逻辑层
type UserService struct {
	userv1.UnimplementedUserServiceServer
	useCase biz.UserUseCase
}

// NewUserService 创建新的用户服务
// 使用依赖注入,接收业务逻辑用例作为参数
func NewUserService(useCase biz.UserUseCase) *UserService {
	return &UserService{
		useCase: useCase,
	}
}

// SayHello 实现 UserService.SayHello 方法
func (s *UserService) SayHello(ctx context.Context, req *userv1.HelloRequest) (*userv1.HelloResponse, error) {
	log.Info("received SayHello request")
	panic("mock panic")

	// 调用业务逻辑层
	message, err := s.useCase.SayHello(ctx)
	if err != nil {
		log.Error("failed to say hello", zap.Error(err))
		return nil, err
	}

	log.Info("SayHello completed", zap.String("message", message))

	return &userv1.HelloResponse{
		Message: message,
	}, nil
}
