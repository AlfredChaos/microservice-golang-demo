package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/alfredchaos/demo/pkg/log"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr              string `yaml:"addr" mapstructure:"addr"`                               // Redis 地址
	Password          string `yaml:"password" mapstructure:"password"`                       // 密码
	DB                int    `yaml:"db" mapstructure:"db"`                                   // 数据库编号
	PoolSize          int    `yaml:"pool_size" mapstructure:"pool_size"`                     // 连接池大小
	MinIdleConns      int    `yaml:"min_idle_conns" mapstructure:"min_idle_conns"`           // 最小空闲连接数
	DialTimeout       int    `yaml:"dial_timeout" mapstructure:"dial_timeout"`               // 连接超时(秒)
	ReadTimeout       int    `yaml:"read_timeout" mapstructure:"read_timeout"`               // 读超时(秒)
	WriteTimeout      int    `yaml:"write_timeout" mapstructure:"write_timeout"`             // 写超时(秒)
	LogLevel          string `yaml:"log_level" mapstructure:"log_level"`                     // 日志级别 (silent, error, warn, info)
	SlowOpThreshold   int    `yaml:"slow_op_threshold" mapstructure:"slow_op_threshold"`     // 慢操作阈值(毫秒)，默认100ms
	EnableDetailedLog bool   `yaml:"enable_detailed_log" mapstructure:"enable_detailed_log"` // 是否记录详细命令
}

// RedisClient Redis 客户端封装
type RedisClient struct {
	client *redis.Client
	config *RedisConfig
}

// NewRedisClient 创建新的 Redis 客户端
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

	// 添加日志 Hook
	if cfg.LogLevel != "" && cfg.LogLevel != "silent" {
		client.AddHook(newRedisLogHook(cfg))
	}

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
func MustNewRedisClient(cfg *RedisConfig) *RedisClient {
	client, err := NewRedisClient(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create redis client: %v", err))
	}
	return client
}

// ============================================================
// Redis 日志 Hook（集成现有的 log 包）
// ============================================================

// redisLogHook Redis 日志钩子
type redisLogHook struct {
	logLevel          string
	slowOpThreshold   time.Duration
	enableDetailedLog bool
}

// newRedisLogHook 创建 Redis 日志钩子
func newRedisLogHook(cfg *RedisConfig) *redisLogHook {
	slowOpThreshold := 100 * time.Millisecond // 默认 100ms
	if cfg.SlowOpThreshold > 0 {
		slowOpThreshold = time.Duration(cfg.SlowOpThreshold) * time.Millisecond
	}

	return &redisLogHook{
		logLevel:          cfg.LogLevel,
		slowOpThreshold:   slowOpThreshold,
		enableDetailedLog: cfg.EnableDetailedLog,
	}
}

func (h *redisLogHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	// 在命令执行前不记录日志，避免过多日志
	return ctx, nil
}

func (h *redisLogHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	// 计算执行时间
	return h.logCommand(ctx, cmd, 0)
}

// BeforeProcessPipeline 在 Pipeline 执行前调用
func (h *redisLogHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

// AfterProcessPipeline 在 Pipeline 执行后调用
func (h *redisLogHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	for _, cmd := range cmds {
		if err := h.logCommand(ctx, cmd, 0); err != nil {
			return err
		}
	}
	return nil
}

// logCommand 记录命令执行日志
func (h *redisLogHook) logCommand(ctx context.Context, cmd redis.Cmder, duration time.Duration) error {
	contextLogger := log.WithContext(ctx).WithOptions(zap.AddCallerSkip(2))

	fields := []zap.Field{
		zap.String("command", cmd.Name()),
	}

	if h.enableDetailedLog {
		fields = append(fields, zap.Any("args", cmd.Args()))
	}

	// 检查命令是否有错误（排除 redis.Nil）
	err := cmd.Err()
	if err != nil && err != redis.Nil {
		fields = append(fields, zap.Error(err))
		contextLogger.Error("redis command failed", fields...)
		return nil
	}

	// 记录缓存命中/未命中
	if cmd.Name() == "get" || cmd.Name() == "mget" {
		fields = append(fields, zap.Bool("cache_hit", err != redis.Nil))
	}

	if h.logLevel == "info" {
		contextLogger.Info("redis command executed", fields...)
	}

	return nil
}
