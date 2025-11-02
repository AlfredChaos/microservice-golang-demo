package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/alfredchaos/demo/internal/user-service/conf"
	"github.com/alfredchaos/demo/migrations"
	"github.com/alfredchaos/demo/pkg/config"
	"github.com/alfredchaos/demo/pkg/db"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

func main() {
	// 解析命令行参数
	var (
		command = flag.String("cmd", "up", "Migration command: up, up-to, down, down-to, status, version, reset")
		version = flag.Int64("version", 0, "Target version (for up-to/down-to commands)")
		cfgPath = flag.String("config", "configs/user-service.yaml", "Configuration file path")
	)
	flag.Parse()

	// 加载配置
	var cfg conf.Config
	if err := config.LoadConfigFromPath(*cfgPath, &cfg); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	log.MustInitLogger(&cfg.Log, cfg.Server.Name)
	defer log.Sync()

	log.Info("Starting database migration", zap.String("command", *command))

	// 创建数据库客户端（不执行迁移）
	client, err := db.NewPostgresClient(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to create database client", zap.Error(err))
	}
	defer client.Close()

	// 获取底层的 sql.DB 对象
	sqlDB, err := client.GetDB().DB()
	if err != nil {
		log.Fatal("Failed to get sql.DB", zap.Error(err))
	}

	// 执行迁移命令
	switch *command {
	case "up":
		if err := migrations.MigrateUp(sqlDB); err != nil {
			log.Fatal("Failed to execute migration", zap.Error(err))
		}
		log.Info("Migration completed successfully")

	case "down":
		if err := migrations.MigrateDown(sqlDB); err != nil {
			log.Fatal("Failed to rollback migration", zap.Error(err))
		}
		log.Info("Rollback completed successfully")

	case "status":
		if err := migrations.MigrateStatus(sqlDB); err != nil {
			log.Fatal("Failed to query migration status", zap.Error(err))
		}

	case "up-to":
		if *version == 0 {
			log.Fatal("up-to command requires -version parameter")
		}
		if err := migrations.MigrateUpTo(sqlDB, *version); err != nil {
			log.Fatal("Failed to migrate up to version", zap.Error(err))
		}
		log.Info("Migrated up to version successfully", zap.Int64("version", *version))

	case "down-to":
		if *version == 0 {
			log.Fatal("down-to command requires -version parameter")
		}
		if err := migrations.MigrateDownTo(sqlDB, *version); err != nil {
			log.Fatal("Failed to migrate down to version", zap.Error(err))
		}
		log.Info("Migrated down to version successfully", zap.Int64("version", *version))

	case "version":
		// 查询当前版本
		currentVersion, err := migrations.GetCurrentVersion(sqlDB)
		if err != nil {
			log.Fatal("Failed to get current version", zap.Error(err))
		}
		log.Info("Current database version", zap.Int64("version", currentVersion))

	case "reset":
		log.Warn("WARNING: About to reset database (will delete all data)")
		if err := migrations.MigrateReset(sqlDB); err != nil {
			log.Fatal("Failed to reset database", zap.Error(err))
		}
		log.Info("Database reset successfully")

	default:
		log.Fatal(fmt.Sprintf("Unknown command: %s", *command))
	}

	log.Info("Migration operation completed")
}
