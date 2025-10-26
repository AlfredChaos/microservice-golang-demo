package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/alfredchaos/demo/internal/nice-service/conf"
	"github.com/alfredchaos/demo/internal/nice-service/consumer"
	"github.com/alfredchaos/demo/pkg/config"
	"github.com/alfredchaos/demo/pkg/log"
	"github.com/alfredchaos/demo/pkg/mq"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	var cfg conf.Config
	config.MustLoadConfig("nice-service", &cfg)
	
	// 初始化日志
	log.MustInitLogger(&cfg.Log, cfg.Server.Name)
	defer log.Sync()
	
	log.Info("starting nice-service", zap.String("name", cfg.Server.Name))
	
	// 初始化 RabbitMQ 客户端
	rabbitMQClient := mq.MustNewRabbitMQClient(&cfg.RabbitMQ)
	defer func() {
		if err := rabbitMQClient.Close(); err != nil {
			log.Error("failed to close rabbitmq client", zap.Error(err))
		}
	}()
	
	// 创建消费者
	mqConsumer := mq.NewRabbitMQConsumer(rabbitMQClient)
	messageConsumer := consumer.NewMessageConsumer(mqConsumer)
	
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 启动消费者
	if err := messageConsumer.Start(ctx); err != nil {
		log.Fatal("failed to start consumer", zap.Error(err))
	}
	
	log.Info("nice-service started, waiting for messages...")
	
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Info("shutting down nice-service")
	cancel()
	
	if err := messageConsumer.Stop(); err != nil {
		log.Error("failed to stop consumer", zap.Error(err))
	}
	
	log.Info("nice-service stopped")
}
