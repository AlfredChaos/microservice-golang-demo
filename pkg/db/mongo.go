package db

import (
	"context"
	"fmt"
	"time"

	"github.com/alfredchaos/demo/pkg/log"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

// MongoConfig MongoDB 配置
type MongoConfig struct {
	URI                string `yaml:"uri" mapstructure:"uri"`                                   // MongoDB 连接 URI
	Database           string `yaml:"database" mapstructure:"database"`                         // 数据库名称
	MaxPoolSize        uint64 `yaml:"max_pool_size" mapstructure:"max_pool_size"`               // 最大连接池大小
	MinPoolSize        uint64 `yaml:"min_pool_size" mapstructure:"min_pool_size"`               // 最小连接池大小
	ConnectTimeout     int    `yaml:"connect_timeout" mapstructure:"connect_timeout"`           // 连接超时(秒)
	LogLevel           string `yaml:"log_level" mapstructure:"log_level"`                       // 日志级别 (silent, error, warn, info)
	SlowQueryThreshold int    `yaml:"slow_query_threshold" mapstructure:"slow_query_threshold"` // 慢查询阈值(毫秒)，默认200ms
	EnableDetailedLog  bool   `yaml:"enable_detailed_log" mapstructure:"enable_detailed_log"`   // 是否记录详细命令
}

// MongoClient MongoDB 客户端封装
type MongoClient struct {
	client   *mongo.Client
	database *mongo.Database
	config   *MongoConfig
}

// NewMongoClient 创建新的 MongoDB 客户端
func NewMongoClient(cfg *MongoConfig) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ConnectTimeout)*time.Second)
	defer cancel()

	// 设置客户端选项
	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize)

	// 配置命令监控（集成日志）
	if cfg.LogLevel != "" && cfg.LogLevel != "silent" {
		clientOptions.SetMonitor(newMongoCommandMonitor(cfg))
	}

	// 连接到 MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	// 验证连接
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	return &MongoClient{
		client:   client,
		database: client.Database(cfg.Database),
		config:   cfg,
	}, nil
}

// GetClient 获取原始 MongoDB 客户端
func (mc *MongoClient) GetClient() *mongo.Client {
	return mc.client
}

// GetDatabase 获取数据库实例
func (mc *MongoClient) GetDatabase() *mongo.Database {
	return mc.database
}

// GetCollection 获取集合实例
func (mc *MongoClient) GetCollection(name string) *mongo.Collection {
	return mc.database.Collection(name)
}

// Close 关闭 MongoDB 连接
func (mc *MongoClient) Close(ctx context.Context) error {
	if mc.client != nil {
		return mc.client.Disconnect(ctx)
	}
	return nil
}

// Ping 检查 MongoDB 连接是否正常
func (mc *MongoClient) Ping(ctx context.Context) error {
	return mc.client.Ping(ctx, readpref.Primary())
}

// WithTransaction 在事务中执行操作
// 提供事务支持,确保数据一致性
func (mc *MongoClient) WithTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error {
	session, err := mc.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	})

	return err
}

// MustNewMongoClient 创建 MongoDB 客户端,失败则 panic
// 适用于服务启动阶段,数据库连接失败应该直接终止程序
func MustNewMongoClient(cfg *MongoConfig) *MongoClient {
	client, err := NewMongoClient(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create mongodb client: %v", err))
	}
	return client
}

// ============================================================
// MongoDB 命令监控器（集成现有的 log 包）
// ============================================================

// newMongoCommandMonitor 创建 MongoDB 命令监控器
func newMongoCommandMonitor(cfg *MongoConfig) *event.CommandMonitor {
	slowThreshold := 200 * time.Millisecond // 默认 200ms
	if cfg.SlowQueryThreshold > 0 {
		slowThreshold = time.Duration(cfg.SlowQueryThreshold) * time.Millisecond
	}

	return &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			// 只在 info 级别且启用详细日志时记录命令开始
			if cfg.LogLevel == "info" && cfg.EnableDetailedLog {
				contextLogger := log.WithContext(ctx).WithOptions(zap.AddCallerSkip(1))
				contextLogger.Info("mongodb command started",
					zap.String("command", evt.CommandName),
					zap.String("database", evt.DatabaseName),
					zap.Int64("request_id", evt.RequestID),
				)
			}
		},

		Succeeded: func(ctx context.Context, evt *event.CommandSucceededEvent) {
			elapsed := time.Duration(evt.DurationNanos)
			contextLogger := log.WithContext(ctx).WithOptions(zap.AddCallerSkip(1))

			// 基础字段
			fields := []zap.Field{
				zap.String("command", evt.CommandName),
				zap.Float64("duration_ms", float64(elapsed.Nanoseconds())/1e6),
				zap.Int64("request_id", evt.RequestID),
			}

			// 根据配置决定是否记录详细信息
			if cfg.EnableDetailedLog {
				fields = append(fields, zap.Any("reply", evt.Reply))
			}

			// 慢查询检测
			if elapsed > slowThreshold {
				fields = append(fields,
					zap.Bool("is_slow_query", true),
					zap.Float64("threshold_ms", float64(slowThreshold.Milliseconds())),
				)
				contextLogger.Warn("mongodb slow query detected", fields...)
			} else if cfg.LogLevel == "info" {
				// 普通查询日志
				contextLogger.Info("mongodb command succeeded", fields...)
			}
		},

		Failed: func(ctx context.Context, evt *event.CommandFailedEvent) {
			elapsed := time.Duration(evt.DurationNanos)
			contextLogger := log.WithContext(ctx).WithOptions(zap.AddCallerSkip(1))

			// 错误日志
			contextLogger.Error("mongodb command failed",
				zap.String("command", evt.CommandName),
				zap.String("failure", evt.Failure),
				zap.Float64("duration_ms", float64(elapsed.Nanoseconds())/1e6),
				zap.Int64("request_id", evt.RequestID),
			)
		},
	}
}
