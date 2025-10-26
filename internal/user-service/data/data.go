package data

import (
	"context"

	"github.com/alfredchaos/demo/pkg/cache"
	"github.com/alfredchaos/demo/pkg/db"
)

// Data 数据访问层容器
// 负责初始化和管理所有数据连接和仓库
// 使用依赖注入模式,便于测试和替换实现
type Data struct {
	// 数据库和缓存客户端
	mongoClient *db.MongoClient
	redisClient *cache.RedisClient
	
	// 仓库实例
	UserRepo UserRepository
}

// NewData 创建新的数据访问层实例
// 使用选项模式(Functional Options Pattern)提供灵活的初始化方式
func NewData(mongoClient *db.MongoClient, redisClient *cache.RedisClient) (*Data, error) {
	d := &Data{
		mongoClient: mongoClient,
		redisClient: redisClient,
	}
	
	// 初始化仓库
	d.UserRepo = NewUserMongoRepository(mongoClient)
	
	return d, nil
}

// Close 关闭所有数据连接
func (d *Data) Close(ctx context.Context) error {
	if d.mongoClient != nil {
		if err := d.mongoClient.Close(ctx); err != nil {
			return err
		}
	}
	
	if d.redisClient != nil {
		if err := d.redisClient.Close(); err != nil {
			return err
		}
	}
	
	return nil
}

// GetMongoClient 获取 MongoDB 客户端
func (d *Data) GetMongoClient() *db.MongoClient {
	return d.mongoClient
}

// GetRedisClient 获取 Redis 客户端
func (d *Data) GetRedisClient() *cache.RedisClient {
	return d.redisClient
}
