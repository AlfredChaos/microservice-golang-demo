package main

import (
	"os"
	"os/signal"
	"syscall"

	bookv1 "github.com/alfredchaos/demo/api/book/v1"
	"github.com/alfredchaos/demo/internal/user-service/conf"
	"github.com/alfredchaos/demo/internal/user-service/dependencies"
	"github.com/alfredchaos/demo/internal/user-service/server"
	"github.com/alfredchaos/demo/pkg/config"
	"github.com/alfredchaos/demo/pkg/grpcclient"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func init() {
	// 注册 gRPC 客户端工厂
	grpcclient.GlobalRegistry.Register("book-service", func(conn *grpc.ClientConn) interface{} {
		return bookv1.NewBookServiceClient(conn)
	})
}

func main() {
	var cfg conf.Config
	config.MustLoadConfig("user-service", &cfg)

	log.MustInitLogger(&cfg.Log, cfg.Server.Name)
	defer log.Sync()

	log.Info("starting user-service",
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
		WithUserService(appCtx.UserService).Build()
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
