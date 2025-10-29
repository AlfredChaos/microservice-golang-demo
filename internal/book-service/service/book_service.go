package service

import (
	"context"

	orderv1 "github.com/alfredchaos/demo/api/order/v1"
	"github.com/alfredchaos/demo/internal/book-service/biz"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

// BookService gRPC 服务实现
type BookService struct {
	orderv1.UnimplementedBookServiceServer
	useCase biz.BookUseCase
}

// NewBookService 创建新的图书服务
func NewBookService(useCase biz.BookUseCase) *BookService {
	return &BookService{
		useCase: useCase,
	}
}

// GetBook 实现 BookService.GetBook 方法
func (s *BookService) GetBook(ctx context.Context, req *orderv1.BookRequest) (*orderv1.BookResponse, error) {
	log.WithContext(ctx).Info("received GetBook request")

	// 调用业务逻辑层
	message, err := s.useCase.GetBook(ctx)
	if err != nil {
		log.WithContext(ctx).Error("failed to get book", zap.Error(err))
		return nil, err
	}

	log.WithContext(ctx).Info("GetBook completed", zap.String("message", message))

	return &orderv1.BookResponse{
		Message: message,
	}, nil
}
