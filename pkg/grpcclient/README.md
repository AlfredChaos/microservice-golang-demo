# gRPC客户端管理公共模块

## 概述

`pkg/grpcclient` 是一个统一的gRPC客户端连接管理模块，提供了配置驱动的连接管理、客户端注册、拦截器支持等功能。

## 功能特性

- ✅ **统一管理**：所有服务使用相同的连接管理逻辑
- ✅ **连接复用**：避免重复创建连接，提高性能
- ✅ **配置驱动**：通过配置文件声明需要连接的服务
- ✅ **易于扩展**：新增服务只需修改配置和注册表
- ✅ **生命周期管理**：统一处理连接建立和关闭
- ✅ **拦截器支持**：统一添加日志、追踪、重试等拦截器

## 核心组件

### 1. Manager (连接管理器)
负责管理所有gRPC连接的生命周期。

### 2. Registry (客户端注册表)
维护服务名称到客户端创建函数的映射。

### 3. Config (配置结构)
定义服务连接的配置参数，包括地址、超时、重试等。

### 4. Interceptor (拦截器)
提供日志记录、链路追踪、重试等通用功能。

## 使用方式

### 1. 配置文件

```yaml
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
```

### 2. 注册客户端工厂

```go
func init() {
    // 注册gRPC客户端工厂
    grpcclient.GlobalRegistry.Register("user-service", func(conn *grpc.ClientConn) interface{} {
        return userv1.NewUserServiceClient(conn)
    })
    
    grpcclient.GlobalRegistry.Register("book-service", func(conn *grpc.ClientConn) interface{} {
        return bookv1.NewBookServiceClient(conn)
    })
}
```

### 3. 初始化和使用

```go
func main() {
    // 创建gRPC客户端管理器
    clientManager := grpcclient.NewManager()
    defer clientManager.Close()
    
    // 注册服务配置
    for _, svc := range cfg.GRPCClients.Services {
        clientManager.Register(&svc)
    }
    
    // 连接所有服务
    if err := clientManager.ConnectAll(); err != nil {
        log.Fatal("failed to connect services", zap.Error(err))
    }
    
    // 获取连接并创建客户端
    userConn, _ := clientManager.GetConnection("user-service")
    userClient := userv1.NewUserServiceClient(userConn)
    
    // 使用客户端进行调用
    resp, err := userClient.SayHello(ctx, &userv1.HelloRequest{})
}
```

## 迁移指南

### 从旧版本迁移

**迁移前**：
```go
// 每个服务都需要重复编写
bookConn, err := grpc.Dial(
    "localhost:9002",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithBlock(),
    grpc.WithTimeout(5*time.Second),
)
defer bookConn.Close()
bookClient := bookv1.NewBookServiceClient(bookConn)
```

**迁移后**：
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

## 优势对比

### 使用公共模块的优势

- ✅ **代码简洁**：减少重复代码
- ✅ **连接复用**：提高性能
- ✅ **配置驱动**：易于管理
- ✅ **统一拦截器**：日志、追踪、重试等功能
- ✅ **生命周期管理**：避免连接泄漏

### 解决的问题

- ❌ 代码重复
- ❌ 没有连接复用
- ❌ 难以统一配置
- ❌ 缺少拦截器支持
- ❌ 连接泄漏风险

## 注意事项

1. **服务注册**：确保在使用前注册所有需要的服务
2. **配置验证**：检查配置文件中的服务地址和参数
3. **错误处理**：妥善处理连接失败的情况
4. **资源释放**：程序退出时调用 `Close()` 方法
5. **并发安全**：Manager 是并发安全的，可以在多个goroutine中使用

## 示例项目

参考 `cmd/api-gateway/main.go` 中的完整使用示例。
