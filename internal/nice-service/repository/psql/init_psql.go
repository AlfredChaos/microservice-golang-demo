package psql

import (
	"fmt"

	"github.com/alfredchaos/demo/pkg/db"
	"github.com/alfredchaos/demo/pkg/log"
)

// InitPostgresClient 初始化 PostgreSQL 客户端
// 注意：不再执行数据库迁移！
// 迁移应该通过独立的 cmd/migrate 工具执行
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

	log.Info("PostgreSQL client initialized successfully")
	log.Info("Note: Database migrations should be run separately using 'make migrate-up' or 'cmd/migrate'")

	return client, nil
}

// MustInitPostgresClient 初始化 PostgreSQL 客户端，失败则 panic
func MustInitPostgresClient(cfg *db.PostgresConfig) *db.PostgresClient {
	client, err := InitPostgresClient(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to init postgres client: %v", err))
	}
	return client
}

