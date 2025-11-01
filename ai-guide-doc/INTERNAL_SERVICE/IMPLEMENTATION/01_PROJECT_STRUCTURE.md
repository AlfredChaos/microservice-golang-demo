# 项目结构和文件清单

> 完整的目录结构和所需创建的文件清单

## 创建目录结构

### 步骤1：创建基础目录

```bash
# 假设创建order-service服务
SERVICE_NAME="order-service"

# 创建内部服务目录
mkdir -p internal/${SERVICE_NAME}/{domain,data,biz,service,server,conf,consumer,migrations}

# 创建命令目录
mkdir -p cmd/${SERVICE_NAME}

# 创建API目录（如果还没有）
mkdir -p api/order/v1

# 创建配置目录（如果还没有）
mkdir -p configs
```

## 完整目录结构

```
microservice-golang-demo/
│
├── api/                                    # gRPC接口定义（共享）
│   └── order/
│       └── v1/
│           ├── order.proto                 # Protobuf定义
│           ├── order.pb.go                 # 生成的Go代码
│           └── order_grpc.pb.go            # 生成的gRPC代码
│
├── cmd/                                    # 服务启动入口
│   └── order-service/
│       └── main.go                         # 主函数
│
├── internal/                               # 内部代码
│   └── order-service/
│       ├── domain/                         # 领域层
│       │   ├── order.go                    # 订单实体
│       │   ├── order_item.go               # 订单项
│       │   └── errors.go                   # 领域错误
│       │
│       ├── data/                           # 数据访问层
│       │   ├── data.go                     # 数据层容器
│       │   ├── order_repo.go               # 仓库接口
│       │   ├── order_pg_repo.go            # PostgreSQL实现
│       │   ├── order_mongo_repo.go         # MongoDB实现（可选）
│       │   ├── order_cache.go              # Redis缓存（可选）
│       │   └── order_cached_repo.go        # 带缓存的仓库（可选）
│       │
│       ├── biz/                            # 业务逻辑层
│       │   └── order_usecase.go            # 订单业务逻辑
│       │
│       ├── service/                        # 服务层
│       │   └── order_service.go            # gRPC服务实现
│       │
│       ├── server/                         # 服务器层
│       │   └── grpc.go                     # gRPC服务器
│       │
│       ├── conf/                           # 配置层
│       │   └── config.go                   # 配置结构
│       │
│       ├── consumer/                       # 消息消费者（可选）
│       │   └── order_consumer.go           # 订单消息消费者
│       │
│       └── migrations/                     # 数据库迁移
│           └── 001_create_orders_table.sql # 迁移SQL
│
├── configs/                                # 配置文件
│   ├── order-service.yaml                  # 开发环境配置
│   └── order-service.prod.yaml             # 生产环境配置（可选）
│
├── pkg/                                    # 公共包（已存在）
│   ├── config/                             # 配置加载
│   ├── log/                                # 日志
│   ├── db/                                 # 数据库
│   ├── cache/                              # 缓存
│   ├── mq/                                 # 消息队列
│   ├── grpcclient/                         # gRPC客户端管理
│   └── errors/                             # 错误处理
│
├── scripts/                                # 脚本（已存在）
│   ├── gen-proto.sh                        # Proto生成脚本
│   └── gen-swagger.sh                      # Swagger生成脚本
│
├── go.mod                                  # Go模块文件
├── go.sum                                  # 依赖校验文件
├── Makefile                                # Make命令
└── README.md                               # 项目说明
```

## 文件清单和说明

### 1. API层文件（共享）

#### api/order/v1/order.proto
```protobuf
syntax = "proto3";
package order.v1;
option go_package = "github.com/alfredchaos/demo/api/order/v1;orderv1";

service OrderService {
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse);
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
}

message CreateOrderRequest { ... }
message CreateOrderResponse { ... }
// ... 其他消息定义
```

**说明**：
- 定义gRPC服务接口
- 所有服务共享此定义
- 使用 `./scripts/gen-proto.sh` 生成Go代码

---

### 2. Domain层文件（领域模型）

#### internal/order-service/domain/order.go
```go
package domain

// Order 订单领域模型
type Order struct {
    ID          string
    UserID      string
    Items       []OrderItem
    TotalAmount float64
    Status      OrderStatus
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// OrderStatus 订单状态枚举
type OrderStatus string

const (
    OrderStatusPending   OrderStatus = "pending"
    OrderStatusConfirmed OrderStatus = "confirmed"
    // ... 其他状态
)

// NewOrder 创建订单（工厂函数）
func NewOrder(...) *Order { }

// Validate 验证订单
func (o *Order) Validate() error { }

// Confirm 确认订单（业务方法）
func (o *Order) Confirm() error { }
```

**说明**：
- 定义核心业务实体
- 包含业务规则和验证
- 不依赖任何外部框架

#### internal/order-service/domain/order_item.go
```go
package domain

// OrderItem 订单项
type OrderItem struct {
    ProductID   string
    ProductName string
    Quantity    int32
    Price       float64
}

// Subtotal 计算小计
func (item *OrderItem) Subtotal() float64 {
    return item.Price * float64(item.Quantity)
}
```

**说明**：
- 值对象定义
- 可包含计算方法

#### internal/order-service/domain/errors.go
```go
package domain

import "errors"

var (
    ErrOrderNotFound      = errors.New("order not found")
    ErrInvalidOrder       = errors.New("invalid order")
    ErrInvalidUserID      = errors.New("invalid user id")
    ErrEmptyOrderItems    = errors.New("order items cannot be empty")
    ErrInvalidQuantity    = errors.New("invalid quantity")
    ErrInvalidOrderStatus = errors.New("invalid order status")
)
```

**说明**：
- 定义领域错误
- 使用标准error类型

---

### 3. Data层文件（数据访问）

#### internal/order-service/data/data.go
```go
package data

// Data 数据访问层容器
type Data struct {
    pgDB        *gorm.DB
    mongoClient *db.MongoClient
    redisClient *cache.RedisClient
    mqClient    *mq.RabbitMQClient
    
    // 仓库实例（导出）
    OrderRepo OrderRepository
}

// NewData 创建数据访问层
func NewData(...) (*Data, error) { }

// Close 关闭所有连接
func (d *Data) Close(ctx context.Context) error { }
```

**说明**：
- 管理所有数据源
- 创建仓库实例
- 负责资源清理

#### internal/order-service/data/order_repo.go
```go
package data

// OrderRepository 订单仓库接口
type OrderRepository interface {
    Create(ctx context.Context, order *domain.Order) error
    GetByID(ctx context.Context, id string) (*domain.Order, error)
    ListByUserID(ctx context.Context, userID string, offset, limit int) ([]*domain.Order, error)
    Update(ctx context.Context, order *domain.Order) error
    Delete(ctx context.Context, id string) error
}
```

**说明**：
- 定义数据访问接口
- 业务层依赖此接口

#### internal/order-service/data/order_pg_repo.go
```go
package data

// OrderPO 订单持久化对象
type OrderPO struct {
    ID          string    `gorm:"column:id;primaryKey"`
    UserID      string    `gorm:"column:user_id;index"`
    Items       string    `gorm:"column:items;type:jsonb"`
    TotalAmount float64   `gorm:"column:total_amount"`
    Status      string    `gorm:"column:status;index"`
    CreatedAt   time.Time `gorm:"column:created_at"`
    UpdatedAt   time.Time `gorm:"column:updated_at"`
}

// TableName 表名
func (OrderPO) TableName() string { return "orders" }

// ToDomain 转换为领域对象
func (po *OrderPO) ToDomain() (*domain.Order, error) { }

// FromDomain 从领域对象转换
func (po *OrderPO) FromDomain(order *domain.Order) error { }

// orderPgRepository PostgreSQL仓库实现
type orderPgRepository struct {
    db *gorm.DB
}

// NewOrderPgRepository 创建PostgreSQL仓库
func NewOrderPgRepository(db *gorm.DB) OrderRepository { }

// 实现接口方法
func (r *orderPgRepository) Create(...) error { }
func (r *orderPgRepository) GetByID(...) (*domain.Order, error) { }
// ... 其他方法
```

**说明**：
- PO（持久化对象）与DO（领域对象）分离
- 实现Repository接口
- 处理数据库操作

#### internal/order-service/data/order_mongo_repo.go（可选）
```go
package data

// UserMongoPO MongoDB持久化对象
type OrderMongoPO struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    UserID    string             `bson:"user_id"`
    // ... 其他字段
}

// orderMongoRepository MongoDB仓库实现
type orderMongoRepository struct {
    collection *mongo.Collection
}

// NewOrderMongoRepository 创建MongoDB仓库
func NewOrderMongoRepository(client *db.MongoClient) OrderRepository { }
```

**说明**：
- MongoDB实现（可选）
- 与PostgreSQL实现相同的接口

#### internal/order-service/data/order_cache.go（可选）
```go
package data

// OrderCache 订单缓存
type OrderCache struct {
    redis *cache.RedisClient
}

// NewOrderCache 创建订单缓存
func NewOrderCache(redis *cache.RedisClient) *OrderCache { }

func (c *OrderCache) Get(ctx context.Context, id string) (*domain.Order, error) { }
func (c *OrderCache) Set(ctx context.Context, order *domain.Order) error { }
func (c *OrderCache) Delete(ctx context.Context, id string) error { }
```

**说明**：
- Redis缓存封装
- 提供Get/Set/Delete方法

#### internal/order-service/data/order_cached_repo.go（可选）
```go
package data

// orderCachedRepository 带缓存的仓库实现
type orderCachedRepository struct {
    repo  OrderRepository
    cache *OrderCache
}

// NewOrderCachedRepository 创建带缓存的仓库
func NewOrderCachedRepository(repo OrderRepository, cache *OrderCache) OrderRepository { }

// 实现接口，加入缓存逻辑
func (r *orderCachedRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
    // 1. 尝试从缓存获取
    // 2. 缓存未命中，从数据库获取
    // 3. 写入缓存
}
```

**说明**：
- 装饰器模式
- 透明地添加缓存层

---

### 4. Biz层文件（业务逻辑）

#### internal/order-service/biz/order_usecase.go
```go
package biz

// OrderUseCase 订单业务逻辑接口
type OrderUseCase interface {
    CreateOrder(ctx context.Context, userID string, items []domain.OrderItem) (*domain.Order, error)
    GetOrder(ctx context.Context, id string) (*domain.Order, error)
    ListOrders(ctx context.Context, userID string, offset, limit int) ([]*domain.Order, error)
    ConfirmOrder(ctx context.Context, id string) error
    CancelOrder(ctx context.Context, id string) error
}

// orderUseCase 订单业务逻辑实现
type orderUseCase struct {
    orderRepo   data.OrderRepository
    userClient  userv1.UserServiceClient  // 可选：调用用户服务
    publisher   mq.Publisher              // 可选：发布消息
}

// NewOrderUseCase 创建订单业务逻辑
func NewOrderUseCase(
    orderRepo data.OrderRepository,
    userClient userv1.UserServiceClient,
    publisher mq.Publisher,
) OrderUseCase { }

// CreateOrder 创建订单
func (uc *orderUseCase) CreateOrder(...) (*domain.Order, error) {
    // 1. 验证用户（可选：调用user-service）
    // 2. 创建领域对象
    // 3. 验证业务规则
    // 4. 持久化
    // 5. 发布事件（可选）
}

// GetOrder 获取订单
func (uc *orderUseCase) GetOrder(...) (*domain.Order, error) { }

// 其他业务方法...
```

**说明**：
- 定义业务逻辑接口
- 编排领域对象和数据访问
- 可集成gRPC客户端和消息队列

---

### 5. Service层文件（gRPC服务）

#### internal/order-service/service/order_service.go
```go
package service

// OrderService gRPC服务实现
type OrderService struct {
    orderv1.UnimplementedOrderServiceServer
    useCase biz.OrderUseCase
}

// NewOrderService 创建订单服务
func NewOrderService(useCase biz.OrderUseCase) *OrderService { }

// CreateOrder 实现CreateOrder接口
func (s *OrderService) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
    // 1. 转换请求（Proto -> Domain）
    // 2. 调用业务逻辑
    // 3. 转换响应（Domain -> Proto）
}

// GetOrder 实现GetOrder接口
func (s *OrderService) GetOrder(...) (*orderv1.GetOrderResponse, error) { }

// ListOrders 实现ListOrders接口
func (s *OrderService) ListOrders(...) (*orderv1.ListOrdersResponse, error) { }

// 辅助方法：Proto <-> Domain 转换
func (s *OrderService) toProtoOrder(order *domain.Order) *orderv1.Order { }
func (s *OrderService) toDomainOrderItem(item *orderv1.OrderItem) domain.OrderItem { }
```

**说明**：
- 实现gRPC接口
- 只做协议转换
- 不包含业务逻辑

---

### 6. Server层文件（gRPC服务器）

#### internal/order-service/server/grpc.go
```go
package server

// GRPCServer gRPC服务器
type GRPCServer struct {
    server       *grpc.Server
    config       *conf.ServerConfig
    orderService *service.OrderService
}

// NewGRPCServer 创建gRPC服务器
func NewGRPCServer(cfg *conf.ServerConfig, orderService *service.OrderService) *GRPCServer {
    // 创建gRPC服务器，添加拦截器
    server := grpc.NewServer(
        grpc.ChainUnaryInterceptor(
            // 日志拦截器
            // 恢复拦截器
            // 认证拦截器（可选）
        ),
    )
    
    // 注册服务
    orderv1.RegisterOrderServiceServer(server, orderService)
    
    // 注册反射服务（用于grpcurl）
    reflection.Register(server)
    
    return &GRPCServer{...}
}

// Start 启动服务器
func (s *GRPCServer) Start() error { }

// Stop 停止服务器
func (s *GRPCServer) Stop() { }
```

**说明**：
- 配置gRPC服务器
- 注册服务和拦截器
- 管理服务器生命周期

---

### 7. Conf层文件（配置）

#### internal/order-service/conf/config.go
```go
package conf

// Config 服务配置
type Config struct {
    Server      ServerConfig           `yaml:"server" mapstructure:"server"`
    Log         log.LogConfig          `yaml:"log" mapstructure:"log"`
    Database    db.DatabaseConfig      `yaml:"database" mapstructure:"database"`
    MongoDB     db.MongoConfig         `yaml:"mongodb" mapstructure:"mongodb"`
    Redis       cache.RedisConfig      `yaml:"redis" mapstructure:"redis"`
    RabbitMQ    mq.RabbitMQConfig      `yaml:"rabbitmq" mapstructure:"rabbitmq"`
    GRPCClients grpcclient.Config      `yaml:"grpc_clients" mapstructure:"grpc_clients"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
    Name string `yaml:"name" mapstructure:"name"`
    Host string `yaml:"host" mapstructure:"host"`
    Port int    `yaml:"port" mapstructure:"port"`
}
```

**说明**：
- 定义配置结构
- 使用公共配置类型

---

### 8. Consumer层文件（消息消费者，可选）

#### internal/order-service/consumer/order_consumer.go
```go
package consumer

// OrderConsumer 订单消息消费者
type OrderConsumer struct {
    mqClient     *mq.RabbitMQClient
    orderUseCase biz.OrderUseCase
}

// NewOrderConsumer 创建订单消费者
func NewOrderConsumer(mqClient *mq.RabbitMQClient, orderUseCase biz.OrderUseCase) *OrderConsumer { }

// Start 开始消费消息
func (c *OrderConsumer) Start(ctx context.Context) error {
    // 1. 订阅队列
    // 2. 处理消息
    // 3. 调用业务逻辑
}

func (c *OrderConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
    // 解析消息
    // 调用业务逻辑
    // 确认消息
}
```

**说明**：
- 消费RabbitMQ消息
- 调用业务逻辑处理

---

### 9. 主函数文件

#### cmd/order-service/main.go
```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    
    // 导入所有需要的包...
)

func init() {
    // 注册gRPC客户端工厂（如需调用其他服务）
    grpcclient.GlobalRegistry.Register("user-service", func(conn *grpc.ClientConn) interface{} {
        return userv1.NewUserServiceClient(conn)
    })
}

func main() {
    // 1. 加载配置
    var cfg conf.Config
    config.MustLoadConfig("order-service", &cfg)
    
    // 2. 初始化日志
    log.MustInitLogger(&cfg.Log, cfg.Server.Name)
    defer log.Sync()
    
    // 3. 初始化基础设施
    // ... PostgreSQL, MongoDB, Redis, RabbitMQ
    
    // 4. 初始化gRPC客户端管理器（可选）
    // ...
    
    // 5. 初始化数据访问层
    dataLayer, err := data.NewData(...)
    if err != nil {
        log.Fatal("failed to initialize data layer", zap.Error(err))
    }
    defer dataLayer.Close(context.Background())
    
    // 6. 初始化业务逻辑层
    orderUseCase := biz.NewOrderUseCase(...)
    
    // 7. 初始化服务层
    orderService := service.NewOrderService(orderUseCase)
    
    // 8. 初始化gRPC服务器
    grpcServer := server.NewGRPCServer(&cfg.Server, orderService)
    
    // 9. 启动消费者（可选）
    // consumer := consumer.NewOrderConsumer(...)
    // consumer.Start(context.Background())
    
    // 10. 启动服务器
    go func() {
        if err := grpcServer.Start(); err != nil {
            log.Fatal("failed to start grpc server", zap.Error(err))
        }
    }()
    
    // 11. 优雅关闭
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Info("shutting down order-service")
    grpcServer.Stop()
    log.Info("order-service stopped")
}
```

**说明**：
- 服务启动入口
- 完整的依赖注入流程
- 优雅关闭处理

---

### 10. 配置文件

#### configs/order-service.yaml
```yaml
server:
  name: order-service
  host: 0.0.0.0
  port: 9003

log:
  level: debug
  format: console
  output_paths:
    - stdout
  enable_console_writer: true

database:
  enabled: true
  driver: postgres
  host: localhost
  port: 5432
  username: postgres
  password: password
  database: order_service
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600
  log_level: info

mongodb:
  enabled: false
  uri: mongodb://admin:password@localhost:27017
  database: order_service

redis:
  enabled: false
  addr: localhost:6379
  password: ""
  db: 0
  pool_size: 10

rabbitmq:
  enabled: false
  url: amqp://guest:guest@localhost:5672/
  exchange: demo_exchange
  exchange_type: topic
  queue: order_service_queue
  routing_key: order.#

# 可选：需要调用其他服务时
grpc_clients:
  services:
    - name: user-service
      address: localhost:9001
      timeout: 5s
      retry:
        max: 3
        timeout: 10s
        backoff: 100ms
```

**说明**：
- YAML格式配置
- 支持多环境配置

---

### 11. 数据库迁移文件

#### internal/order-service/migrations/001_create_orders_table.sql
```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    items JSONB NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
```

**说明**：
- 使用Goose格式
- 包含Up和Down迁移

---

## 快速创建脚本

### create-service.sh
```bash
#!/bin/bash

# 使用方法: ./create-service.sh order-service

SERVICE_NAME=$1

if [ -z "$SERVICE_NAME" ]; then
    echo "使用方法: ./create-service.sh <service-name>"
    exit 1
fi

echo "创建服务: $SERVICE_NAME"

# 创建目录
mkdir -p internal/${SERVICE_NAME}/{domain,data,biz,service,server,conf,consumer,migrations}
mkdir -p cmd/${SERVICE_NAME}
mkdir -p api/$(echo $SERVICE_NAME | sed 's/-service//')/v1
mkdir -p configs

# 创建占位文件
touch internal/${SERVICE_NAME}/domain/.gitkeep
touch internal/${SERVICE_NAME}/data/.gitkeep
touch internal/${SERVICE_NAME}/biz/.gitkeep
touch internal/${SERVICE_NAME}/service/.gitkeep
touch internal/${SERVICE_NAME}/server/.gitkeep
touch internal/${SERVICE_NAME}/conf/.gitkeep
touch internal/${SERVICE_NAME}/consumer/.gitkeep
touch internal/${SERVICE_NAME}/migrations/.gitkeep
touch cmd/${SERVICE_NAME}/.gitkeep

echo "✅ 目录结构创建完成"
echo "📝 下一步: 参照编码方案文档实现各层代码"
```

## 文件创建顺序建议

### 阶段1：接口和配置
1. ✅ `api/order/v1/order.proto`
2. ✅ `configs/order-service.yaml`
3. ✅ `internal/order-service/conf/config.go`

### 阶段2：领域层（无依赖）
4. ✅ `internal/order-service/domain/errors.go`
5. ✅ `internal/order-service/domain/order_item.go`
6. ✅ `internal/order-service/domain/order.go`

### 阶段3：数据层（依赖Domain）
7. ✅ `internal/order-service/data/order_repo.go`
8. ✅ `internal/order-service/data/order_pg_repo.go`
9. ⚠️ `internal/order-service/data/order_cache.go`（可选）
10. ⚠️ `internal/order-service/data/order_cached_repo.go`（可选）
11. ✅ `internal/order-service/data/data.go`

### 阶段4：业务层（依赖Data接口）
12. ✅ `internal/order-service/biz/order_usecase.go`

### 阶段5：服务层（依赖Biz接口）
13. ✅ `internal/order-service/service/order_service.go`

### 阶段6：服务器层（依赖Service）
14. ✅ `internal/order-service/server/grpc.go`

### 阶段7：启动入口（组装所有层）
15. ✅ `cmd/order-service/main.go`

### 阶段8：数据库迁移
16. ✅ `internal/order-service/migrations/001_create_orders_table.sql`

### 阶段9：可选功能
17. ⚠️ `internal/order-service/consumer/order_consumer.go`

## 检查清单

### 必需文件
- [ ] Proto文件已创建并生成Go代码
- [ ] 配置文件已创建
- [ ] Domain层文件已创建
- [ ] Data层接口和实现已创建
- [ ] Biz层已创建
- [ ] Service层已创建
- [ ] Server层已创建
- [ ] 主函数已创建
- [ ] 数据库迁移文件已创建

### 可选文件
- [ ] MongoDB仓库实现（如需要）
- [ ] Redis缓存实现（如需要）
- [ ] 消息消费者实现（如需要）
- [ ] 生产环境配置文件（如需要）

## 下一步

项目结构创建完成后，请按以下顺序实现代码：

1. 📋 [领域层实现](./02_DOMAIN_LAYER.md)
2. 💾 [数据层实现](./03_DATA_LAYER.md)
3. 💼 [业务层实现](./04_BIZ_LAYER.md)
4. 🔌 [服务层实现](./05_SERVICE_LAYER.md)
5. 🖥️ [服务器层实现](./06_SERVER_LAYER.md)
6. ⚙️ [配置实现](./07_CONFIGURATION.md)
7. 🚀 [主函数实现](./08_MAIN_ENTRY.md)

---

**提示**: 
- 使用提供的快速创建脚本可以快速搭建目录结构
- 按照推荐的文件创建顺序可以避免依赖问题
- 参考现有的user-service或book-service示例代码
