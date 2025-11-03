package server

import (
	bookv1 "github.com/alfredchaos/demo/api/book/v1"
	"github.com/alfredchaos/demo/internal/book-service/conf"
	"github.com/alfredchaos/demo/internal/book-service/service"
	"github.com/alfredchaos/demo/pkg/middleware"
	"google.golang.org/grpc"
)

// ServiceRegistrar 服务注册函数类型
type ServiceRegistrar func(*grpc.Server)

type GRPCServerBuilder struct {
	config     *conf.ServerConfig
	registrars []ServiceRegistrar
}

func NewGRPCServerBuilder(cfg *conf.ServerConfig) *GRPCServerBuilder {
	return &GRPCServerBuilder{
		config:     cfg,
		registrars: make([]ServiceRegistrar, 0),
	}
}

// WithBookService 添加Book服务
func (b *GRPCServerBuilder) WithBookService(svc *service.BookService) *GRPCServerBuilder {
	b.registrars = append(b.registrars, func(s *grpc.Server) {
		bookv1.RegisterBookServiceServer(s, svc)
	})
	return b
}

// Build 构建 gRPC 服务器
func (b *GRPCServerBuilder) Build() *GRPCServer {
	server := grpc.NewServer(
		// 一元拦截器（按顺序执行）
		grpc.ChainUnaryInterceptor(
			middleware.UnaryServerRecovery(), // 1. Panic恢复
			middleware.UnaryServerTracing(),  // 2. 追踪
			middleware.UnaryServerLogging(),  // 3. 日志记录
		),
		// 流拦截器（按顺序执行）
		grpc.ChainStreamInterceptor(
			middleware.StreamServerRecovery(),
			middleware.StreamServerTracing(),
			middleware.StreamServerLogging(),
		),
	)

	// 注册所有服务
	for _, registrar := range b.registrars {
		registrar(server)
	}

	return &GRPCServer{
		server: server,
		config: b.config,
	}
}
