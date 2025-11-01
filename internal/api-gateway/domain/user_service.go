package domain

import (
	"context"
)

// IUserService 用户服务领域接口
// 定义用户相关的业务能力
type IUserService interface {
	// SayHello 问候接口
	// 返回问候消息
	SayHello(ctx context.Context) (string, error)
}
