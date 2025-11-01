package client

import (
	"fmt"
	"sync"
	"time"

	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ConnectionManager gRPC 连接管理器
// 统一管理所有后端服务的 gRPC 连接
type ConnectionManager struct {
	connections map[string]*grpc.ClientConn
	mu          sync.RWMutex
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*grpc.ClientConn),
	}
}

// Connect 连接到指定服务
// serviceName: 服务名称（用于日志和标识）
// addr: 服务地址
func (cm *ConnectionManager) Connect(serviceName, addr string) (*grpc.ClientConn, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 检查是否已存在连接
	if conn, exists := cm.connections[serviceName]; exists {
		return conn, nil
	}

	// 创建新连接
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", serviceName, err)
	}

	cm.connections[serviceName] = conn
	log.Info("grpc connection established",
		zap.String("service", serviceName),
		zap.String("addr", addr))

	return conn, nil
}

// GetConnection 获取指定服务的连接
func (cm *ConnectionManager) GetConnection(serviceName string) (*grpc.ClientConn, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, exists := cm.connections[serviceName]
	if !exists {
		return nil, fmt.Errorf("connection not found for service: %s", serviceName)
	}

	return conn, nil
}

// Close 关闭所有连接
func (cm *ConnectionManager) Close() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var errs []error
	for serviceName, conn := range cm.connections {
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
	cm.connections = make(map[string]*grpc.ClientConn)

	if len(errs) > 0 {
		return fmt.Errorf("failed to close %d connection(s)", len(errs))
	}

	return nil
}
