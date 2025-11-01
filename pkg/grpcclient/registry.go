package grpcclient

import (
	"fmt"
	"sync"
	
	"google.golang.org/grpc"
)

// ClientFactory 客户端创建函数
type ClientFactory func(conn *grpc.ClientConn) interface{}

// Registry 客户端注册表
// 维护服务名称到客户端创建函数的映射
type Registry struct {
	factories map[string]ClientFactory
	mu        sync.RWMutex
}

// GlobalRegistry 全局注册表实例
var GlobalRegistry = NewRegistry()

// NewRegistry 创建注册表
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]ClientFactory),
	}
}

// Register 注册客户端工厂
func (r *Registry) Register(serviceName string, factory ClientFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[serviceName] = factory
}

// CreateClient 创建客户端
func (r *Registry) CreateClient(serviceName string, conn *grpc.ClientConn) (interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	factory, exists := r.factories[serviceName]
	if !exists {
		return nil, fmt.Errorf("client factory not found for service: %s", serviceName)
	}
	
	return factory(conn), nil
}

// 使用示例：注册客户端工厂
// func init() {
//     // 注册user-service客户端
//     grpcclient.GlobalRegistry.Register("user-service", func(conn *grpc.ClientConn) interface{} {
//         return userv1.NewUserServiceClient(conn)
//     })
//     
//     // 注册book-service客户端
//     grpcclient.GlobalRegistry.Register("book-service", func(conn *grpc.ClientConn) interface{} {
//         return bookv1.NewBookServiceClient(conn)
//     })
// }
