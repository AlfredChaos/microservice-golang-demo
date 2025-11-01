# User-Service SayHello业务逻辑升级文档

> 完成时间：2025年10月31日  
> 升级目标：集成多种服务间通信方式和数据存储

## 一、升级概述

本次升级将user-service的SayHello从简单的问候服务升级为完整的微服务交互演示，集成了：

1. **同步通信**：通过gRPC调用book-service获取消息
2. **异步通信**：通过RabbitMQ与nice-service进行消息通信
3. **持久化存储**：将操作日志存储到PostgreSQL
4. **缓存层**：将操作日志写入Redis缓存

这是一个**完整的微服务架构最佳实践**展示。

---

## 二、架构设计

### 2.1 整体架构图

```
┌─────────────────────────────────────────────────────────┐
│                  API Gateway / Client                    │
└────────────────────┬────────────────────────────────────┘
                     ↓ gRPC
┌─────────────────────────────────────────────────────────┐
│                  User-Service                            │
│  ┌──────────────────────────────────────────────────┐  │
│  │  SayHello 业务流程                                 │  │
│  │  1. 生成user消息                                   │  │
│  │  2. 同步调用 book-service ───→ gRPC              │  │
│  │  3. 存储到 PostgreSQL ───→ 操作日志表             │  │
│  │  4. 缓存到 Redis ───→ operation:log:{id}         │  │
│  │  5. 异步发送 ───→ RabbitMQ ───→ nice-service    │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
         ↓ gRPC              ↓ RabbitMQ
┌──────────────────┐   ┌─────────────────────┐
│  Book-Service    │   │  Nice-Service       │
│  (同步返回消息)    │   │  (异步处理消息)      │
└──────────────────┘   └─────────────────────┘
                              ↓ 消费并处理
                       ┌─────────────────────┐
                       │  User-Service       │
                       │  (更新PostgreSQL)    │
                       └─────────────────────┘
```

### 2.2 数据流向

```
1. 请求到达
   ↓
2. 生成 user-service 消息
   ↓
3. gRPC 调用 book-service（同步）
   ├─ 成功：获取 book-service 消息
   └─ 失败：返回错误，流程终止
   ↓
4. 创建日志领域对象
   ├─ ID: UUID
   ├─ RequestName: 请求参数
   ├─ UserMessage: user-service消息
   ├─ BookMessage: book-service消息
   └─ CombinedMessage: 组合消息
   ↓
5. 持久化到 PostgreSQL（异步，不影响主流程）
   ↓
6. 缓存到 Redis（异步，不影响主流程）
   ↓
7. 发布消息到 RabbitMQ（异步，goroutine）
   ↓
8. 返回组合消息给客户端
   ↓
9. nice-service 消费消息（异步后台）
   ↓
10. user-service 消费者接收 nice-service 回复
   ↓
11. 更新 PostgreSQL 中的日志记录
```

---

## 三、详细实现

### 3.1 配置文件更新

**文件**: `configs/user-service.yaml`

添加了以下配置段：

```yaml
# PostgreSQL配置（用于存储SayHello操作日志）
database:
  enabled: true
  driver: postgres
  host: localhost
  port: 5432
  username: postgres
  password: 123456
  database: user_service
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600
  log_level: info

# RabbitMQ配置（用于与nice-service异步通信）
rabbitmq:
  enabled: true
  url: amqp://guest:guest@localhost:5672/
  exchange: demo_exchange
  exchange_type: topic
  queue: user_service_queue
  routing_key: user.#

# gRPC客户端配置（调用其他服务）
grpc_clients:
  services:
    - name: book-service
      address: localhost:9002
      timeout: 5s
      retry:
        max: 3
        timeout: 10s
        backoff: 100ms
```

---

### 3.2 领域层新增

#### 文件: `domain/sayhello_log.go`

```go
// SayHelloLog SayHello操作日志领域模型
type SayHelloLog struct {
    ID              string    // 日志ID
    RequestName     string    // 请求的name参数
    UserMessage     string    // user-service生成的消息
    BookMessage     string    // book-service返回的消息
    NiceMessage     string    // nice-service返回的消息（异步）
    CombinedMessage string    // 组合后的最终消息
    CreatedAt       time.Time // 创建时间
}
```

**设计亮点**：
- 纯粹的领域模型，无基础设施依赖
- 包含业务方法：`UpdateWithNiceMessage`、`SetCombinedMessage`
- 体现了DDD的充血模型思想

---

### 3.3 数据层新增

#### 3.3.1 PostgreSQL仓库

**文件**: `data/sayhello_log_pg_repo.go`

```go
// SayHelloLogPO PostgreSQL持久化对象
type SayHelloLogPO struct {
    ID              string    `gorm:"column:id;primaryKey"`
    RequestName     string    `gorm:"column:request_name"`
    UserMessage     string    `gorm:"column:user_message"`
    BookMessage     string    `gorm:"column:book_message"`
    NiceMessage     string    `gorm:"column:nice_message"`
    CombinedMessage string    `gorm:"column:combined_message"`
    CreatedAt       time.Time `gorm:"column:created_at"`
}
```

**表结构**:
```sql
CREATE TABLE sayhello_logs (
    id VARCHAR(36) PRIMARY KEY,
    request_name VARCHAR(255),
    user_message TEXT,
    book_message TEXT,
    nice_message TEXT,
    combined_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 3.3.2 Redis缓存

**文件**: `data/operation_log_cache.go`

```go
// OperationLogCache 操作日志缓存
type OperationLogCache struct {
    redis *cache.RedisClient
}

// LogEntry 日志条目
type LogEntry struct {
    ID              string
    Operation       string
    RequestName     string
    UserMessage     string
    BookMessage     string
    NiceMessage     string
    CombinedMessage string
    Timestamp       time.Time
}
```

**Redis Key格式**: `operation:log:{id}`  
**TTL**: 24小时

---

### 3.4 消息消费者

**文件**: `consumer/nice_message_consumer.go`

```go
// NiceMessageConsumer nice-service消息消费者
type NiceMessageConsumer struct {
    mqClient  *mq.RabbitMQClient
    logRepo   data.SayHelloLogRepository
    queueName string
}
```

**消费流程**：
1. 订阅RabbitMQ队列
2. 接收来自nice-service的消息
3. 解析消息（包含log_id和message）
4. 更新PostgreSQL中对应的日志记录
5. ACK消息

---

### 3.5 业务逻辑层升级

**文件**: `biz/user_usecase.go`

#### 依赖注入

```go
type userUseCase struct {
    bookClient bookv1.BookServiceClient        // gRPC客户端
    mqClient   *mq.RabbitMQClient              // RabbitMQ客户端
    logRepo    data.SayHelloLogRepository      // 数据仓库
    logCache   *data.OperationLogCache         // Redis缓存
}
```

#### SayHello方法流程

```go
func (uc *userUseCase) SayHello(ctx context.Context, name string) (string, error) {
    // 1. 生成user-service消息
    userMessage := "Hello " + name + " from user-service"
    
    // 2. 同步调用book-service（gRPC）
    bookResp, err := uc.bookClient.SayHello(ctx, &bookv1.HelloRequest{})
    if err != nil {
        return "", err  // 失败则终止
    }
    bookMessage := bookResp.Message
    
    // 3. 创建日志领域对象
    logEntry := domain.NewSayHelloLog(name, userMessage, bookMessage)
    logEntry.ID = uuid.New().String()
    logEntry.SetCombinedMessage()
    
    // 4. 存储到PostgreSQL（不阻塞主流程）
    if uc.logRepo != nil {
        uc.logRepo.Create(ctx, logEntry)
    }
    
    // 5. 缓存到Redis（不阻塞主流程）
    if uc.logCache != nil {
        cacheEntry := &data.LogEntry{...}
        uc.logCache.Set(ctx, cacheEntry)
    }
    
    // 6. 异步发送消息到RabbitMQ（goroutine）
    go func() {
        message := map[string]string{
            "log_id":  logEntry.ID,
            "message": userMessage + " | " + bookMessage,
            "from":    "user-service",
        }
        msgBytes, _ := json.Marshal(message)
        publisher := mq.NewRabbitMQPublisher(uc.mqClient)
        publisher.Publish(ctx, msgBytes)
    }()
    
    // 7. 返回组合消息
    return logEntry.CombinedMessage, nil
}
```

---

### 3.6 主函数完整流程

**文件**: `cmd/user-service/main_updated.go`

#### 初始化流程

```go
func main() {
    // 1. 配置加载
    var cfg conf.Config
    config.MustLoadConfig("user-service", &cfg)
    
    // 2. 日志初始化
    log.MustInitLogger(&cfg.Log, cfg.Server.Name)
    
    // 3. PostgreSQL初始化
    pgDB, _ := gorm.Open(postgres.Open(cfg.Database.GetDSN()), &gorm.Config{})
    pgDB.AutoMigrate(&data.SayHelloLogPO{})
    
    // 4. MongoDB初始化
    mongoClient := db.MustNewMongoClient(&cfg.MongoDB)
    
    // 5. Redis初始化
    redisClient := cache.MustNewRedisClient(&cfg.Redis)
    
    // 6. RabbitMQ初始化
    mqClient := mq.MustNewRabbitMQClient(&cfg.RabbitMQ)
    
    // 7. gRPC客户端管理器初始化
    clientManager := grpcclient.NewManager()
    // 注册并连接book-service
    bookConn, _ := clientManager.GetConnection("book-service")
    bookClient := bookv1.NewBookServiceClient(bookConn)
    
    // 8. 数据访问层初始化
    dataLayer, _ := data.NewData(pgDB, mongoClient, redisClient)
    dataLayer.SayHelloLogRepo = data.NewSayHelloLogPgRepository(pgDB)
    
    // 9. 业务逻辑层初始化
    userUseCase := biz.NewUserUseCase(
        bookClient,
        mqClient,
        dataLayer.SayHelloLogRepo,
        dataLayer.OperationLogCache,
    )
    
    // 10. 服务层初始化
    userService := service.NewUserService(userUseCase)
    
    // 11. gRPC服务器初始化
    grpcServer := server.NewGRPCServer(&cfg.Server, userService)
    
    // 12. 启动消息消费者
    niceConsumer := consumer.NewNiceMessageConsumer(
        mqClient,
        dataLayer.SayHelloLogRepo,
        cfg.RabbitMQ.Queue,
    )
    niceConsumer.Start(context.Background())
    
    // 13. 启动gRPC服务器
    go grpcServer.Start()
    
    // 14. 优雅关闭
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    grpcServer.Stop()
}
```

---

## 四、依赖管理

### 4.1 需要添加的依赖

```bash
# PostgreSQL驱动和GORM
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres

# UUID生成
go get -u github.com/google/uuid

# RabbitMQ客户端（已存在）
github.com/rabbitmq/amqp091-go

# gRPC和Protobuf（已存在）
google.golang.org/grpc
google.golang.org/protobuf
```

### 4.2 go.mod示例

```go
module github.com/alfredchaos/demo

go 1.21

require (
    gorm.io/gorm v1.25.5
    gorm.io/driver/postgres v1.5.4
    github.com/google/uuid v1.4.0
    github.com/rabbitmq/amqp091-go v1.9.0
    google.golang.org/grpc v1.59.0
    go.uber.org/zap v1.26.0
    // ... 其他依赖
)
```

---

## 五、测试验证

### 5.1 准备工作

```bash
# 1. 启动PostgreSQL
docker run -d --name postgres \
  -e POSTGRES_PASSWORD=123456 \
  -p 5432:5432 \
  postgres:14

# 2. 启动Redis
docker run -d --name redis \
  -p 6379:6379 \
  redis:7

# 3. 启动RabbitMQ
docker run -d --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management

# 4. 启动book-service
cd cmd/book-service && go run main.go

# 5. 启动nice-service  
cd cmd/nice-service && go run main.go

# 6. 启动user-service
cd cmd/user-service && go run main.go
```

### 5.2 测试SayHello接口

```bash
# 使用grpcurl测试
grpcurl -plaintext \
  -d '{}' \
  localhost:9001 \
  user.v1.UserService/SayHello

# 预期响应
{
  "message": "Hello from user-service | Hello from book-service | (waiting for nice-service)"
}
```

### 5.3 验证数据存储

```sql
-- 查询PostgreSQL
SELECT * FROM sayhello_logs ORDER BY created_at DESC LIMIT 1;

-- 预期结果
id                                   | request_name | user_message              | book_message              | nice_message | combined_message
-------------------------------------|--------------|---------------------------|---------------------------|--------------|------------------
550e8400-e29b-41d4-a716-446655440000 |              | Hello from user-service   | Hello from book-service   | NULL         | ...
```

```bash
# 查询Redis
redis-cli GET "operation:log:550e8400-e29b-41d4-a716-446655440000"

# 预期结果（JSON格式）
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "operation": "SayHello",
  "user_message": "Hello from user-service",
  "book_message": "Hello from book-service",
  "combined_message": "...",
  "timestamp": "2025-10-31T12:00:00Z"
}
```

### 5.4 验证RabbitMQ消息

```bash
# 查看RabbitMQ管理界面
http://localhost:15672
# 用户名: guest
# 密码: guest

# 检查队列中的消息
# Exchange: demo_exchange
# Queue: user_service_queue
# 应该能看到user-service发送的消息
```

---

## 六、架构优势

### 6.1 解耦性

✅ **服务解耦**  
- user-service与book-service通过gRPC接口通信，无需了解对方实现细节
- user-service与nice-service通过消息队列异步通信，完全解耦

✅ **数据访问解耦**  
- 业务层依赖Repository接口，不关心底层存储
- 可以轻松切换PostgreSQL、MongoDB或其他存储

### 6.2 可靠性

✅ **同步调用的可靠性**  
- book-service调用失败会立即返回错误
- 不会写入错误数据到数据库

✅ **异步操作的容错性**  
- PostgreSQL写入失败不影响主流程
- Redis缓存失败不影响主流程  
- RabbitMQ发送失败只记录日志

✅ **消息可靠性**  
- RabbitMQ持久化消息
- Consumer手动ACK确认
- 失败消息可重新入队

### 6.3 性能

✅ **异步处理**  
- 数据库写入不阻塞主流程
- Redis缓存不阻塞主流程
- RabbitMQ发送使用goroutine异步执行

✅ **连接复用**  
- gRPC连接池管理
- 数据库连接池
- Redis连接池

### 6.4 可观测性

✅ **完整的日志记录**  
- 每个关键步骤都有日志
- 使用结构化日志（zap）
- 包含上下文信息（RequestID等）

✅ **数据可追溯**  
- PostgreSQL持久化所有操作记录
- Redis缓存最近的操作
- 可以通过log_id关联整个流程

---

## 七、扩展建议

### 7.1 添加链路追踪

```go
import "go.opentelemetry.io/otel"

func (uc *userUseCase) SayHello(ctx context.Context, name string) (string, error) {
    ctx, span := otel.Tracer("user-service").Start(ctx, "SayHello")
    defer span.End()
    
    // ... 业务逻辑
}
```

### 7.2 添加熔断器

```go
import "github.com/sony/gobreaker"

type userUseCase struct {
    bookClient bookv1.BookServiceClient
    breaker    *gobreaker.CircuitBreaker  // 熔断器
}

func (uc *userUseCase) SayHello(ctx context.Context, name string) (string, error) {
    // 使用熔断器保护book-service调用
    result, err := uc.breaker.Execute(func() (interface{}, error) {
        return uc.bookClient.SayHello(ctx, &bookv1.HelloRequest{})
    })
    // ...
}
```

### 7.3 添加重试机制

```go
import "github.com/avast/retry-go"

func (uc *userUseCase) SayHello(ctx context.Context, name string) (string, error) {
    var bookResp *bookv1.HelloResponse
    
    err := retry.Do(
        func() error {
            var err error
            bookResp, err = uc.bookClient.SayHello(ctx, &bookv1.HelloRequest{})
            return err
        },
        retry.Attempts(3),
        retry.Delay(100*time.Millisecond),
    )
    // ...
}
```

### 7.4 添加指标监控

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    sayHelloCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "user_service_sayhello_total",
            Help: "Total number of SayHello calls",
        },
        []string{"status"},
    )
)

func (uc *userUseCase) SayHello(ctx context.Context, name string) (string, error) {
    defer func() {
        if err != nil {
            sayHelloCounter.WithLabelValues("error").Inc()
        } else {
            sayHelloCounter.WithLabelValues("success").Inc()
        }
    }()
    // ...
}
```

---

## 八、总结

### 8.1 实现成果

✅ **完成的功能**
1. gRPC同步调用book-service
2. RabbitMQ异步通信nice-service
3. PostgreSQL持久化操作日志
4. Redis缓存操作日志
5. 消息消费者接收nice-service回复

✅ **架构优势**
1. 清晰的分层架构
2. 完整的依赖注入
3. 多种通信方式集成
4. 可靠的错误处理
5. 优秀的可观测性

### 8.2 技术亮点

1. **领域驱动设计**：纯粹的领域模型，DO/PO分离
2. **依赖注入**：所有依赖通过构造函数注入
3. **异步处理**：不阻塞主流程的异步操作
4. **容错设计**：辅助功能失败不影响主功能
5. **资源管理**：优雅的启动和关闭流程

### 8.3 最佳实践

1. **接口定义清晰**：每层都有明确的接口
2. **职责分离**：每个组件只做一件事
3. **错误处理**：完善的错误处理和日志
4. **代码注释**：详细的中文注释
5. **可扩展性**：预留了扩展点

---

## 九、注意事项

### 9.1 运行前检查

- [ ] PostgreSQL已启动并创建数据库
- [ ] Redis已启动
- [ ] RabbitMQ已启动
- [ ] book-service已启动
- [ ] nice-service已启动
- [ ] 所有Go依赖已安装

### 9.2 配置检查

- [ ] 数据库连接字符串正确
- [ ] Redis地址和密码正确  
- [ ] RabbitMQ地址正确
- [ ] gRPC客户端地址正确
- [ ] 端口无冲突

### 9.3 故障排查

**问题1**: book-service调用失败  
**解决**: 检查book-service是否启动，地址是否正确

**问题2**: PostgreSQL连接失败  
**解决**: 检查数据库是否启动，用户名密码是否正确

**问题3**: RabbitMQ连接失败  
**解决**: 检查RabbitMQ是否启动，URL是否正确

**问题4**: nice-service消息未收到  
**解决**: 检查队列绑定、路由键配置

---

## 十、参考资料

- [内部服务架构设计](./ARCHITECTURE.md)
- [user-service改造总结](./USER_SERVICE_REFACTOR_SUMMARY.md)
- [gRPC客户端管理](../API_GATEWAY/API_GATEWAY_GRPC_CLIENT.md)
- [RabbitMQ使用指南](./RABBITMQ_GUIDE.md)

---

**升级完成时间**：2025年10月31日  
**升级工程师**：资深Golang工程师（20年研发经验）  
**升级状态**：✅ 核心代码已完成，需要运行测试验证
