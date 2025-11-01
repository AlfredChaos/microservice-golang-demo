# API Gateway 依赖注入架构重构方案

> 基于依赖注入框架重新设计 api-gateway 的完整实现方案
> 
> 生成时间: 2025-10-30
> 版本: v1.0

## 目录

- [1. 架构概述](#1-架构概述)
- [2. 设计原则](#2-设计原则)
- [3. 目录结构](#3-目录结构)
- [4. 分层架构](#4-分层架构)
- [5. 实现步骤](#5-实现步骤)
- [6. 代码实现](#6-代码实现)
- [7. 使用说明](#7-使用说明)

---

## 1. 架构概述

### 1.1 重构目标

- **避免胖网关**：API Gateway 仅做路由转发，不编排业务逻辑
- **依赖注入**：基于接口编程，解耦各层依赖
- **统一管理**：gRPC 连接统一管理，易于扩展
- **清晰分层**：Domain → Service → Controller 三层架构

### 1.2 核心组件

```
┌─────────────────────────────────────────────────┐
│                   main.go                       │
│  - 加载配置                                      │
│  - 初始化连接管理器                               │
│  - 执行依赖注入                                   │
│  - 启动HTTP服务                                   │
└────────────┬────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────┐
│            wire/InjectDependencies              │
│  - 创建客户端工厂                                 │
│  - 创建Service层                                 │
│  - 创建Controller层                              │
│  - 返回AppContext                                │
└────────────┬────────────────────────────────────┘
             │
    ┌────────┴────────┐
    ▼                 ▼
┌─────────┐     ┌──────────┐
│ Service │     │Controller│
│  Layer  │────▶│  Layer   │
└─────────┘     └──────────┘
    ▼                 ▼
┌─────────┐     ┌──────────┐
│ Domain  │     │  Router  │
│Interface│     │          │
└─────────┘     └──────────┘
```

---

## 2. 设计原则

### 2.1 单一职责原则

- **Domain 层**：定义业务能力接口
- **Service 层**：实现业务逻辑，封装外部调用
- **Controller 层**：处理 HTTP 请求响应
- **Wire 层**：组装依赖关系

### 2.2 依赖倒置原则

- Controller 依赖 Domain 接口，而非具体实现
- Service 实现 Domain 接口
- 通过依赖注入组装具体实现

### 2.3 开闭原则

- 添加新服务只需扩展，无需修改现有代码
- 通过接口隔离，易于替换实现

---

## 3. 目录结构

```
internal/api-gateway/
├── client/                      # gRPC 客户端管理
│   ├── connection_manager.go    # 连接管理器
│   └── client_factory.go        # 客户端工厂
├── domain/                      # 领域接口层
│   ├── user_service.go          # 用户服务接口
│   └── book_service.go          # 图书服务接口
├── service/                     # 服务实现层
│   ├── user_service.go          # 用户服务实现
│   └── book_service.go          # 图书服务实现
├── controller/                  # 控制器层
│   ├── user_controller.go       # 用户控制器
│   └── book_controller.go       # 图书控制器
├── wire/                        # 依赖注入
│   └── wire.go                  # 依赖注入配置
├── router/                      # 路由配置
│   └── router.go                # 路由设置
├── dto/                         # 数据传输对象
│   └── response.go              # 响应结构
└── middleware/                  # 中间件
    └── ...                      # 各种中间件
```

---

## 4. 分层架构

### 4.1 Client 层（gRPC 客户端管理）

**职责**：
- 管理所有 gRPC 连接的生命周期
- 提供客户端创建工厂方法
- 线程安全的连接管理

**核心类型**：
- `ConnectionManager`：连接管理器
- `ClientFactory`：客户端工厂

### 4.2 Domain 层（领域接口）

**职责**：
- 定义业务能力接口
- 纯接口，无具体实现
- 供 Controller 层依赖

**示例接口**：
```go
type IUserService interface {
    SayHello(ctx context.Context) (string, error)
}
```

### 4.3 Service 层（服务实现）

**职责**：
- 实现 Domain 接口
- 封装 gRPC 调用逻辑
- 处理跨服务调用细节（trace ID 等）

**依赖**：
- gRPC 客户端（注入）

### 4.4 Controller 层（控制器）

**职责**：
- 处理 HTTP 请求
- 调用 Service 层
- 返回 HTTP 响应

**依赖**：
- Domain 接口（注入）

### 4.5 Wire 层（依赖注入）

**职责**：
- 定义 `AppContext`（持有所有控制器）
- 定义 `Dependencies`（封装外部依赖）
- 实现 `InjectDependencies`（组装依赖）

---

## 5. 实现步骤

### 步骤 1：创建 Client 层

1. 创建 `connection_manager.go`
2. 创建 `client_factory.go`

### 步骤 2：创建 Domain 层

1. 创建 `user_service.go`
2. 创建 `book_service.go`

### 步骤 3：创建 Service 层

1. 创建 `user_service.go`
2. 创建 `book_service.go`

### 步骤 4：创建 Controller 层

1. 创建 `user_controller.go`
2. 创建 `book_controller.go`

### 步骤 5：创建 Wire 层

1. 创建 `wire.go`

### 步骤 6：更新 Router 和 Main

1. 更新 `router/router.go`
2. 更新 `cmd/api-gateway/main.go`

### 步骤 7：清理旧代码

1. 删除 `client/grpc_client.go`
2. 删除 `controller/hello_controller.go`

---

## 6. 代码实现

详细代码实现请参考：
- [Client 层实现](./API_GATEWAY_CLIENT_LAYER.md)
- [Domain & Service 层实现](./API_GATEWAY_DOMAIN_SERVICE_LAYER.md)
- [Controller & Wire 层实现](./API_GATEWAY_CONTROLLER_WIRE_LAYER.md)
- [Router & Main 实现](./API_GATEWAY_ROUTER_MAIN.md)

---

## 7. 使用说明

### 7.1 添加新服务

假设要添加 Order Service：

**步骤 1**: 在配置文件添加服务地址
```yaml
services:
  user_service: "localhost:9001"
  book_service: "localhost:9002"
  order_service: "localhost:9003"  # 新增
```

**步骤 2**: 在 `main.go` 中连接服务
```go
if _, err := connManager.Connect("order-service", cfg.Services.OrderService); err != nil {
    log.Fatal("failed to connect to order-service", zap.Error(err))
}
```

**步骤 3**: 在 `ClientFactory` 添加创建方法
```go
func (f *ClientFactory) CreateOrderClient() (orderv1.OrderServiceClient, error) {
    conn, err := f.connManager.GetConnection("order-service")
    if err != nil {
        return nil, err
    }
    return orderv1.NewOrderServiceClient(conn), nil
}
```

**步骤 4**: 创建 Domain 接口
```go
// domain/order_service.go
type IOrderService interface {
    CreateOrder(ctx context.Context, req *OrderRequest) (*Order, error)
}
```

**步骤 5**: 实现 Service 层
```go
// service/order_service.go
type orderService struct {
    orderClient orderv1.OrderServiceClient
}

func NewOrderService(client orderv1.OrderServiceClient) domain.IOrderService {
    return &orderService{orderClient: client}
}
```

**步骤 6**: 创建 Controller
```go
// controller/order_controller.go
type IOrderController interface {
    CreateOrder(c *gin.Context)
}

type orderController struct {
    orderService domain.IOrderService
}

func NewOrderController(service domain.IOrderService) IOrderController {
    return &orderController{orderService: service}
}
```

**步骤 7**: 在 Wire 中注入
```go
// wire/wire.go
type AppContext struct {
    UserController  controller.IUserController
    BookController  controller.IBookController
    OrderController controller.IOrderController  // 新增
}

func InjectDependencies(deps *Dependencies) *AppContext {
    // ...
    orderClient, _ := clientFactory.CreateOrderClient()
    orderService := service.NewOrderService(orderClient)
    orderController := controller.NewOrderController(orderService)
    
    return &AppContext{
        UserController:  userController,
        BookController:  bookController,
        OrderController: orderController,  // 新增
    }
}
```

**步骤 8**: 添加路由
```go
// router/router.go
func SetupRouter(appCtx *wire.AppContext) *gin.Engine {
    // ...
    apiV1 := router.Group("/api/v1")
    {
        UserRouter(apiV1, appCtx.UserController)
        BookRouter(apiV1, appCtx.BookController)
        OrderRouter(apiV1, appCtx.OrderController)  // 新增
    }
}

func OrderRouter(router *gin.RouterGroup, controller controller.IOrderController) {
    orderGroup := router.Group("/order")
    {
        orderGroup.POST("", controller.CreateOrder)
    }
}
```

### 7.2 运行项目

```bash
# 启动 api-gateway
cd cmd/api-gateway
go run main.go

# 或使用 Makefile
make run-gateway
```

### 7.3 测试接口

```bash
# 测试用户服务
curl http://localhost:8080/api/v1/user/hello

# 测试图书服务
curl http://localhost:8080/api/v1/book

# 健康检查
curl http://localhost:8080/health

# 查看 Swagger 文档
open http://localhost:8080/swagger/index.html
```

---

## 8. 优势总结

### 8.1 架构优势

- ✅ **解耦性强**：各层通过接口依赖，易于测试和替换
- ✅ **扩展性好**：添加新服务只需按步骤扩展，不影响现有代码
- ✅ **维护性高**：职责清晰，代码组织良好
- ✅ **避免胖网关**：业务逻辑在后端服务，网关只做转发

### 8.2 技术优势

- ✅ **统一连接管理**：ConnectionManager 统一管理所有 gRPC 连接
- ✅ **线程安全**：并发访问安全
- ✅ **资源管理**：统一的生命周期管理，防止连接泄漏
- ✅ **依赖注入**：清晰的依赖关系，易于理解和维护

---

## 9. 注意事项

### 9.1 错误处理

- Service 层应返回具体的错误信息
- Controller 层应将错误转换为合适的 HTTP 状态码
- 使用统一的错误响应格式

### 9.2 日志记录

- 使用 `log.WithContext(ctx)` 自动附加 trace ID
- 在关键节点记录日志（请求开始、服务调用、错误等）
- 日志级别要合理（Info、Warn、Error）

### 9.3 超时控制

- gRPC 调用应设置超时
- HTTP 请求通过中间件统一控制超时
- 避免无限等待导致资源耗尽

### 9.4 并发安全

- ConnectionManager 使用读写锁保证线程安全
- 避免在 Service/Controller 中使用全局可变状态

---

## 10. 后续优化方向

### 10.1 配置管理

- 支持动态配置刷新
- 支持服务发现（如 Consul、Etcd）

### 10.2 监控告警

- 集成 Prometheus 指标
- 添加健康检查详情
- 请求链路追踪（OpenTelemetry）

### 10.3 限流熔断

- 集成 Sentinel 或自研限流器
- 熔断降级机制
- 请求重试策略

### 10.4 测试覆盖

- 单元测试（使用 mock 接口）
- 集成测试
- 压力测试

---

## 附录

### A. 相关文档

- [依赖注入框架示例](./prompt/di.md)
- [依赖注入设计文档](./DI_DESIGN.md)

### B. 参考资料

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Dependency Injection in Go](https://blog.drewolson.org/dependency-injection-in-go)
- [gRPC Best Practices](https://grpc.io/docs/guides/performance/)

---

**文档结束**
