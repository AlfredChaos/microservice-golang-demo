package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/alfredchaos/demo/internal/book-service/biz"
	"github.com/alfredchaos/demo/internal/book-service/conf"
	"github.com/alfredchaos/demo/internal/book-service/server"
	"github.com/alfredchaos/demo/internal/book-service/service"
	"github.com/alfredchaos/demo/pkg/config"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	var cfg conf.Config
	config.MustLoadConfig("book-service", &cfg)
	
	// 初始化日志
	log.MustInitLogger(&cfg.Log, cfg.Server.Name)
	defer log.Sync()
	
	log.Info("starting book-service", zap.String("name", cfg.Server.Name))
	
	// 初始化业务逻辑层
	bookUseCase := biz.NewBookUseCase()
	
	// 初始化服务层
	bookService := service.NewBookService(bookUseCase)
	
	// 初始化 gRPC 服务器
	grpcServer := server.NewGRPCServer(&cfg.Server, bookService)
	
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
	
	log.Info("shutting down book-service")
	grpcServer.Stop()
	log.Info("book-service stopped")
}
