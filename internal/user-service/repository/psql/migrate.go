package psql

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/alfredchaos/demo/pkg/db"
	"github.com/alfredchaos/demo/pkg/log"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// MigrateUp 执行数据库迁移（升级到最新版本）
func MigrateUp(client *db.PostgresClient) error {
	sqlDB, err := client.GetDB().DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(sqlDB, "migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info("database migrations completed successfully")
	return nil
}

// MigrateDown 回滚最后一次迁移
func MigrateDown(client *db.PostgresClient) error {
	sqlDB, err := client.GetDB().DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Down(sqlDB, "migrations"); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	log.Info("database migration rolled back successfully")
	return nil
}

// MigrateStatus 查看迁移状态
func MigrateStatus(client *db.PostgresClient) error {
	sqlDB, err := client.GetDB().DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Status(sqlDB, "migrations"); err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	return nil
}

// MigrateVersion 迁移到指定版本
func MigrateVersion(client *db.PostgresClient, version int64) error {
	sqlDB, err := client.GetDB().DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.UpTo(sqlDB, "migrations", version); err != nil {
		return fmt.Errorf("failed to migrate to version %d: %w", version, err)
	}

	log.Info("database migrated to version", zap.Int64("version", version))
	return nil
}

// MigrateReset 重置数据库（回滚所有迁移）
func MigrateReset(client *db.PostgresClient) error {
	sqlDB, err := client.GetDB().DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Reset(sqlDB, "migrations"); err != nil {
		return fmt.Errorf("failed to reset migrations: %w", err)
	}

	log.Info("database migrations reset successfully")
	return nil
}

// GetCurrentVersion 获取当前数据库版本
func GetCurrentVersion(client *db.PostgresClient) (int64, error) {
	sqlDB, err := client.GetDB().DB()
	if err != nil {
		return 0, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return 0, fmt.Errorf("failed to set goose dialect: %w", err)
	}

	version, err := goose.GetDBVersion(sqlDB)
	if err != nil {
		return 0, fmt.Errorf("failed to get db version: %w", err)
	}

	return version, nil
}
