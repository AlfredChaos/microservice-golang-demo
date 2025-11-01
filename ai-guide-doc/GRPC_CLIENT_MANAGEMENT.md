# gRPC客户端连接管理公共模块设计方案

## 问题分析

### 当前现状

1. **api-gateway的实现**：
   - 已有完善的gRPC连接管理（`ConnectionManager`）
   - 使用客户端工厂模式（`ClientFactory`）
   - 集中管理所有后端服务连接
   - 在启动时建立连接，关闭时统一释放

2. **内部服务的问题**：
   - 内部服务作为客户端调用其他服务时，没有统一的连接管理
   - 每次需要调用其他服务时，都需要单独创建连接
   - 缺乏连接复用和生命周期管理
   - 代码重复，维护成本高

3. **具体痛点**：
   ```go
   // 在每个服务中都需要重复这样的代码
   bookConn, err := grpc.Dial(bookServiceAddr, ...)
   bookClient := bookv1.NewBookServiceClient(bookConn)
   ```

---

## 解决方案：抽取公共gRPC客户端管理模块

### 方案概述

将api-gateway中的gRPC连接管理能力抽取到 `pkg/grpcclient` 包中，作为所有服务（api-gateway和内部服务）的统一客户端管理模块。

### 核心优势

✅ **统一管理**：所有服务使用相同的连接管理逻辑  
✅ **连接复用**：避免重复创建连接，提高性能  
✅ **配置驱动**：通过配置文件声明需要连接的服务  
✅ **易于扩展**：新增服务只需修改配置和注册表  
✅ **生命周期管理**：统一处理连接建立和关闭  
✅ **拦截器支持**：统一添加日志、追踪、重试等拦截器  

---

## 设计方案

### 1. 目录结构

```
pkg/
└── grpcclient/
    ├── manager.go           # 连接管理器
    ├── config.go            # 配置结构
    ├── registry.go          # 客户端注册表
    ├── interceptor.go       # 拦截器工具
    └── README.md            # 使用文档
```

---

### 2. 核心组件设计

#### 2.1 配置结构

```go
// pkg/grpcclient/config.go
package grpcclient

import "time"

// Config gRPC客户端配置
type Config struct {
    Services []ServiceConfig `yaml:"services" mapstructure:"services"`
}

// ServiceConfig 单个服务配置
type ServiceConfig struct {
    Name    string        `yaml:"name" mapstructure:"name"`       // 服务名称
    Address string        `yaml:"address" mapstructure:"address"` // 服务地址
    Timeout time.Duration `yaml:"timeout" mapstructure:"timeout"` // 连接超时
    
    // 可选配置
    Retry   *RetryConfig  `yaml:"retry" mapstructure:"retry"`     // 重试配置
    TLS     *TLSConfig    `yaml:"tls" mapstructure:"tls"`         // TLS配置
}

// RetryConfig 重试配置
type RetryConfig struct {
    Max         int           `yaml:"max" mapstructure:"max"`                   // 最大重试次数
    Timeout     time.Duration `yaml:"timeout" mapstructure:"timeout"`           // 重试超时
    Backoff     time.Duration `yaml:"backoff" mapstructure:"backoff"`           // 退避时间
}

// TLSConfig TLS配置
type TLSConfig struct {
    Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`   // 是否启用TLS
    CertFile string `yaml:"cert_file" mapstructure:"cert_file"` // 证书文件
    KeyFile  string `yaml:"key_file" mapstructure:"key_file"`   // 密钥文件
}
```

**配置文件示例**：

```yaml
# configs/user-service.yaml
grpc_clients:
  services:
    - name: book-service
      address: localhost:9002
      timeout: 5s
      retry:
        max: 3
        timeout: 10s
        backoff: 100ms
    - name: order-service
      address: localhost:9003
      timeout: 5s
```

---

#### 2.2 连接管理器

```go
// pkg/grpcclient/manager.go
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
    configs     map[string]*ServiceConfig
    mu          sync.RWMutex
}

// NewManager 创建连接管理器
func NewManager() *Manager {
    return &Manager{
        connections: make(map[string]*grpc.ClientConn),
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
```

---

#### 2.3 客户端注册表

```go
// pkg/grpcclient/registry.go
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
```

---

#### 2.4 拦截器

```go
// pkg/grpcclient/interceptor.go
package grpcclient

import (
    "context"
    "time"
    
    "github.com/alfredchaos/demo/pkg/log"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
)

// LoggingInterceptor 日志拦截器
func LoggingInterceptor() grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
        start := time.Now()
        
        log.WithContext(ctx).Info("grpc client call",
            zap.String("method", method),
            zap.String("target", cc.Target()))
        
        err := invoker(ctx, method, req, reply, cc, opts...)
        
        duration := time.Since(start)
        if err != nil {
            log.WithContext(ctx).Error("grpc client call failed",
                zap.String("method", method),
                zap.Duration("duration", duration),
                zap.Error(err))
        } else {
            log.WithContext(ctx).Info("grpc client call completed",
                zap.String("method", method),
                zap.Duration("duration", duration))
        }
        
        return err
    }
}

// TracingInterceptor 追踪拦截器
// 将trace ID从context传递到gRPC metadata
func TracingInterceptor() grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
        // 从context中提取trace ID
        traceID := ""
        if val := ctx.Value("X-Request-ID"); val != nil {
            if id, ok := val.(string); ok {
                traceID = id
            }
        }
        
        // 添加到metadata
        if traceID != "" {
            md := metadata.Pairs("X-Trace-ID", traceID)
            ctx = metadata.NewOutgoingContext(ctx, md)
        }
        
        return invoker(ctx, method, req, reply, cc, opts...)
    }
}

// RetryInterceptor 重试拦截器
func RetryInterceptor(cfg *RetryConfig) grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
        var err error
        
        for i := 0; i <= cfg.Max; i++ {
            err = invoker(ctx, method, req, reply, cc, opts...)
            if err == nil {
                return nil
            }
            
            // 最后一次不需要等待
            if i < cfg.Max {
                time.Sleep(cfg.Backoff)
            }
        }
        
        return err
    }
}
```

---

### 3. 使用方式

#### 3.1 在api-gateway中使用

```go
// cmd/api-gateway/main.go
package main

import (
    "github.com/alfredchaos/demo/pkg/grpcclient"
    userv1 "github.com/alfredchaos/demo/api/user/v1"
    bookv1 "github.com/alfredchaos/demo/api/book/v1"
)

func init() {
    // 注册客户端工厂
    grpcclient.GlobalRegistry.Register("user-service", func(conn *grpc.ClientConn) interface{} {
        return userv1.NewUserServiceClient(conn)
    })
    
    grpcclient.GlobalRegistry.Register("book-service", func(conn *grpc.ClientConn) interface{} {
        return bookv1.NewBookServiceClient(conn)
    })
}

func main() {
    // 加载配置
    var cfg Config
    config.MustLoadConfig("api-gateway", &cfg)
    
    // 初始化gRPC客户端管理器
    clientManager := grpcclient.NewManager()
    
    // 注册服务
    for _, svc := range cfg.GRPCClients.Services {
        clientManager.Register(&svc)
    }
    
    // 连接所有服务
    if err := clientManager.ConnectAll(); err != nil {
        log.Fatal("failed to connect services", zap.Error(err))
    }
    defer clientManager.Close()
    
    // 获取客户端
    userConn, _ := clientManager.GetConnection("user-service")
    userClient := userv1.NewUserServiceClient(userConn)
    
    // ... 后续业务逻辑
}
```

#### 3.2 在内部服务中使用

```go
// cmd/user-service/main.go
package main

import (
    "github.com/alfredchaos/demo/pkg/grpcclient"
    bookv1 "github.com/alfredchaos/demo/api/book/v1"
)

func init() {
    // 注册需要调用的其他服务
    grpcclient.GlobalRegistry.Register("book-service", func(conn *grpc.ClientConn) interface{} {
        return bookv1.NewBookServiceClient(conn)
    })
}

func main() {
    // 加载配置
    var cfg conf.Config
    config.MustLoadConfig("user-service", &cfg)
    
    // 初始化gRPC客户端管理器（如果需要调用其他服务）
    var clientManager *grpcclient.Manager
    if len(cfg.GRPCClients.Services) > 0 {
        clientManager = grpcclient.NewManager()
        
        // 注册服务
        for _, svc := range cfg.GRPCClients.Services {
            clientManager.Register(&svc)
        }
        
        // 连接所有服务
        if err := clientManager.ConnectAll(); err != nil {
            log.Fatal("failed to connect services", zap.Error(err))
        }
        defer clientManager.Close()
    }
    
    // 初始化数据层
    dataLayer, _ := data.NewData(pgDB, mongoClient, redisClient, mqClient)
    
    // 初始化业务层（注入gRPC客户端）
    var bookClient bookv1.BookServiceClient
    if clientManager != nil {
        bookConn, _ := clientManager.GetConnection("book-service")
        bookClient = bookv1.NewBookServiceClient(bookConn)
    }
    
    userUseCase := biz.NewUserUseCase(dataLayer.UserRepo, bookClient)
    
    // ... 后续初始化
}
```

---

### 4. 配置文件示例

#### 4.1 api-gateway配置

```yaml
# configs/api-gateway.yaml
server:
  name: api-gateway
  host: 0.0.0.0
  port: 8080

log:
  level: info
  format: console

grpc_clients:
  services:
    - name: user-service
      address: localhost:9001
      timeout: 5s
      retry:
        max: 3
        timeout: 10s
        backoff: 100ms
    - name: book-service
      address: localhost:9002
      timeout: 5s
      retry:
        max: 3
        timeout: 10s
        backoff: 100ms
    - name: order-service
      address: localhost:9003
      timeout: 5s
```

#### 4.2 内部服务配置

```yaml
# configs/user-service.yaml
server:
  name: user-service
  host: 0.0.0.0
  port: 9001

log:
  level: debug

database:
  enabled: true
  # ...

# gRPC客户端配置（仅当需要调用其他服务时）
grpc_clients:
  services:
    - name: book-service
      address: localhost:9002
      timeout: 5s
    - name: order-service
      address: localhost:9003
      timeout: 5s
```

---

## 迁移计划

### 阶段1：创建公共模块（第1周）

1. ✅ 在 `pkg/grpcclient` 中实现核心组件
2. ✅ 编写单元测试
3. ✅ 编写使用文档

### 阶段2：迁移api-gateway（第2周）

1. 修改api-gateway使用新的公共模块
2. 删除 `internal/api-gateway/client` 中的旧代码
3. 测试功能正常

### 阶段3：更新内部服务（第3周）

1. 更新user-service使用新模块
2. 更新book-service使用新模块
3. 更新其他服务

### 阶段4：文档和培训（第4周）

1. 更新架构文档
2. 编写最佳实践
3. 团队培训

---

## 优势对比

### 使用公共模块前

```go
// 每个服务都需要重复编写
bookConn, err := grpc.Dial(
    "localhost:9002",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithBlock(),
    grpc.WithTimeout(5*time.Second),
)
if err != nil {
    // 错误处理
}
defer bookConn.Close()

bookClient := bookv1.NewBookServiceClient(bookConn)
```

**问题**：
- ❌ 代码重复
- ❌ 没有连接复用
- ❌ 难以统一配置
- ❌ 缺少拦截器支持
- ❌ 连接泄漏风险

### 使用公共模块后

```go
// 在main.go初始化时
clientManager := grpcclient.NewManager()
clientManager.Register(&grpcclient.ServiceConfig{
    Name:    "book-service",
    Address: "localhost:9002",
    Timeout: 5 * time.Second,
})
clientManager.ConnectAll()
defer clientManager.Close()

// 在业务代码中
bookConn, _ := clientManager.GetConnection("book-service")
bookClient := bookv1.NewBookServiceClient(bookConn)
```

**优势**：
- ✅ 代码简洁
- ✅ 连接复用
- ✅ 配置驱动
- ✅ 统一拦截器
- ✅ 生命周期管理

---

## 注意事项

1. **向后兼容**：迁移过程中保持API兼容
2. **渐进式迁移**：先迁移api-gateway，再迁移内部服务
3. **配置验证**：添加配置校验，避免运行时错误
4. **监控和日志**：增加连接状态监控
5. **错误处理**：完善错误处理和重试机制

---

## 总结

通过抽取公共gRPC客户端管理模块，我们可以：

1. **统一管理**：所有服务使用相同的连接管理逻辑
2. **简化开发**：新增服务调用只需配置，无需编写重复代码
3. **提高性能**：连接复用，减少连接开销
4. **易于维护**：集中管理，修改一处即可
5. **增强功能**：统一支持日志、追踪、重试等功能

这是一个值得投资的重构，将大大提升代码质量和开发效率。
