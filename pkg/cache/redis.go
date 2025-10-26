package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr         string `yaml:"addr" mapstructure:"addr"`                   // Redis 地址
	Password     string `yaml:"password" mapstructure:"password"`           // 密码
	DB           int    `yaml:"db" mapstructure:"db"`                       // 数据库编号
	PoolSize     int    `yaml:"pool_size" mapstructure:"pool_size"`         // 连接池大小
	MinIdleConns int    `yaml:"min_idle_conns" mapstructure:"min_idle_conns"` // 最小空闲连接数
	DialTimeout  int    `yaml:"dial_timeout" mapstructure:"dial_timeout"`   // 连接超时(秒)
	ReadTimeout  int    `yaml:"read_timeout" mapstructure:"read_timeout"`   // 读超时(秒)
	WriteTimeout int    `yaml:"write_timeout" mapstructure:"write_timeout"` // 写超时(秒)
}

// RedisClient Redis 客户端封装
type RedisClient struct {
	client *redis.Client
	config *RedisConfig
}

// NewRedisClient 创建新的 Redis 客户端
// 使用工厂模式创建客户端实例
func NewRedisClient(cfg *RedisConfig) (*RedisClient, error) {
	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
	})
	
	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	
	return &RedisClient{
		client: client,
		config: cfg,
	}, nil
}

// GetClient 获取原始 Redis 客户端
func (rc *RedisClient) GetClient() *redis.Client {
	return rc.client
}

// Set 设置键值对
func (rc *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return rc.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取键对应的值
func (rc *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return rc.client.Get(ctx, key).Result()
}

// Del 删除键
func (rc *RedisClient) Del(ctx context.Context, keys ...string) error {
	return rc.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (rc *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return rc.client.Exists(ctx, keys...).Result()
}

// Expire 设置键的过期时间
func (rc *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return rc.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取键的剩余生存时间
func (rc *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return rc.client.TTL(ctx, key).Result()
}

// Incr 将键的值加1
func (rc *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return rc.client.Incr(ctx, key).Result()
}

// Decr 将键的值减1
func (rc *RedisClient) Decr(ctx context.Context, key string) (int64, error) {
	return rc.client.Decr(ctx, key).Result()
}

// Close 关闭 Redis 连接
func (rc *RedisClient) Close() error {
	if rc.client != nil {
		return rc.client.Close()
	}
	return nil
}

// Ping 检查 Redis 连接是否正常
func (rc *RedisClient) Ping(ctx context.Context) error {
	return rc.client.Ping(ctx).Err()
}

// MustNewRedisClient 创建 Redis 客户端,失败则 panic
// 适用于服务启动阶段
func MustNewRedisClient(cfg *RedisConfig) *RedisClient {
	client, err := NewRedisClient(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create redis client: %v", err))
	}
	return client
}
