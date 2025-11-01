package grpcclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// Manager gRPC客户端连接管理器
type Manager struct {
	connections map[string]*grpc.ClientConn
	clients     map[string]interface{} // 缓存客户端实例
	configs     map[string]*ServiceConfig
	mu          sync.RWMutex
}

// 初始化gRPC客户端管理器
func InitGRPCClientManager(cfg *Config) *Manager {
	clientManager := NewManager()

	// 注册服务配置
	for _, svc := range cfg.Services {
		log.Info("registering service", zap.String("service", svc.Name))
		if err := clientManager.Register(&svc); err != nil {
			log.Fatal("failed to register service",
				zap.String("service", svc.Name),
				zap.Error(err))
		}
	}

	// 连接所有服务
	if err := clientManager.ConnectAll(); err != nil {
		log.Fatal("failed to connect services", zap.Error(err))
	}

	log.Info("grpc client manager initialized")
	return clientManager
}

// NewManager 创建连接管理器
func NewManager() *Manager {
	return &Manager{
		connections: make(map[string]*grpc.ClientConn),
		clients:     make(map[string]interface{}),
		configs:     make(map[string]*ServiceConfig),
	}
}

// Register 注册服务配置
func (m *Manager) Register(cfg *ServiceConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cfg.Name == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	if cfg.Address == "" {
		return fmt.Errorf("service address cannot be empty")
	}

	m.configs[cfg.Name] = cfg
	log.Info("service registered", zap.String("service", cfg.Name), zap.String("addr", cfg.Address))
	return nil
}

// Connect 连接到指定服务
func (m *Manager) Connect(serviceName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已连接
	if _, exists := m.connections[serviceName]; exists {
		return nil
	}

	// 获取配置
	cfg, exists := m.configs[serviceName]
	if !exists {
		return fmt.Errorf("service %s not registered", serviceName)
	}

	// 构建连接选项
	opts := m.buildDialOptions(cfg)

	// 设置超时
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	// 创建连接
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, cfg.Address, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", serviceName, err)
	}

	m.connections[serviceName] = conn
	log.Info("grpc connection established",
		zap.String("service", serviceName),
		zap.String("addr", cfg.Address))

	return nil
}

// ConnectAll 连接所有已注册的服务
func (m *Manager) ConnectAll() error {
	m.mu.RLock()
	serviceNames := make([]string, 0, len(m.configs))
	for name := range m.configs {
		serviceNames = append(serviceNames, name)
	}
	m.mu.RUnlock()

	for _, name := range serviceNames {
		if err := m.Connect(name); err != nil {
			return err
		}
	}

	return nil
}

// GetConnection 获取指定服务的连接
func (m *Manager) GetConnection(serviceName string) (*grpc.ClientConn, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conn, exists := m.connections[serviceName]
	if !exists {
		return nil, fmt.Errorf("connection not found for service: %s", serviceName)
	}

	return conn, nil
}

// GetClient 获取指定服务的客户端实例
// 如果客户端已创建则返回缓存，否则使用注册表创建新客户端
func (m *Manager) GetClient(serviceName string) (interface{}, error) {
	m.mu.RLock()
	// 检查客户端缓存
	if client, exists := m.clients[serviceName]; exists {
		m.mu.RUnlock()
		return client, nil
	}
	m.mu.RUnlock()

	// 获取连接
	conn, err := m.GetConnection(serviceName)
	if err != nil {
		return nil, err
	}

	// 使用全局注册表创建客户端
	client, err := GlobalRegistry.CreateClient(serviceName, conn)
	if err != nil {
		return nil, err
	}

	// 缓存客户端
	m.mu.Lock()
	m.clients[serviceName] = client
	m.mu.Unlock()

	return client, nil
}

// Close 关闭所有连接
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for serviceName, conn := range m.connections {
		if err := conn.Close(); err != nil {
			log.Error("failed to close grpc connection",
				zap.String("service", serviceName),
				zap.Error(err))
			errs = append(errs, err)
		} else {
			log.Info("grpc connection closed", zap.String("service", serviceName))
		}
	}

	// 清空连接map
	m.connections = make(map[string]*grpc.ClientConn)

	if len(errs) > 0 {
		return fmt.Errorf("failed to close %d connection(s)", len(errs))
	}

	return nil
}

// buildDialOptions 构建连接选项
func (m *Manager) buildDialOptions(cfg *ServiceConfig) []grpc.DialOption {
	opts := []grpc.DialOption{
		// 保持连接活跃
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	// TLS配置
	if cfg.TLS != nil && cfg.TLS.Enabled {
		// TODO: 添加TLS凭证
		// creds, _ := credentials.NewClientTLSFromFile(cfg.TLS.CertFile, "")
		// opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// 添加拦截器
	unaryInterceptors := []grpc.UnaryClientInterceptor{
		LoggingInterceptor(),
		TracingInterceptor(),
	}

	// 重试配置
	if cfg.Retry != nil {
		unaryInterceptors = append(unaryInterceptors, RetryInterceptor(cfg.Retry))
	}

	opts = append(opts, grpc.WithChainUnaryInterceptor(unaryInterceptors...))

	return opts
}
