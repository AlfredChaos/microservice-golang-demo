package biz

import (
	"context"

	"github.com/alfredchaos/demo/internal/user-service/data"
	"github.com/alfredchaos/demo/internal/user-service/domain"
)

// UserUseCase 用户业务逻辑用例接口
// 定义业务层的抽象接口,遵循依赖倒置原则
type UserUseCase interface {
	// SayHello 返回问候语
	SayHello(ctx context.Context) (string, error)
	
	// CreateUser 创建用户
	CreateUser(ctx context.Context, username, email string) (*domain.User, error)
	
	// GetUser 获取用户
	GetUser(ctx context.Context, id string) (*domain.User, error)
}

// userUseCase 用户业务逻辑用例实现
type userUseCase struct {
	userRepo data.UserRepository
}

// NewUserUseCase 创建新的用户业务逻辑用例
// 使用依赖注入,接收仓库接口作为参数
func NewUserUseCase(userRepo data.UserRepository) UserUseCase {
	return &userUseCase{
		userRepo: userRepo,
	}
}

// SayHello 返回问候语
// 这是一个简单的演示方法,实际项目中会包含更复杂的业务逻辑
func (uc *userUseCase) SayHello(ctx context.Context) (string, error) {
	// 这里可以添加业务逻辑,如记录日志、统计调用次数等
	return "Hello", nil
}

// CreateUser 创建用户
func (uc *userUseCase) CreateUser(ctx context.Context, username, email string) (*domain.User, error) {
	// 创建用户领域对象
	user := domain.NewUser(username, email)
	
	// 验证用户数据
	if err := user.Validate(); err != nil {
		return nil, err
	}
	
	// 检查用户是否已存在
	existingUser, err := uc.userRepo.GetByUsername(ctx, username)
	if err != nil && err != domain.ErrUserNotFound {
		return nil, err
	}
	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}
	
	// 保存用户
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	
	return user, nil
}

// GetUser 获取用户
func (uc *userUseCase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}
