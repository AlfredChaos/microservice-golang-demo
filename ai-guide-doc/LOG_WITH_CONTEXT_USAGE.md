# log.WithContext() 使用指南

## 概述

`log.WithContext()` 提供了一种优雅的方式，从 `context.Context` 中自动提取日志相关的上下文信息，无需手动传递各种字段。

## 功能特性

### 自动提取的字段

`WithContext()` 会自动从 context 中提取以下信息（如果存在）：

- **trace_id**: 追踪ID，用于分布式链路追踪
- **request_id**: 请求ID，用于单个请求的日志关联
- **user_id**: 用户ID（如果业务代码设置了）
- **request**: 请求信息对象
  - method: 请求方法 (GET, POST等)
  - path: 请求路径
  - client_ip: 客户端IP

如果某个字段在 context 中不存在，则会自动忽略该字段。

## 使用方式

### 1. 基本用法

在 Controller 中使用：

```go
func (h *HelloController) SayHello(c *gin.Context) {
    ctx := c.Request.Context()
    
    // 自动附加 request_id, trace_id, request info 等字段
    log.WithContext(ctx).Info("received hello request")
}
```

输出示例：
```json
{
  "timestamp": "2024-10-29T16:05:30.123456789+08:00",
  "level": "info",
  "message": "received hello request",
  "service": "api-gateway",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "request": {
    "method": "POST",
    "path": "/api/v1/hello",
    "client_ip": "192.168.1.100"
  }
}
```

### 2. 附加额外的业务字段

```go
// 可以继续链式调用，附加额外的业务字段
log.WithContext(ctx).Info("business logic completed",
    log.DurationMs(100),
    log.ExtraData("order_id", "12345"),
    zap.String("status", "success"),
)
```

输出示例：
```json
{
  "timestamp": "2024-10-29T16:05:30.123456789+08:00",
  "level": "info",
  "message": "business logic completed",
  "service": "api-gateway",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "request": {
    "method": "POST",
    "path": "/api/v1/hello",
    "client_ip": "192.168.1.100"
  },
  "duration_ms": 100,
  "extra_data": {
    "order_id": "12345"
  },
  "status": "success"
}
```

### 3. 在业务代码中手动添加用户ID

如果某些业务逻辑需要附加用户ID到 context：

```go
func (h *HelloController) SayHello(c *gin.Context) {
    ctx := c.Request.Context()
    
    // 从认证信息中获取用户ID
    userID := getUserIDFromAuth(c)
    
    // 将用户ID附加到 context
    ctx = log.ContextWithUserID(ctx, userID)
    
    // 之后的日志都会自动包含 user_id
    log.WithContext(ctx).Info("user action")
}
```

## Middleware 自动注入

系统已配置以下中间件自动将信息注入到 context：

### 1. RequestID 中间件
- 文件: `internal/api-gateway/middleware/request_id.go`
- 功能: 为每个请求生成或提取 `request_id`
- 自动注入到: `c.Request.Context()`

### 2. Logger 中间件
- 文件: `internal/api-gateway/middleware/logger.go`
- 功能: 将请求信息（method, path, client_ip）注入到 context
- 自动注入到: `c.Request.Context()`

### 3. Tracing 中间件（gRPC）
- 文件: `pkg/middleware/tracing.go`
- 功能: 为 gRPC 请求生成或提取 `trace_id`
- 自动注入到: gRPC context

## Context 辅助函数

### 存储到 Context

```go
// 存储 trace_id
ctx = log.ContextWithTraceID(ctx, "trace-123")

// 存储 request_id
ctx = log.ContextWithRequestID(ctx, "req-456")

// 存储 user_id
ctx = log.ContextWithUserID(ctx, "user-789")

// 存储请求信息
ctx = log.ContextWithRequestInfo(ctx, "GET", "/api/users", "192.168.1.1")
```

### 从 Context 获取

```go
// 获取 trace_id
traceID := log.GetTraceIDFromContext(ctx)

// 获取 request_id
requestID := log.GetRequestIDFromContext(ctx)

// 获取 user_id
userID := log.GetUserIDFromContext(ctx)

// 获取请求信息
reqInfo := log.GetRequestInfoFromContext(ctx)
if reqInfo != nil {
    fmt.Println(reqInfo.Method, reqInfo.Path, reqInfo.ClientIP)
}
```

## 最佳实践

### ✅ 推荐做法

1. **在 Controller 层使用 WithContext**
   ```go
   log.WithContext(c.Request.Context()).Info("processing request")
   ```

2. **传递 context 到业务层**
   ```go
   func ProcessOrder(ctx context.Context, orderID string) error {
       log.WithContext(ctx).Info("processing order", zap.String("order_id", orderID))
       // ...
   }
   ```

3. **在 goroutine 中使用独立的 context**
   ```go
   go func() {
       // 创建新的 context，但保留父 context 的值
       ctx := context.WithValue(c.Request.Context(), "key", "value")
       log.WithContext(ctx).Info("async task started")
   }()
   ```

### ❌ 不推荐做法

1. **不要在 goroutine 中直接使用 gin.Context**
   ```go
   // ❌ 错误：gin.Context 不是并发安全的
   go func() {
       log.WithContext(c.Request.Context()).Info("async task")
   }()
   
   // ✅ 正确：先提取 context
   ctx := c.Request.Context()
   go func() {
       log.WithContext(ctx).Info("async task")
   }()
   ```

2. **不要重复手动传递已在 context 中的字段**
   ```go
   // ❌ 不推荐：request_id 已在 context 中
   log.WithContext(ctx).Info("test", zap.String("request_id", requestID))
   
   // ✅ 推荐：自动提取
   log.WithContext(ctx).Info("test")
   ```

## 并发安全性

`WithContext()` 是并发安全的：

- `zap.Logger.With()` 返回新的 logger 实例，不会修改原 logger
- 每个请求获得独立的 logger 副本
- 多个 goroutine 可以同时调用 `WithContext(ctx)` 而不会相互影响

```go
// 并发场景示例
func handleRequest(c *gin.Context) {
    ctx := c.Request.Context()
    
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            // 每个 goroutine 获得独立的 logger，互不干扰
            log.WithContext(ctx).Info("worker", zap.Int("worker_id", id))
        }(i)
    }
    wg.Wait()
}
```

## 与旧 API 的兼容性

新的 `WithContext` API 与旧的单字段 API 完全兼容：

```go
// 旧 API - 仍然可用
log.WithTraceID("trace-123").Info("test")
log.WithUserID("user-456").Info("test")

// 新 API - 推荐使用
log.WithContext(ctx).Info("test")

// 组合使用
logger := log.WithContext(ctx)
logger = logger.With(zap.String("custom_field", "value"))
logger.Info("test")
```

## 总结

`log.WithContext()` 提供了一种简洁、统一的方式来处理日志上下文信息：

1. **自动化**: 无需手动传递常见字段
2. **灵活性**: 可以随时添加额外字段
3. **可维护性**: 上下文信息集中管理
4. **并发安全**: 可在多 goroutine 中安全使用
5. **向后兼容**: 不影响现有代码
