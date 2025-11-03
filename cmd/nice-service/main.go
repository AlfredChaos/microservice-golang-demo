package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/alfredchaos/demo/internal/nice-service/conf"
	"github.com/alfredchaos/demo/internal/nice-service/dependencies"
	// "github.com/alfredchaos/demo/internal/nice-service/server"
	"github.com/alfredchaos/demo/pkg/config"
	"github.com/alfredchaos/demo/pkg/grpcclient"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

// func init() {
// 	// 注册 gRPC 客户端工厂
// 	grpcclient.GlobalRegistry.Register("user-service", func(conn *grpc.ClientConn) interface{} {
// 		return userv1.NewUserServiceClient(conn)
// 	})
// }

func main() {
	var cfg conf.Config
	config.MustLoadConfig("nice-service", &cfg)

	log.MustInitLogger(&cfg.Log, cfg.Server.Name)
	defer log.Sync()

	log.Info("starting nice-service", zap.String("name", cfg.Server.Name))

	// 初始化 gRPC 客户端管理器（未来可能需要调用其他服务）
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

	// ============================================================
	// gRPC 服务器（暂时注释，未来可能需要同时支持同步和异步通信）
	// ============================================================
	// grpcServer := server.NewGRPCServerBuilder(&cfg.Server).
	// 	WithNiceService(appCtx.NiceService).Build()
	// log.Info("grpc server initialized")
	// go func() {
	// 	if err := grpcServer.Start(); err != nil {
	// 		log.Fatal("failed to start grpc server", zap.Error(err))
	// 	}
	// }()

	// ============================================================
	// RabbitMQ 消费者启动
	// ============================================================
	if appCtx.Consumer != nil && appCtx.HandleService != nil {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// 启动消费者
		go func() {
			log.Info("starting rabbitmq consumer",
				zap.String("queue", cfg.RabbitMQ.Queue),
				zap.String("routing_key", cfg.RabbitMQ.RoutingKey))

			// 使用 HandleService.HandleMessage 作为消息处理器
			if err := appCtx.Consumer.Consume(ctx, appCtx.HandleService.HandleMessage); err != nil {
				log.Error("consumer stopped with error", zap.Error(err))
			}
		}()
		log.Info("rabbitmq consumer started successfully")
	} else {
		log.Warn("consumer or handle service is not initialized, skipping consumer startup")
	}

	// ============================================================
	// 优雅关闭
	// ============================================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down nice-service...")

	// 关闭消费者
	if appCtx.Consumer != nil {
		if err := appCtx.Consumer.Close(); err != nil {
			log.Error("failed to close consumer", zap.Error(err))
		} else {
			log.Info("consumer closed successfully")
		}
	}

	// 关闭消息队列
	if appCtx.MessageQueue != nil {
		if err := appCtx.MessageQueue.Close(); err != nil {
			log.Error("failed to close message queue", zap.Error(err))
		} else {
			log.Info("message queue closed successfully")
		}
	}

	// 未来如果启用 gRPC 服务器
	// grpcServer.Stop()

	log.Info("nice-service stopped gracefully")
}
