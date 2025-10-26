package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alfredchaos/demo/internal/api-gateway/client"
	"github.com/alfredchaos/demo/internal/api-gateway/controller"
	"github.com/alfredchaos/demo/internal/api-gateway/router"
	"github.com/alfredchaos/demo/pkg/config"
	"github.com/alfredchaos/demo/pkg/log"
	"github.com/alfredchaos/demo/pkg/mq"
	"go.uber.org/zap"
)

// Config api-gateway 配置结构
type Config struct {
	Server   ServerConfig      `yaml:"server" mapstructure:"server"`       // 服务器配置
	Log      log.LogConfig     `yaml:"log" mapstructure:"log"`             // 日志配置
	Services ServicesConfig    `yaml:"services" mapstructure:"services"`   // 后端服务配置
	RabbitMQ mq.RabbitMQConfig `yaml:"rabbitmq" mapstructure:"rabbitmq"`   // RabbitMQ 配置
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Name string `yaml:"name" mapstructure:"name"` // 服务名称
	Host string `yaml:"host" mapstructure:"host"` // 监听地址
	Port int    `yaml:"port" mapstructure:"port"` // 监听端口
}

// ServicesConfig 后端服务配置
type ServicesConfig struct {
	UserService string `yaml:"user_service" mapstructure:"user_service"` // user-service 地址
	BookService string `yaml:"book_service" mapstructure:"book_service"` // book-service 地址
}

// @title Demo API Gateway
// @version 1.0
// @description 微服务架构演示项目的 API 网关
// @host localhost:8080
// @BasePath /
func main() {
	// 加载配置
	var cfg Config
	config.MustLoadConfig("api-gateway", &cfg)
	
	// 初始化日志
	log.MustInitLogger(&cfg.Log, cfg.Server.Name)
	defer log.Sync()
	
	log.Info("starting api-gateway", zap.String("name", cfg.Server.Name))
	
	// 初始化 gRPC 客户端
	grpcClients, err := client.NewGRPCClients(cfg.Services.UserService, cfg.Services.BookService)
	if err != nil {
		log.Fatal("failed to create grpc clients", zap.Error(err))
	}
	log.Info("grpc clients initialized")
	
	// 初始化 RabbitMQ 客户端
	rabbitMQClient := mq.MustNewRabbitMQClient(&cfg.RabbitMQ)
	defer func() {
		if err := rabbitMQClient.Close(); err != nil {
			log.Error("failed to close rabbitmq client", zap.Error(err))
		}
	}()
	
	// 创建消息发布者
	publisher := mq.NewRabbitMQPublisher(rabbitMQClient)
	log.Info("rabbitmq publisher initialized")
	
	// 初始化控制器
	helloController := controller.NewHelloController(grpcClients, publisher)
	
	// 设置路由
	r := router.SetupRouter(helloController)
	
	// 启动 HTTP 服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Info("http server starting", zap.String("addr", addr))
	
	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatal("failed to start http server", zap.Error(err))
		}
	}()
	
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Info("shutting down api-gateway")
	log.Info("api-gateway stopped")
}
