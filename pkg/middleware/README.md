# gRPC 中间件（拦截器）

gRPC 服务的拦截器集合，可被所有 gRPC 服务复用。

## 拦截器列表

### 1. Recovery（Panic 恢复）
**文件**: `recovery.go`

**功能**: 捕获 gRPC 处理过程中的 panic，防止服务崩溃

**拦截器**:
- `UnaryServerRecovery()` - 一元 RPC 拦截器
- `StreamServerRecovery()` - 流式 RPC 拦截器

**特性**:
- 捕获所有 panic
- 记录完整的堆栈信息
- 返回 `codes.Internal` 错误
- 使用结构化日志记录

**使用**:
```go
server := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        middleware.UnaryServerRecovery(),
    ),
    grpc.ChainStreamInterceptor(
        middleware.StreamServerRecovery(),
    ),
)
```

---

### 2. Logging（日志记录）
**文件**: `logging.go`

**功能**: 记录每个 gRPC 请求的详细信息

**拦截器**:
- `UnaryServerLogging()` - 一元 RPC 拦截器
- `StreamServerLogging()` - 流式 RPC 拦截器

**记录内容**:
- 方法名（FullMethod）
- 请求耗时
- 错误信息（如果有）
- 流类型（仅流式RPC）

**日志级别**:
- `ERROR`: 请求返回错误
- `INFO`: 请求成功

**使用**:
```go
server := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        middleware.UnaryServerLogging(),
    ),
    grpc.ChainStreamInterceptor(
        middleware.StreamServerLogging(),
    ),
)
```

---

### 3. Tracing（追踪）
**文件**: `tracing.go`

**功能**: 从 metadata 中提取追踪 ID，传递到上下文

**拦截器**:
- `UnaryServerTracing()` - 一元 RPC 拦截器
- `StreamServerTracing()` - 流式 RPC 拦截器

**特性**:
- 从 metadata 读取 `x-trace-id`
- 将 trace-id 存储到 context
- 支持分布式追踪

**使用**:
```go
server := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        middleware.UnaryServerTracing(),
    ),
    grpc.ChainStreamInterceptor(
        middleware.StreamServerTracing(),
    ),
)

// 在业务代码中获取 trace-id
traceID := middleware.GetTraceID(ctx)
```

---

## 拦截器顺序

推荐的拦截器执行顺序：

```go
server := grpc.NewServer(
    // 一元拦截器
    grpc.ChainUnaryInterceptor(
        middleware.UnaryServerRecovery(), // 1. 最先执行，捕获panic
        middleware.UnaryServerTracing(),  // 2. 提取追踪ID
        middleware.UnaryServerLogging(),  // 3. 记录日志
    ),
    // 流拦截器
    grpc.ChainStreamInterceptor(
        middleware.StreamServerRecovery(), // 1. 最先执行，捕获panic
        middleware.StreamServerTracing(),  // 2. 提取追踪ID
        middleware.StreamServerLogging(),  // 3. 记录日志
    ),
)
```

## 拦截器类型

gRPC 支持两种类型的拦截器：

### 一元拦截器 (Unary Interceptor)
用于普通的请求-响应模式的 RPC 调用。

**签名**:
```go
type UnaryServerInterceptor func(
    ctx context.Context,
    req interface{},
    info *UnaryServerInfo,
    handler UnaryHandler,
) (resp interface{}, err error)
```

### 流拦截器 (Stream Interceptor)
用于流式 RPC 调用（客户端流、服务端流、双向流）。

**签名**:
```go
type StreamServerInterceptor func(
    srv interface{},
    ss ServerStream,
    info *StreamServerInfo,
    handler StreamHandler,
) error
```

## 设计原则

1. **单一职责**: 每个拦截器只负责一个特定功能
2. **可复用**: 所有 gRPC 服务都可以使用
3. **高内聚低耦合**: 拦截器之间相互独立
4. **统一日志**: 使用共享的日志系统
5. **错误处理**: 规范的错误码和错误信息

## 扩展建议

未来可以添加的拦截器：

- **Authentication**: 认证拦截器（JWT/mTLS）
- **Authorization**: 授权拦截器（RBAC）
- **RateLimit**: 限流拦截器
- **Metrics**: 指标收集拦截器（Prometheus）
- **Validation**: 参数验证拦截器
- **Retry**: 重试拦截器（客户端）
- **CircuitBreaker**: 熔断器（客户端）
- **LoadBalancing**: 负载均衡（客户端）

## 客户端拦截器

目前只实现了服务端拦截器，未来可以添加客户端拦截器：

```go
// 客户端拦截器示例
conn, err := grpc.Dial(
    address,
    grpc.WithChainUnaryInterceptor(
        clientLogging(),
        clientRetry(),
        clientTimeout(),
    ),
)
```
