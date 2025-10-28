package server

import (
	"fmt"
	"net"

	orderv1 "github.com/alfredchaos/demo/api/order/v1"
	"github.com/alfredchaos/demo/internal/book-service/conf"
	"github.com/alfredchaos/demo/internal/book-service/service"
	"github.com/alfredchaos/demo/pkg/log"
	"github.com/alfredchaos/demo/pkg/middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// GRPCServer gRPC 服务器封装
type GRPCServer struct {
	server  *grpc.Server
	config  *conf.ServerConfig
	service *service.BookService
}

// NewGRPCServer 创建新的 gRPC 服务器
func NewGRPCServer(cfg *conf.ServerConfig, bookService *service.BookService) *GRPCServer {
	// 创建 gRPC 服务器，应用拦截器链
	server := grpc.NewServer(
		// 一元拦截器（按顺序执行）
		grpc.ChainUnaryInterceptor(
			middleware.UnaryServerRecovery(), // 1. Panic恢复
			middleware.UnaryServerTracing(),  // 2. 追踪
			middleware.UnaryServerLogging(),  // 3. 日志记录
		),
		// 流拦截器（按顺序执行）
		grpc.ChainStreamInterceptor(
			middleware.StreamServerRecovery(), // 1. Panic恢复
			middleware.StreamServerTracing(),  // 2. 追踪
			middleware.StreamServerLogging(),  // 3. 日志记录
		),
	)
	
	// 注册服务
	orderv1.RegisterBookServiceServer(server, bookService)
	
	return &GRPCServer{
		server:  server,
		config:  cfg,
		service: bookService,
	}
}

// Start 启动 gRPC 服务器
func (s *GRPCServer) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	
	log.Info("gRPC server starting", zap.String("addr", addr))
	
	if err := s.server.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	
	return nil
}

// Stop 停止 gRPC 服务器
func (s *GRPCServer) Stop() {
	log.Info("stopping gRPC server")
	s.server.GracefulStop()
}

// GetServer 获取原始 gRPC 服务器实例
func (s *GRPCServer) GetServer() *grpc.Server {
	return s.server
}
