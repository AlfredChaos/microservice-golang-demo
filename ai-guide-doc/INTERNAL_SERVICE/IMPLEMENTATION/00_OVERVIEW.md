# 内部服务完整编码方案 - 总览

> 本文档基于内部服务设计文档，提供完整的编码实施方案

## 文档导航

| 文档 | 说明 |
|------|------|
| [00_OVERVIEW.md](./00_OVERVIEW.md) | 编码方案总览（本文档） |
| [01_PROJECT_STRUCTURE.md](./01_PROJECT_STRUCTURE.md) | 项目结构和文件清单 |
| [02_DOMAIN_LAYER.md](./02_DOMAIN_LAYER.md) | 领域层实现方案 |
| [03_DATA_LAYER.md](./03_DATA_LAYER.md) | 数据访问层实现方案 |
| [04_BIZ_LAYER.md](./04_BIZ_LAYER.md) | 业务逻辑层实现方案 |
| [05_SERVICE_LAYER.md](./05_SERVICE_LAYER.md) | 服务层实现方案 |
| [06_SERVER_LAYER.md](./06_SERVER_LAYER.md) | 服务器层实现方案 |
| [07_CONFIGURATION.md](./07_CONFIGURATION.md) | 配置管理实现方案 |
| [08_MAIN_ENTRY.md](./08_MAIN_ENTRY.md) | 主函数和依赖注入实现 |
| [09_GRPC_CLIENT.md](./09_GRPC_CLIENT.md) | gRPC客户端集成方案 |
| [10_DATABASE_MIGRATION.md](./10_DATABASE_MIGRATION.md) | 数据库迁移方案 |
| [11_RABBITMQ_INTEGRATION.md](./11_RABBITMQ_INTEGRATION.md) | RabbitMQ集成方案 |
| [12_TESTING.md](./12_TESTING.md) | 测试方案 |
| [13_CHECKLIST.md](./13_CHECKLIST.md) | 实施检查清单 |

## 编码方案概述

本编码方案基于以下设计文档：
- ✅ 分层架构设计 (ARCHITECTURE.md)
- ✅ 数据存储层设计 (DATA_STORAGE.md)
- ✅ 依赖注入设计 (DI_AND_WIRE.md)
- ✅ 服务间通信设计 (SERVICE_COMMUNICATION.md)
- ✅ 开发指南 (DEVELOPMENT_GUIDE.md)

## 核心设计原则

### 1. 统一架构模式
所有内部服务遵循相同的分层架构：
```
Server Layer (gRPC服务器)
    ↓
Service Layer (gRPC接口实现)
    ↓
Biz Layer (业务逻辑)
    ↓
Domain Layer (领域模型)
    ↓
Data Layer (数据访问)
    ↓
Infrastructure (PostgreSQL/MongoDB/Redis/RabbitMQ)
```

### 2. 依赖注入原则
- **依赖倒置**：高层模块依赖接口，不依赖具体实现
- **接口隔离**：定义最小化的接口
- **手动注入**：不使用Wire，保持代码透明
- **构造函数注入**：通过构造函数传递依赖

### 3. 数据存储原则
- **仓储模式**：通过Repository接口抽象数据访问
- **PO/DO分离**：持久化对象和领域对象分离
- **多数据源支持**：PostgreSQL、MongoDB、Redis可选配置
- **事务支持**：在需要时提供事务封装

### 4. 服务通信原则
- **gRPC接口统一管理**：所有Proto文件放在 `api/` 目录
- **使用pkg/grpcclient管理客户端**：统一的连接管理和配置
- **双向通信**：既作为服务端也可作为客户端
- **异步通信**：通过RabbitMQ实现事件驱动

## 实施步骤

### 阶段1：项目基础设施（必需）
1. ✅ 定义gRPC接口（Proto文件）
2. ✅ 生成Go代码
3. ✅ 创建目录结构
4. ✅ 配置文件准备

### 阶段2：核心业务实现（必需）
5. ✅ 实现Domain层（领域模型）
6. ✅ 实现Data层（数据访问）
7. ✅ 实现Biz层（业务逻辑）
8. ✅ 实现Service层（gRPC服务）
9. ✅ 实现Server层（gRPC服务器）

### 阶段3：配置和启动（必需）
10. ✅ 实现配置加载
11. ✅ 实现依赖注入（main.go）
12. ✅ 数据库迁移

### 阶段4：高级功能（可选）
13. ⚠️ gRPC客户端集成（如需调用其他服务）
14. ⚠️ RabbitMQ集成（如需异步通信）
15. ⚠️ Redis缓存集成（如需缓存）

### 阶段5：测试和部署
16. ✅ 单元测试
17. ✅ 集成测试
18. ✅ 部署配置

## 技术栈

### 核心框架
- **gRPC**: 服务间通信协议
- **Protobuf**: 接口定义和序列化

### 数据存储
- **GORM**: PostgreSQL ORM
- **mongo-driver**: MongoDB官方驱动
- **go-redis**: Redis客户端
- **Goose**: 数据库迁移工具

### 消息队列
- **RabbitMQ**: 异步消息队列
- **amqp091-go**: RabbitMQ Go客户端

### 工具库
- **Viper**: 配置管理
- **Zap**: 结构化日志
- **UUID**: 唯一ID生成

### 测试
- **testify**: 测试断言和Mock
- **mockery**: Mock生成工具

## 示例服务

本编码方案以**order-service（订单服务）**作为完整示例，包含：

### 功能需求
- 创建订单
- 获取订单详情
- 列出用户订单
- 订单状态管理

### 数据存储
- **PostgreSQL**: 订单数据（主存储）
- **Redis**: 订单缓存（可选）
- **MongoDB**: 订单事件日志（可选）

### 服务依赖
- **user-service**: 验证用户存在
- **product-service**: 获取商品信息（可选）

### 消息队列
- 发布订单创建事件
- 消费订单状态更新事件

## 代码规范

### 1. 命名规范
```go
// 接口名：名词 + Interface 或直接名词
type UserRepository interface { }
type UserUseCase interface { }

// 实现名：接口名 + 实现标识（小写开头，私有）
type userPgRepository struct { }
type userUseCase struct { }

// 构造函数：New + 接口名 + 实现标识（可选）
func NewUserPgRepository(db *gorm.DB) UserRepository { }
func NewUserUseCase(repo UserRepository) UserUseCase { }

// 文件名：小写 + 下划线
// user_repo.go
// user_pg_repo.go
// user_usecase.go
```

### 2. 包组织
```
internal/order-service/
├── domain/          # 领域模型（不依赖外部）
├── data/            # 数据访问（依赖infrastructure）
├── biz/             # 业务逻辑（依赖domain和data接口）
├── service/         # gRPC服务（依赖biz接口）
├── server/          # 服务器（依赖service）
├── conf/            # 配置（独立）
└── consumer/        # 消息消费者（可选）
```

### 3. 错误处理
```go
// 定义领域错误
var (
    ErrOrderNotFound = errors.New("order not found")
    ErrInvalidOrder  = errors.New("invalid order")
)

// 在Service层转换为gRPC错误
if errors.Is(err, domain.ErrOrderNotFound) {
    return nil, status.Error(codes.NotFound, "order not found")
}
```

### 4. 日志规范
```go
// 使用WithContext获取带追踪信息的logger
log.WithContext(ctx).Info("creating order", 
    zap.String("user_id", userID),
    zap.Int("items_count", len(items)))

// 错误日志必须包含error
log.WithContext(ctx).Error("failed to create order", zap.Error(err))
```

### 5. 注释规范
```go
// CreateOrder 创建订单
// 参数：
//   - ctx: 上下文，用于超时控制和追踪
//   - userID: 用户ID
//   - items: 订单项列表
// 返回：
//   - *domain.Order: 创建的订单
//   - error: 错误信息
func (uc *orderUseCase) CreateOrder(ctx context.Context, userID string, items []domain.OrderItem) (*domain.Order, error) {
    // 实现...
}
```

## 配置管理

### 配置文件结构
```yaml
# configs/order-service.yaml
server:
  name: order-service
  host: 0.0.0.0
  port: 9003

log:
  level: debug
  format: console

database:
  enabled: true
  driver: postgres
  host: localhost
  port: 5432
  username: postgres
  password: password
  database: order_service

redis:
  enabled: false
  addr: localhost:6379

mongodb:
  enabled: false
  uri: mongodb://localhost:27017

rabbitmq:
  enabled: false
  url: amqp://guest:guest@localhost:5672/

grpc_clients:  # 可选：需要调用其他服务时
  services:
    - name: user-service
      address: localhost:9001
      timeout: 5s
```

### 配置加载
```go
var cfg conf.Config
config.MustLoadConfig("order-service", &cfg)
```

## 依赖注入流程

```go
func main() {
    // 1. 加载配置
    var cfg conf.Config
    config.MustLoadConfig("order-service", &cfg)
    
    // 2. 初始化日志
    log.MustInitLogger(&cfg.Log, cfg.Server.Name)
    
    // 3. 初始化基础设施（根据配置）
    var pgDB *gorm.DB
    if cfg.Database.Enabled {
        pgDB = db.MustNewPostgresDB(&cfg.Database)
        defer closePgDB(pgDB)
    }
    
    var redisClient *cache.RedisClient
    if cfg.Redis.Enabled {
        redisClient = cache.MustNewRedisClient(&cfg.Redis)
        defer redisClient.Close()
    }
    
    // 4. 初始化数据访问层
    dataLayer, err := data.NewData(pgDB, nil, redisClient, nil)
    if err != nil {
        log.Fatal("failed to initialize data layer", zap.Error(err))
    }
    defer dataLayer.Close(context.Background())
    
    // 5. 初始化业务逻辑层
    orderUseCase := biz.NewOrderUseCase(dataLayer.OrderRepo)
    
    // 6. 初始化服务层
    orderService := service.NewOrderService(orderUseCase)
    
    // 7. 初始化gRPC服务器
    grpcServer := server.NewGRPCServer(&cfg.Server, orderService)
    
    // 8. 启动服务器
    go grpcServer.Start()
    
    // 9. 优雅关闭
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    grpcServer.Stop()
}
```

## 测试策略

### 单元测试
- **Domain层**：测试业务规则和验证逻辑
- **Biz层**：使用Mock Repository测试业务逻辑
- **Data层**：使用测试数据库测试数据访问

### 集成测试
- 启动测试容器（PostgreSQL、Redis等）
- 测试完整的业务流程
- 测试服务间调用

### Mock策略
```go
// 使用接口实现Mock
type MockOrderRepository struct {
    CreateFunc func(ctx context.Context, order *domain.Order) error
}

func (m *MockOrderRepository) Create(ctx context.Context, order *domain.Order) error {
    if m.CreateFunc != nil {
        return m.CreateFunc(ctx, order)
    }
    return nil
}
```

## 关键注意事项

### 1. 数据库连接管理
- ✅ 使用连接池
- ✅ 设置合理的超时
- ✅ 在程序退出时关闭连接
- ⚠️ 注意连接泄漏

### 2. gRPC客户端管理
- ✅ 使用 `pkg/grpcclient` 统一管理
- ✅ 配置重试和超时
- ✅ 处理连接错误
- ⚠️ 避免阻塞主goroutine

### 3. 错误处理
- ✅ 定义领域错误
- ✅ 在Service层转换为gRPC错误
- ✅ 记录详细的错误日志
- ⚠️ 不要泄露敏感信息

### 4. 并发安全
- ✅ Repository实现必须是并发安全的
- ✅ 使用Context传递请求上下文
- ✅ 注意goroutine泄漏

### 5. 性能优化
- ⚠️ 使用缓存减少数据库查询
- ⚠️ 批量操作而非循环
- ⚠️ 使用数据库索引
- ⚠️ 监控慢查询

## 下一步

请按以下顺序阅读详细的实施方案：

1. 📁 [项目结构](./01_PROJECT_STRUCTURE.md) - 创建目录和文件
2. 🔷 [领域层](./02_DOMAIN_LAYER.md) - 实现领域模型
3. 💾 [数据层](./03_DATA_LAYER.md) - 实现数据访问
4. 💼 [业务层](./04_BIZ_LAYER.md) - 实现业务逻辑
5. 🔌 [服务层](./05_SERVICE_LAYER.md) - 实现gRPC接口
6. 🖥️ [服务器层](./06_SERVER_LAYER.md) - 配置gRPC服务器
7. ⚙️ [配置管理](./07_CONFIGURATION.md) - 实现配置加载
8. 🚀 [主函数](./08_MAIN_ENTRY.md) - 实现依赖注入
9. 🔗 [gRPC客户端](./09_GRPC_CLIENT.md) - 集成服务间调用（可选）
10. 🗄️ [数据库迁移](./10_DATABASE_MIGRATION.md) - 管理数据库版本
11. 📨 [RabbitMQ](./11_RABBITMQ_INTEGRATION.md) - 集成消息队列（可选）
12. ✅ [测试](./12_TESTING.md) - 编写测试代码
13. 📋 [检查清单](./13_CHECKLIST.md) - 验证实施完整性

## 参考资源

### 设计文档
- [ARCHITECTURE.md](../ARCHITECTURE.md) - 分层架构设计
- [DATA_STORAGE.md](../DATA_STORAGE.md) - 数据存储设计
- [DI_AND_WIRE.md](../DI_AND_WIRE.md) - 依赖注入设计
- [SERVICE_COMMUNICATION.md](../SERVICE_COMMUNICATION.md) - 服务间通信
- [DEVELOPMENT_GUIDE.md](../DEVELOPMENT_GUIDE.md) - 开发指南

### 公共模块
- `pkg/config/` - 配置加载
- `pkg/log/` - 日志管理
- `pkg/db/` - 数据库连接
- `pkg/cache/` - Redis客户端
- `pkg/mq/` - RabbitMQ客户端
- `pkg/grpcclient/` - gRPC客户端管理

### 示例代码
- `internal/user-service/` - 用户服务示例
- `internal/book-service/` - 书籍服务示例
- `internal/nice-service/` - Nice服务示例

## 问题反馈

如果在实施过程中遇到问题：
1. 查阅对应的设计文档
2. 参考示例服务代码
3. 检查公共模块文档
4. 查看实施检查清单

---

**版本**: v1.0  
**更新日期**: 2025-10-31  
**适用范围**: 所有内部服务（user-service, book-service, order-service等）
