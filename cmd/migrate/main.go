package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/alfredchaos/demo/internal/user-service/conf"
	"github.com/alfredchaos/demo/internal/user-service/repository/psql"
	"github.com/alfredchaos/demo/pkg/config"
	"github.com/alfredchaos/demo/pkg/db"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

func main() {
	// 解析命令行参数
	var (
		command = flag.String("cmd", "up", "迁移命令: up, down, status, version, reset")
		version = flag.Int64("version", 0, "迁移到指定版本（仅用于 version 命令）")
		cfgPath = flag.String("config", "configs/user-service.yaml", "配置文件路径")
	)
	flag.Parse()

	// 加载配置
	var cfg conf.Config
	if err := config.LoadConfigFromPath(*cfgPath, &cfg); err != nil {
		fmt.Printf("❌ 加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	log.MustInitLogger(&cfg.Log, cfg.Server.Name)
	defer log.Sync()

	log.Info("🚀 开始数据库迁移", zap.String("command", *command))

	// 创建数据库客户端（不执行迁移）
	client, err := db.NewPostgresClient(&cfg.Database)
	if err != nil {
		log.Fatal("❌ 创建数据库客户端失败", zap.Error(err))
	}
	defer client.Close()

	// 执行迁移命令
	switch *command {
	case "up":
		if err := psql.MigrateUp(client); err != nil {
			log.Fatal("❌ 执行迁移失败", zap.Error(err))
		}
		log.Info("✅ 迁移成功完成")

	case "down":
		if err := psql.MigrateDown(client); err != nil {
			log.Fatal("❌ 回滚迁移失败", zap.Error(err))
		}
		log.Info("✅ 回滚成功完成")

	case "status":
		if err := psql.MigrateStatus(client); err != nil {
			log.Fatal("❌ 查询迁移状态失败", zap.Error(err))
		}

	case "version":
		if *version == 0 {
			// 查询当前版本
			currentVersion, err := psql.GetCurrentVersion(client)
			if err != nil {
				log.Fatal("❌ 获取当前版本失败", zap.Error(err))
			}
			log.Info("📌 当前数据库版本", zap.Int64("version", currentVersion))
		} else {
			// 迁移到指定版本
			if err := psql.MigrateVersion(client, *version); err != nil {
				log.Fatal("❌ 迁移到指定版本失败", zap.Error(err))
			}
			log.Info("✅ 迁移到指定版本成功", zap.Int64("version", *version))
		}

	case "reset":
		log.Warn("⚠️  警告：即将重置数据库（删除所有数据）")
		if err := psql.MigrateReset(client); err != nil {
			log.Fatal("❌ 重置数据库失败", zap.Error(err))
		}
		log.Info("✅ 数据库重置成功")

	default:
		log.Fatal(fmt.Sprintf("❌ 未知命令: %s", *command))
	}

	log.Info("🎉 迁移操作完成")
}
