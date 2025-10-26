package data

import (
	"context"

	"github.com/alfredchaos/demo/internal/user-service/domain"
)

// UserRepository 用户仓库接口
// 定义数据访问的抽象接口,遵循依赖倒置原则
// 业务层依赖接口而非具体实现,便于测试和替换存储方案
type UserRepository interface {
	// Create 创建用户
	Create(ctx context.Context, user *domain.User) error
	
	// GetByID 根据ID获取用户
	GetByID(ctx context.Context, id string) (*domain.User, error)
	
	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	
	// Update 更新用户
	Update(ctx context.Context, user *domain.User) error
	
	// Delete 删除用户
	Delete(ctx context.Context, id string) error
	
	// List 列出用户
	List(ctx context.Context, offset, limit int) ([]*domain.User, error)
}
