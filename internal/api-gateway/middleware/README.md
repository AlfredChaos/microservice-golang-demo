# API Gateway 中间件

API Gateway 的 HTTP 中间件集合，基于 Gin 框架实现。

## 中间件列表

### 1. Recovery（Panic 恢复）
**文件**: `recovery.go`

**功能**: 捕获 HTTP 处理过程中的 panic，防止服务崩溃

**特性**:
- 捕获所有 panic
- 记录完整的堆栈信息
- 返回统一的 500 错误响应
- 包含 request_id 用于追踪

**使用**:
```go
router.Use(middleware.Recovery())
```

---

### 2. RequestID（请求追踪）
**文件**: `request_id.go`

**功能**: 为每个请求生成或传递唯一的请求ID

**特性**:
- 支持从请求头读取已有的 `X-Request-ID`
- 如果没有则自动生成 UUID
- 将 request_id 设置到上下文和响应头
- 便于日志追踪和问题排查

**使用**:
```go
router.Use(middleware.RequestID())

// 在处理函数中获取
requestID := middleware.GetRequestID(c)
```

---

### 3. Logger（请求日志）
**文件**: `logger.go`

**功能**: 记录每个 HTTP 请求的详细信息

**记录内容**:
- 请求ID
- 请求方法
- 请求路径
- 响应状态码
- 客户端IP
- 请求耗时
- User-Agent
- 错误信息（如果有）

**日志级别**:
- `ERROR`: 状态码 >= 500
- `WARN`: 状态码 >= 400
- `INFO`: 其他

**使用**:
```go
router.Use(middleware.Logger())
```

---

### 4. CORS（跨域资源共享）
**文件**: `cors.go`

**功能**: 处理跨域请求，允许前端应用访问API

**配置**:
- `Access-Control-Allow-Origin`: `*` (允许所有来源)
- `Access-Control-Allow-Methods`: `GET, POST, PUT, DELETE, OPTIONS, PATCH`
- `Access-Control-Allow-Headers`: `Content-Type, Authorization, X-Request-ID, X-Trace-ID`
- `Access-Control-Max-Age`: `86400` (24小时)

**使用**:
```go
router.Use(middleware.CORS())
```

---

### 5. Timeout（请求超时）
**文件**: `timeout.go`

**功能**: 为每个请求设置超时时间，防止请求长时间占用资源

**特性**:
- 基于 context.WithTimeout 实现
- 超时后返回 408 错误
- 自动取消超时的请求
- 包含 request_id

**使用**:
```go
// 设置 30 秒超时
router.Use(middleware.Timeout(30 * time.Second))
```

---

## 中间件顺序

中间件的执行顺序很重要，推荐顺序：

```go
router.Use(
    middleware.Recovery(),      // 1. 最先执行，确保能捕获所有panic
    middleware.RequestID(),     // 2. 生成请求ID，供后续中间件使用
    middleware.Logger(),        // 3. 记录日志
    middleware.CORS(),          // 4. 处理跨域
    middleware.Timeout(30*time.Second), // 5. 设置超时
)
```

## 设计原则

1. **单一职责**: 每个中间件只负责一个特定功能
2. **可组合**: 中间件可以灵活组合使用
3. **高内聚低耦合**: 中间件之间相互独立
4. **依赖注入**: 通过参数传递配置，不依赖全局变量
5. **错误处理**: 完善的错误处理和日志记录

## 扩展建议

未来可以添加的中间件：

- **RateLimit**: 限流中间件，防止API被滥用
- **Authentication**: 认证中间件（JWT/OAuth2）
- **Authorization**: 授权中间件（RBAC/ABAC）
- **Metrics**: 指标收集中间件（Prometheus）
- **Tracing**: 分布式追踪中间件（OpenTelemetry）
- **Cache**: 响应缓存中间件
- **Compression**: 响应压缩中间件（gzip）
