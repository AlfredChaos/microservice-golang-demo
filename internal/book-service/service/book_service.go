package service

import (
	"context"

	bookv1 "github.com/alfredchaos/demo/api/book/v1"
	"github.com/alfredchaos/demo/internal/book-service/biz"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

// bookService gRPC服务实现
type BookService struct {
	bookv1.UnimplementedBookServiceServer
	useCase *biz.BookUseCase
}

// NewBookService 创建新的用户服务
func NewBookService(useCase *biz.BookUseCase) *BookService {
	return &BookService{
		useCase: useCase,
	}
}

// SayHello 实现bookService.SayHello方法
func (s *BookService) JustTellMe(ctx context.Context, req *bookv1.TellMeRequest) (*bookv1.TellMeResponse, error) {
	log.WithContext(ctx).Info("received SayHello request")

	// 调用业务逻辑层
	message, err := s.useCase.JustTellMe(ctx, "")
	if err != nil {
		log.WithContext(ctx).Error("failed to say hello", zap.Error(err))
		return nil, err
	}

	log.WithContext(ctx).Info("SayHello completed", zap.String("message", message))

	// 构造gRPC响应
	return &bookv1.TellMeResponse{
		Message: message,
	}, nil
}
