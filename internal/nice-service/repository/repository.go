package repository

import (
	"context"
	"fmt"

	"github.com/alfredchaos/demo/pkg/db"
)

// Data 数据访问层容器
type Data struct {
	pgClient    *db.PostgresClient
	mongoClient *db.MongoClient
}

// NewData 创建新的数据访问层实例
func NewData(
	pgClient *db.PostgresClient,
	mongoClient *db.MongoClient,
) *Data {
	return &Data{
		pgClient:    pgClient,
		mongoClient: mongoClient,
	}
}

// Close 关闭所有数据连接
func (d *Data) Close(ctx context.Context) error {
	var errs []error

	// 关闭 MongoDB
	if err := d.closeMongo(ctx); err != nil {
		errs = append(errs, err)
	}

	// 关闭 PostgreSQL
	if err := d.closePostgres(); err != nil {
		errs = append(errs, err)
	}

	// 汇总错误
	if len(errs) > 0 {
		return fmt.Errorf("failed to close data layer: %v", errs)
	}

	return nil
}

// closePostgres 关闭 PostgreSQL 连接
func (d *Data) closePostgres() error {
	if d.pgClient != nil {
		if err := d.pgClient.Close(); err != nil {
			return fmt.Errorf("failed to close postgres: %w", err)
		}
	}
	return nil
}

// closeMongo 关闭 MongoDB 连接
func (d *Data) closeMongo(ctx context.Context) error {
	if d.mongoClient != nil {
		if err := d.mongoClient.Close(ctx); err != nil {
			return fmt.Errorf("failed to close mongodb: %w", err)
		}
	}
	return nil
}

// GetPostgresClient 获取 PostgreSQL 客户端
func (d *Data) GetPostgresClient() *db.PostgresClient {
	return d.pgClient
}

// GetMongoClient 获取 MongoDB 客户端
func (d *Data) GetMongoClient() *db.MongoClient {
	return d.mongoClient
}
