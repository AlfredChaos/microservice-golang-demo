package psql

import (
	"context"
	"fmt"
	"time"

	"github.com/alfredchaos/demo/internal/user-service/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserPgPO 用户持久化对象（PostgreSQL）
// 负责与PostgreSQL交互的数据结构
type UserPgPO struct {
	ID        string    `gorm:"column:id;primaryKey"`
	Username  string    `gorm:"column:username;uniqueIndex;not null"`
	Email     string    `gorm:"column:email;not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName 指定表名
func (UserPgPO) TableName() string {
	return "users"
}

// BeforeCreate GORM 钩子：创建前自动设置时间戳
func (po *UserPgPO) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	if po.CreatedAt.IsZero() {
		po.CreatedAt = now
	}
	if po.UpdatedAt.IsZero() {
		po.UpdatedAt = now
	}
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动刷新 UpdatedAt
func (po *UserPgPO) BeforeUpdate(tx *gorm.DB) error {
	po.UpdatedAt = time.Now()
	return nil
}

// ToDomain 将持久化对象转换为领域对象
func (po *UserPgPO) ToDomain() *domain.User {
	return &domain.User{
		ID:        po.ID,
		Username:  po.Username,
		Email:     po.Email,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
}

// FromDomainUser 从领域对象创建持久化对象
func FromDomainUser(user *domain.User) *UserPgPO {
	return &UserPgPO{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// userPgRepository PostgreSQL仓库实现
type UserPgRepository struct {
	db *gorm.DB
}

// NewUserPgRepository 创建PostgreSQL用户仓库
func NewUserPgRepository(db *gorm.DB) *UserPgRepository {
	return &UserPgRepository{db: db}
}

// Create 创建用户
func (r *UserPgRepository) Create(ctx context.Context, user *domain.User) error {
	// 生成UUID作为ID
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// 验证用户数据
	if err := user.Validate(); err != nil {
		return fmt.Errorf("invalid user data: %w", err)
	}

	po := FromDomainUser(user)
	// GORM 会自动设置 CreatedAt 和 UpdatedAt
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// 将 GORM 自动生成的时间戳同步回领域对象
	user.CreatedAt = po.CreatedAt
	user.UpdatedAt = po.UpdatedAt

	return nil
}

// GetByID 根据ID获取用户
func (r *UserPgRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var po UserPgPO
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&po).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return po.ToDomain(), nil
}

// GetByUsername 根据用户名获取用户
func (r *UserPgRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var po UserPgPO
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&po).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return po.ToDomain(), nil
}

// Update 更新用户
func (r *UserPgRepository) Update(ctx context.Context, user *domain.User) error {
	if user.ID == "" {
		return fmt.Errorf("user id is required for update")
	}

	// 验证用户数据
	if err := user.Validate(); err != nil {
		return fmt.Errorf("invalid user data: %w", err)
	}

	po := FromDomainUser(user)
	result := r.db.WithContext(ctx).
		Model(&UserPgPO{}).
		Where("id = ?", user.ID).
		Select("username", "email", "updated_at").
		Updates(po)

	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	// 同步 Hook 设置的时间戳到领域对象
	user.UpdatedAt = po.UpdatedAt

	return nil
}

// Delete 删除用户
func (r *UserPgRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("user id is required for delete")
	}

	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&UserPgPO{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// List 列出用户
func (r *UserPgRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	var pos []UserPgPO

	query := r.db.WithContext(ctx)

	// 设置分页参数
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	// 按创建时间倒序排列
	if err := query.Order("created_at DESC").Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// 转换为领域对象
	users := make([]*domain.User, 0, len(pos))
	for _, po := range pos {
		users = append(users, po.ToDomain())
	}

	return users, nil
}
