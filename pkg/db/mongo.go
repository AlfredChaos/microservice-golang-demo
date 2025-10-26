package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoConfig MongoDB 配置
type MongoConfig struct {
	URI            string `yaml:"uri" mapstructure:"uri"`                           // MongoDB 连接 URI
	Database       string `yaml:"database" mapstructure:"database"`                 // 数据库名称
	MaxPoolSize    uint64 `yaml:"max_pool_size" mapstructure:"max_pool_size"`       // 最大连接池大小
	MinPoolSize    uint64 `yaml:"min_pool_size" mapstructure:"min_pool_size"`       // 最小连接池大小
	ConnectTimeout int    `yaml:"connect_timeout" mapstructure:"connect_timeout"`   // 连接超时(秒)
}

// MongoClient MongoDB 客户端封装
type MongoClient struct {
	client   *mongo.Client
	database *mongo.Database
	config   *MongoConfig
}

// NewMongoClient 创建新的 MongoDB 客户端
// 使用工厂模式创建客户端实例,便于测试和依赖注入
func NewMongoClient(cfg *MongoConfig) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ConnectTimeout)*time.Second)
	defer cancel()
	
	// 设置客户端选项
	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize)
	
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
