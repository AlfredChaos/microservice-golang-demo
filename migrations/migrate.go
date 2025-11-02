package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

// SharedDBMigrations 共享数据库的迁移文件
// 所有使用共享数据库的微服务的迁移文件都应放在这里
//
//go:embed shared-db/*.sql
var SharedDBMigrations embed.FS

// MigrateUp 执行数据库迁移（升级到最新版本）
func MigrateUp(db *sql.DB) error {
	if err := setupGoose(); err != nil {
		return err
	}

	if err := goose.Up(db, "shared-db"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// MigrateDown 回滚最后一次迁移
func MigrateDown(db *sql.DB) error {
	if err := setupGoose(); err != nil {
		return err
	}

	if err := goose.Down(db, "shared-db"); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	return nil
}

// MigrateStatus 查看迁移状态
func MigrateStatus(db *sql.DB) error {
	if err := setupGoose(); err != nil {
		return err
	}

	if err := goose.Status(db, "shared-db"); err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	return nil
}

// MigrateUpTo 迁移到指定版本（up-to）
func MigrateUpTo(db *sql.DB, version int64) error {
	if err := setupGoose(); err != nil {
		return err
	}

	if err := goose.UpTo(db, "shared-db", version); err != nil {
		return fmt.Errorf("failed to migrate up to version %d: %w", version, err)
	}

	return nil
}

// MigrateDownTo 回滚到指定版本（down-to）
func MigrateDownTo(db *sql.DB, version int64) error {
	if err := setupGoose(); err != nil {
		return err
	}

	if err := goose.DownTo(db, "shared-db", version); err != nil {
		return fmt.Errorf("failed to migrate down to version %d: %w", version, err)
	}

	return nil
}

// MigrateReset 重置数据库（回滚所有迁移）
func MigrateReset(db *sql.DB) error {
	if err := setupGoose(); err != nil {
		return err
	}

	if err := goose.Reset(db, "shared-db"); err != nil {
		return fmt.Errorf("failed to reset migrations: %w", err)
	}

	return nil
}

// GetCurrentVersion 获取当前数据库版本
func GetCurrentVersion(db *sql.DB) (int64, error) {
	if err := setupGoose(); err != nil {
		return 0, err
	}

	version, err := goose.GetDBVersion(db)
	if err != nil {
		return 0, fmt.Errorf("failed to get db version: %w", err)
	}

	return version, nil
}

// setupGoose 配置 goose
func setupGoose() error {
	goose.SetBaseFS(SharedDBMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	return nil
}
