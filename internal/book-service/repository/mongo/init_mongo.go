package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/alfredchaos/demo/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config MongoDB 配置
// 用于从配置文件映射 mongodb 配置段
type Config struct {
	URI            string `yaml:"uri" mapstructure:"uri"`
	Database       string `yaml:"database" mapstructure:"database"`
	MaxPoolSize    uint64 `yaml:"max_pool_size" mapstructure:"max_pool_size"`
	MinPoolSize    uint64 `yaml:"min_pool_size" mapstructure:"min_pool_size"`
	ConnectTimeout int    `yaml:"connect_timeout" mapstructure:"connect_timeout"`
}

// InitMongoClient 初始化 MongoDB 客户端
// 1. 加载配置并创建 MongoDB 客户端
// 2. 可选：创建必要的索引
func InitMongoClient(cfg *Config) (*db.MongoClient, error) {
	// 验证配置
	if cfg.URI == "" {
		return nil, fmt.Errorf("mongodb uri is required")
	}
	if cfg.Database == "" {
		return nil, fmt.Errorf("mongodb database name is required")
	}

	// 构建 MongoConfig
	mongoCfg := &db.MongoConfig{
		URI:            cfg.URI,
		Database:       cfg.Database,
		MaxPoolSize:    cfg.MaxPoolSize,
		MinPoolSize:    cfg.MinPoolSize,
		ConnectTimeout: cfg.ConnectTimeout,
	}

	// 设置默认值
	if mongoCfg.MaxPoolSize == 0 {
		mongoCfg.MaxPoolSize = 100
	}
	if mongoCfg.MinPoolSize == 0 {
		mongoCfg.MinPoolSize = 10
	}
	if mongoCfg.ConnectTimeout == 0 {
		mongoCfg.ConnectTimeout = 10
	}

	// 创建 MongoDB 客户端
	client, err := db.NewMongoClient(mongoCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create mongodb client: %w", err)
	}

	// 创建索引（可选）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := createIndexes(ctx, client); err != nil {
		// 索引创建失败时关闭客户端
		client.Close(context.Background())
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return client, nil
}

// createIndexes 创建必要的索引
func createIndexes(ctx context.Context, client *db.MongoClient) error {
	// 获取用户集合
	collection := client.GetCollection(CollectionBooks)

	// 定义索引
	indexes := []mongo.IndexModel{
		{
			// bookname 唯一索引
			Keys:    bson.D{{Key: "bookname", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("idx_bookname"),
		},
		{
			// email 索引（允许重复，用于查询优化）
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetName("idx_email"),
		},
		{
			// created_at 索引（用于排序和范围查询）
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_created_at"),
		},
	}

	// 创建索引
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// MustInitMongoClient 初始化 MongoDB 客户端，失败则 panic
// 适用于服务启动阶段，数据库初始化失败应该直接终止程序
func MustInitMongoClient(cfg *Config) *db.MongoClient {
	client, err := InitMongoClient(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to init mongodb client: %v", err))
	}
	return client
}

// FromMongoConfig 从 db.MongoConfig 结构创建 mongo.Config
// 提供配置转换函数，简化配置映射
func FromMongoConfig(mc *db.MongoConfig) *Config {
	return &Config{
		URI:            mc.URI,
		Database:       mc.Database,
		MaxPoolSize:    mc.MaxPoolSize,
		MinPoolSize:    mc.MinPoolSize,
		ConnectTimeout: mc.ConnectTimeout,
	}
}

// ToMongoConfig 转换为 db.MongoConfig
// 用于需要传递给 pkg/db 的场景
func (c *Config) ToMongoConfig() *db.MongoConfig {
	return &db.MongoConfig{
		URI:            c.URI,
		Database:       c.Database,
		MaxPoolSize:    c.MaxPoolSize,
		MinPoolSize:    c.MinPoolSize,
		ConnectTimeout: c.ConnectTimeout,
	}
}

// NewConfig 创建新的 MongoDB 配置
// 提供便捷的配置创建函数
func NewConfig(uri, database string, maxPoolSize, minPoolSize uint64, connectTimeout int) *Config {
	return &Config{
		URI:            uri,
		Database:       database,
		MaxPoolSize:    maxPoolSize,
		MinPoolSize:    minPoolSize,
		ConnectTimeout: connectTimeout,
	}
}
