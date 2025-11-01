# API Gateway - Client 层实现

> gRPC 客户端管理层的完整实现代码

## 1. ConnectionManager - 连接管理器

### 文件路径
`internal/api-gateway/client/connection_manager.go`

### 完整代码

```go
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
```

### 功能说明

1. **线程安全**：使用 `sync.RWMutex` 保证并发访问安全
2. **连接复用**：检查连接是否已存在，避免重复创建
3. **统一管理**：集中管理所有 gRPC 连接的生命周期
4. **日志记录**：记录连接建立和关闭的日志

---

## 2. ClientFactory - 客户端工厂

### 文件路径
`internal/api-gateway/client/client_factory.go`

### 完整代码

```go
package client

import (
	orderv1 "github.com/alfredchaos/demo/api/order/v1"
	userv1 "github.com/alfredchaos/demo/api/user/v1"
)

// ClientFactory gRPC 客户端工厂
// 提供创建各服务 gRPC 客户端的方法
type ClientFactory struct {
	connManager *ConnectionManager
}

// NewClientFactory 创建客户端工厂
func NewClientFactory(connManager *ConnectionManager) *ClientFactory {
	return &ClientFactory{
		connManager: connManager,
	}
}

// CreateUserClient 创建用户服务客户端
func (f *ClientFactory) CreateUserClient() (userv1.UserServiceClient, error) {
	conn, err := f.connManager.GetConnection("user-service")
	if err != nil {
		return nil, err
	}
	return userv1.NewUserServiceClient(conn), nil
}

// CreateBookClient 创建图书服务客户端
func (f *ClientFactory) CreateBookClient() (orderv1.BookServiceClient, error) {
	conn, err := f.connManager.GetConnection("book-service")
	if err != nil {
		return nil, err
	}
	return orderv1.NewBookServiceClient(conn), nil
}

// 扩展示例：添加更多服务客户端创建方法
//
// CreateOrderClient 创建订单服务客户端
// func (f *ClientFactory) CreateOrderClient() (orderv1.OrderServiceClient, error) {
//     conn, err := f.connManager.GetConnection("order-service")
//     if err != nil {
//         return nil, err
//     }
//     return orderv1.NewOrderServiceClient(conn), nil
// }
//
// CreatePaymentClient 创建支付服务客户端
// func (f *ClientFactory) CreatePaymentClient() (paymentv1.PaymentServiceClient, error) {
//     conn, err := f.connManager.GetConnection("payment-service")
//     if err != nil {
//         return nil, err
//     }
//     return paymentv1.NewPaymentServiceClient(conn), nil
// }
```

### 功能说明

1. **工厂模式**：封装客户端创建逻辑
2. **统一接口**：所有客户端创建方法遵循相同模式
3. **易于扩展**：添加新服务只需添加新方法
4. **错误处理**：统一处理连接获取错误

---

## 3. 使用示例

### 3.1 在 main.go 中初始化

```go
// 创建连接管理器
connManager := client.NewConnectionManager()
defer connManager.Close()

// 连接各个后端服务
if _, err := connManager.Connect("user-service", "localhost:9001"); err != nil {
    log.Fatal("failed to connect to user-service", zap.Error(err))
}

if _, err := connManager.Connect("book-service", "localhost:9002"); err != nil {
    log.Fatal("failed to connect to book-service", zap.Error(err))
}
```

### 3.2 在 wire.go 中使用

```go
// 创建客户端工厂
clientFactory := client.NewClientFactory(deps.ConnManager)

// 创建各服务的客户端
userClient, err := clientFactory.CreateUserClient()
if err != nil {
    log.Fatal("failed to create user client", zap.Error(err))
}

bookClient, err := clientFactory.CreateBookClient()
if err != nil {
    log.Fatal("failed to create book client", zap.Error(err))
}
```

---

## 4. 设计优势

### 4.1 关注点分离

- **ConnectionManager**：专注于连接管理
- **ClientFactory**：专注于客户端创建

### 4.2 易于测试

- 可以 mock `ConnectionManager` 进行单元测试
- 可以注入测试连接进行集成测试

### 4.3 扩展性强

添加新服务只需：
1. 在 `main.go` 中调用 `connManager.Connect()`
2. 在 `ClientFactory` 中添加对应的创建方法

### 4.4 资源管理

- 统一的连接生命周期管理
- 防止连接泄漏
- 优雅关闭所有连接

---

## 5. 注意事项

### 5.1 连接超时

当前设置为 5 秒，可根据实际情况调整：

```go
grpc.WithTimeout(5*time.Second)
```

### 5.2 连接重试

当前实现不包含自动重试，如需重试可以添加：

```go
import "google.golang.org/grpc/backoff"

// 添加重试配置
grpc.WithConnectParams(grpc.ConnectParams{
    Backoff: backoff.Config{
        BaseDelay:  1.0 * time.Second,
        Multiplier: 1.6,
        Jitter:     0.2,
        MaxDelay:   120 * time.Second,
    },
})
```

### 5.3 健康检查

可以添加连接健康检查：

```go
import "google.golang.org/grpc/health/grpc_health_v1"

// 检查连接健康状态
func (cm *ConnectionManager) HealthCheck(serviceName string) error {
    conn, err := cm.GetConnection(serviceName)
    if err != nil {
        return err
    }
    
    // 使用 gRPC Health Check Protocol
    healthClient := grpc_health_v1.NewHealthClient(conn)
    resp, err := healthClient.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
    if err != nil {
        return err
    }
    
    if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
        return fmt.Errorf("service %s is not healthy", serviceName)
    }
    
    return nil
}
```

---

**Client 层实现完成**
