package domain

import (
	"context"
)

// IBookService 图书服务领域接口
// 定义图书相关的业务能力
type IBookService interface {
	// GetBook 获取图书信息
	// 返回图书信息消息
	GetBook(ctx context.Context) (string, error)
}
