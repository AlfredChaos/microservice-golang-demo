package server

import (
	"time"

	userv1 "github.com/alfredchaos/demo/api/user/v1"
	"github.com/alfredchaos/demo/internal/user-service/conf"
	"github.com/alfredchaos/demo/internal/user-service/service"
	"github.com/alfredchaos/demo/pkg/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
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

// WithUserService 添加用户服务
func (b *GRPCServerBuilder) WithUserService(svc *service.UserService) *GRPCServerBuilder {
	b.registrars = append(b.registrars, func(s *grpc.Server) {
		userv1.RegisterUserServiceServer(s, svc)
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
		// KeepAlive 策略：允许客户端发送 ping
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             30 * time.Second, // 允许客户端最快30秒发一次ping（小于客户端的60秒）
			PermitWithoutStream: true,             // 允许在没有活动流时发送ping
		}),
		// KeepAlive 参数：服务器端的连接管理
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Minute, // 连接空闲15分钟后关闭
			MaxConnectionAge:      30 * time.Minute, // 连接最多存活30分钟
			MaxConnectionAgeGrace: 5 * time.Second,  // 优雅关闭等待5秒
			Time:                  5 * time.Minute,  // 服务器每5分钟发一次ping
			Timeout:               1 * time.Second,  // ping超时1秒
		}),
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
