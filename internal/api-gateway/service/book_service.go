package service

import (
	"context"
	"fmt"

	orderv1 "github.com/alfredchaos/demo/api/order/v1"
	"github.com/alfredchaos/demo/internal/api-gateway/domain"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

// bookService 图书服务实现
// 封装对 book-service 的 gRPC 调用
type bookService struct {
	baseService
	bookClient orderv1.BookServiceClient
}

// NewBookService 创建图书服务实例
// 注入 gRPC 客户端依赖
func NewBookService(bookClient orderv1.BookServiceClient) domain.IBookService {
	return &bookService{
		baseService: baseService{},
		bookClient:  bookClient,
	}
}

// GetBook 调用 book-service 的 GetBook 接口
func (s *bookService) GetBook(ctx context.Context) (string, error) {
	// 传递 trace ID 到 gRPC metadata
	ctx = s.withTraceID(ctx)

	// 调用 book-service
	resp, err := s.bookClient.GetBook(ctx, &orderv1.BookRequest{})
	if err != nil {
		log.WithContext(ctx).Error("failed to call book service", zap.Error(err))
		return "", fmt.Errorf("failed to call book service: %w", err)
	}

	log.WithContext(ctx).Info("book service GetBook success", zap.String("message", resp.Message))
	return resp.Message, nil
}
