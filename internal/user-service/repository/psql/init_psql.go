package psql

import (
	"fmt"
	"os"

	"github.com/alfredchaos/demo/pkg/db"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// InitPostgresClient 初始化 PostgreSQL 客户端
func InitPostgresClient(cfg *db.PostgresConfig) (*db.PostgresClient, error) {
	// 检查是否启用 PostgreSQL
	if !cfg.Enabled {
		return nil, fmt.Errorf("postgresql is not enabled in config")
	}

	// 设置默认值
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "warn"
	}

	// 创建 PostgreSQL 客户端
	client, err := db.NewPostgresClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres client: %w", err)
	}

	// 根据环境选择迁移方式
	env := os.Getenv("ENV")
	if env == "development" {
		// 开发环境：使用 AutoMigrate 快速迭代
		log.Info("using GORM AutoMigrate in development environment")
		if err := migrateModels(client.GetDB()); err != nil {
			client.Close()
			return nil, fmt.Errorf("failed to auto migrate models: %w", err)
		}
	} else {
		// 生产/测试环境：使用 Goose 版本化迁移
		log.Info("using Goose migrations", zap.String("env", env))
		if err := MigrateUp(client); err != nil {
			client.Close()
			return nil, fmt.Errorf("failed to run goose migrations: %w", err)
		}
	}

	return client, nil
}

// migrateModels 执行数据模型自动迁移
func migrateModels(db *gorm.DB) error {
	// 注册需要迁移的模型
	models := []interface{}{
		&UserPgPO{},
		// 在此添加更多需要迁移的模型
	}

	// 执行自动迁移
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto migrate failed: %w", err)
	}

	return nil
}

// MustInitPostgresClient 初始化 PostgreSQL 客户端，失败则 panic
func MustInitPostgresClient(cfg *db.PostgresConfig) *db.PostgresClient {
	client, err := InitPostgresClient(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to init postgres client: %v", err))
	}
	return client
}

