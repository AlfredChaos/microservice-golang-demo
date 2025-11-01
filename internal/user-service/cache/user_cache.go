package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alfredchaos/demo/internal/user-service/domain"
	"github.com/alfredchaos/demo/pkg/cache"
	"github.com/go-redis/redis/v8"
)

const (
	// Redis Key 前缀
	userCacheKeyPrefix = "user:id:"
)

type UserCache interface {
	// SetUser 缓存用户信息（按 ID）
	// ttl: 缓存过期时间（秒），0 表示永不过期
	SetUser(ctx context.Context, user *domain.User, ttl int) error

	// GetUser 获取缓存的用户信息（按 ID）
	// 如果缓存不存在或已过期，返回 nil
	GetUser(ctx context.Context, userID string) (*domain.User, error)

	// DeleteUser 删除用户缓存（按 ID）
	DeleteUser(ctx context.Context, userID string) error
}

// userRedisCache Redis 缓存仓库实现
// 实现 UserCache 接口，提供基于 Redis 的快速缓存
type UserRedisCache struct {
	client *cache.RedisClient
}

// NewUserRedisCache 创建 Redis 缓存仓库
func NewUserRedisCache(cfg *cache.RedisConfig) *UserRedisCache {
	client := cache.MustNewRedisClient(cfg)
	return &UserRedisCache{
		client: client,
	}
}

// buildUserKey 构建用户 ID 缓存键
func buildUserKey(userID string) string {
	return userCacheKeyPrefix + userID
}

// serializeUser 序列化用户对象为 JSON
func serializeUser(user *domain.User) (string, error) {
	data, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("failed to serialize user: %w", err)
	}
	return string(data), nil
}

// deserializeUser 反序列化 JSON 为用户对象
func deserializeUser(data string) (*domain.User, error) {
	if data == "" {
		return nil, nil
	}

	var user domain.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, fmt.Errorf("failed to deserialize user: %w", err)
	}
	return &user, nil
}

// SetUser 缓存用户信息（按 ID）
func (r *UserRedisCache) SetUser(ctx context.Context, user *domain.User, ttl int) error {
	if user == nil || user.ID == "" {
		return fmt.Errorf("user or user ID is empty")
	}

	key := buildUserKey(user.ID)
	data, err := serializeUser(user)
	if err != nil {
		return err
	}

	expiration := time.Duration(0)
	if ttl > 0 {
		expiration = time.Duration(ttl) * time.Second
	}

	if err := r.client.Set(ctx, key, data, expiration); err != nil {
		return fmt.Errorf("failed to set user cache: %w", err)
	}

	return nil
}

// GetUser 获取缓存的用户信息（按 ID）
func (r *UserRedisCache) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is empty")
	}

	key := buildUserKey(userID)
	data, err := r.client.Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			// 缓存不存在
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user cache: %w", err)
	}

	return deserializeUser(data)
}

// DeleteUser 删除用户缓存（按 ID）
func (r *UserRedisCache) DeleteUser(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID is empty")
	}

	key := buildUserKey(userID)
	if err := r.client.Del(ctx, key); err != nil {
		return fmt.Errorf("failed to delete user cache: %w", err)
	}

	return nil
}
