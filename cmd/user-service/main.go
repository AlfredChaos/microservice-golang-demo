package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/alfredchaos/demo/internal/user-service/biz"
	"github.com/alfredchaos/demo/internal/user-service/conf"
	"github.com/alfredchaos/demo/internal/user-service/data"
	"github.com/alfredchaos/demo/internal/user-service/server"
	"github.com/alfredchaos/demo/internal/user-service/service"
	"github.com/alfredchaos/demo/pkg/cache"
	"github.com/alfredchaos/demo/pkg/config"
	"github.com/alfredchaos/demo/pkg/db"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	var cfg conf.Config
	config.MustLoadConfig("user-service", &cfg)
	
	// 初始化日志
	log.MustInitLogger(&cfg.Log, cfg.Server.Name)
	defer log.Sync()
	
	log.Info("starting user-service", zap.String("name", cfg.Server.Name))
	
	// 初始化 MongoDB 客户端
	mongoClient := db.MustNewMongoClient(&cfg.MongoDB)
	defer func() {
		if err := mongoClient.Close(context.Background()); err != nil {
			log.Error("failed to close mongodb client", zap.Error(err))
		}
	}()
	
	// 初始化 Redis 客户端
	redisClient := cache.MustNewRedisClient(&cfg.Redis)
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Error("failed to close redis client", zap.Error(err))
		}
	}()
	
	// 初始化数据访问层
	dataLayer, err := data.NewData(mongoClient, redisClient)
	if err != nil {
		log.Fatal("failed to initialize data layer", zap.Error(err))
	}
	defer func() {
		if err := dataLayer.Close(context.Background()); err != nil {
			log.Error("failed to close data layer", zap.Error(err))
		}
	}()
	
	// 初始化业务逻辑层
	userUseCase := biz.NewUserUseCase(dataLayer.UserRepo)
	
	// 初始化服务层
	userService := service.NewUserService(userUseCase)
	
	// 初始化 gRPC 服务器
	grpcServer := server.NewGRPCServer(&cfg.Server, userService)
	
	// 启动服务器
	go func() {
		if err := grpcServer.Start(); err != nil {
			log.Fatal("failed to start grpc server", zap.Error(err))
		}
	}()
	
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Info("shutting down user-service")
	grpcServer.Stop()
	log.Info("user-service stopped")
}
