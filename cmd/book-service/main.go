package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/alfredchaos/demo/internal/book-service/conf"
	"github.com/alfredchaos/demo/internal/book-service/dependencies"
	"github.com/alfredchaos/demo/internal/book-service/server"
	"github.com/alfredchaos/demo/pkg/config"
	"github.com/alfredchaos/demo/pkg/grpcclient"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

func main() {
	var cfg conf.Config
	config.MustLoadConfig("book-service", &cfg)

	log.MustInitLogger(&cfg.Log, cfg.Server.Name)
	defer log.Sync()

	log.Info("starting book-service",
		zap.String("name", cfg.Server.Name),
		zap.String("addr", cfg.Server.GetAddr()))

	// 初始化 gRPC 客户端管理器
	clientManager := grpcclient.InitGRPCClientManager(&cfg.GRPCClients)
	defer func() {
		if err := clientManager.Close(); err != nil {
			log.Error("failed to close grpc client manager", zap.Error(err))
		}
	}()

	// 依赖注入
	deps := &dependencies.Dependencies{
		ClientManager: clientManager,
		Cfg:           &cfg,
	}
	appCtx, err := dependencies.InjectDependencies(deps)
	if err != nil {
		log.Error("failed to inject dependencies", zap.Error(err))
		return
	}
	log.Info("dependencies injected successfully")

	grpcServer := server.NewGRPCServerBuilder(&cfg.Server).
		WithBookService(appCtx.BookService).Build()
	log.Info("grpc server initialized")
	go func() {
		if err := grpcServer.Start(); err != nil {
			log.Fatal("failed to start grpc server", zap.Error(err))
		}
	}()

	// ============================================================
	// 优雅关闭
	// ============================================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down user-service...")
	grpcServer.Stop()
	log.Info("user-service stopped gracefully")
}
