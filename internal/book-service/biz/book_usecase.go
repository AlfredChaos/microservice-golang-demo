package biz

import (
	"context"

	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

// BookUseCase 用户业务逻辑用例接口
type IBookUseCase interface {
	JustTellMe(ctx context.Context, name string) (string, error)
}

// BookUseCase Book业务逻辑用例实现
type BookUseCase struct {
}

// NewBookUseCase 创建新的Book业务逻辑用例
func NewBookUseCase() *BookUseCase {
	return &BookUseCase{}
}

func (uc *BookUseCase) JustTellMe(ctx context.Context, name string) (string, error) {
	log.WithContext(ctx).Info("processing JustTellMe request", zap.String("name", name))

	// 1. 生成Book-service的消息
	BookMessage := "World"
	if name != "" {
		BookMessage = "World " + name
	}

	return BookMessage, nil
}
