package biz

import "context"

// BookUseCase 图书业务逻辑用例接口
type BookUseCase interface {
	// GetBook 获取图书信息
	GetBook(ctx context.Context) (string, error)
}

// bookUseCase 图书业务逻辑用例实现
type bookUseCase struct {
}

// NewBookUseCase 创建新的图书业务逻辑用例
func NewBookUseCase() BookUseCase {
	return &bookUseCase{}
}

// GetBook 获取图书信息
// 返回 "World" 字符串
func (uc *bookUseCase) GetBook(ctx context.Context) (string, error) {
	return "World", nil
}
