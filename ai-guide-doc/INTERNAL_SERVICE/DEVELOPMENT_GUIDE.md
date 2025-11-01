# 内部服务开发指南

## 概述

本文档指导开发者如何创建一个新的内部服务。我们将以创建 `order-service`（订单服务）为例，演示完整的开发流程。

---

## 开发流程概览

```
1. 定义gRPC接口（Proto文件）
   ↓
2. 生成Go代码
   ↓
3. 创建服务目录结构
   ↓
4. 实现领域层（Domain）
   ↓
5. 实现数据访问层（Data）
   ↓
6. 实现业务逻辑层（Biz）
   ↓
7. 实现服务层（Service）
   ↓
8. 实现服务器层（Server）
   ↓
9. 配置和启动（Main）
   ↓
10. 数据库迁移（可选）
   ↓
11. 测试
```

---

## 步骤详解

### 步骤1：定义gRPC接口

创建Proto文件定义服务接口。

```bash
mkdir -p api/order/v1
```

```protobuf
// api/order/v1/order.proto
syntax = "proto3";

package order.v1;

option go_package = "github.com/alfredchaos/demo/api/order/v1;orderv1";

// OrderService 订单服务接口
service OrderService {
  // CreateOrder 创建订单
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
  
  // GetOrder 获取订单
  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse);
  
  // ListOrders 列出订单
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
}

// CreateOrderRequest 创建订单请求
message CreateOrderRequest {
  string user_id = 1;
  repeated OrderItem items = 2;
}

// CreateOrderResponse 创建订单响应
message CreateOrderResponse {
  Order order = 1;
}

// GetOrderRequest 获取订单请求
message GetOrderRequest {
  string id = 1;
}

// GetOrderResponse 获取订单响应
message GetOrderResponse {
  Order order = 1;
}

// ListOrdersRequest 列出订单请求
message ListOrdersRequest {
  string user_id = 1;
  int32 offset = 2;
  int32 limit = 3;
}

// ListOrdersResponse 列出订单响应
message ListOrdersResponse {
  repeated Order orders = 1;
  int32 total = 2;
}

// Order 订单信息
message Order {
  string id = 1;
  string user_id = 2;
  repeated OrderItem items = 3;
  double total_amount = 4;
  string status = 5;
  int64 created_at = 6;
  int64 updated_at = 7;
}

// OrderItem 订单项
message OrderItem {
  string product_id = 1;
  string product_name = 2;
  int32 quantity = 3;
  double price = 4;
}
```

**生成Go代码**：

```bash
./scripts/gen-proto.sh
```

---

### 步骤2：创建服务目录结构

```bash
# 创建目录
mkdir -p internal/order-service/{domain,data,biz,service,server,conf,consumer}
mkdir -p internal/order-service/migrations
mkdir -p cmd/order-service
mkdir -p configs
```

**完整目录结构**：

```
internal/order-service/
├── domain/              # 领域层
│   ├── order.go
│   └── errors.go
├── data/                # 数据访问层
│   ├── data.go
│   ├── order_repo.go
│   └── order_pg_repo.go
├── biz/                 # 业务逻辑层
│   └── order_usecase.go
├── service/             # 服务层
│   └── order_service.go
├── server/              # 服务器层
│   ├── grpc.go
│   └── client.go
├── conf/                # 配置层
│   └── config.go
├── consumer/            # 消费者（可选）
│   └── message_consumer.go
└── migrations/          # 数据库迁移
    └── 001_create_orders_table.sql

cmd/order-service/
└── main.go              # 启动入口

configs/
└── order-service.yaml   # 配置文件
```

---

### 步骤3：实现领域层

```go
// internal/order-service/domain/order.go
package domain

import (
    "time"
)

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

// OrderItem 订单项
type OrderItem struct {
    ProductID   string
    ProductName string
    Quantity    int32
    Price       float64
}

// OrderStatus 订单状态
type OrderStatus string

const (
    OrderStatusPending   OrderStatus = "pending"
    OrderStatusConfirmed OrderStatus = "confirmed"
    OrderStatusShipped   OrderStatus = "shipped"
    OrderStatusDelivered OrderStatus = "delivered"
    OrderStatusCancelled OrderStatus = "cancelled"
)

// NewOrder 创建新订单
func NewOrder(userID string, items []OrderItem) *Order {
    now := time.Now()
    
    // 计算总金额
    var totalAmount float64
    for _, item := range items {
        totalAmount += item.Price * float64(item.Quantity)
    }
    
    return &Order{
        UserID:      userID,
        Items:       items,
        TotalAmount: totalAmount,
        Status:      OrderStatusPending,
        CreatedAt:   now,
        UpdatedAt:   now,
    }
}

// Validate 验证订单
func (o *Order) Validate() error {
    if o.UserID == "" {
        return ErrInvalidUserID
    }
    if len(o.Items) == 0 {
        return ErrEmptyOrderItems
    }
    for _, item := range o.Items {
        if item.Quantity <= 0 {
            return ErrInvalidQuantity
        }
        if item.Price < 0 {
            return ErrInvalidPrice
        }
    }
    return nil
}

// Confirm 确认订单
func (o *Order) Confirm() error {
    if o.Status != OrderStatusPending {
        return ErrInvalidOrderStatus
    }
    o.Status = OrderStatusConfirmed
    o.UpdatedAt = time.Now()
    return nil
}
```

```go
// internal/order-service/domain/errors.go
package domain

import "errors"

var (
    ErrOrderNotFound       = errors.New("order not found")
    ErrInvalidUserID       = errors.New("invalid user id")
    ErrEmptyOrderItems     = errors.New("order items cannot be empty")
    ErrInvalidQuantity     = errors.New("invalid quantity")
    ErrInvalidPrice        = errors.New("invalid price")
    ErrInvalidOrderStatus  = errors.New("invalid order status")
    ErrOrderAlreadyExists  = errors.New("order already exists")
)
```

---

### 步骤4：实现数据访问层

**定义仓库接口**：

```go
// internal/order-service/data/order_repo.go
package data

import (
    "context"
    "github.com/alfredchaos/demo/internal/order-service/domain"
)

// OrderRepository 订单仓库接口
type OrderRepository interface {
    Create(ctx context.Context, order *domain.Order) error
    GetByID(ctx context.Context, id string) (*domain.Order, error)
    ListByUserID(ctx context.Context, userID string, offset, limit int) ([]*domain.Order, error)
    Update(ctx context.Context, order *domain.Order) error
    Delete(ctx context.Context, id string) error
}
```

**PostgreSQL实现**：

```go
// internal/order-service/data/order_pg_repo.go
package data

import (
    "context"
    "encoding/json"
    "errors"
    "time"
    
    "github.com/alfredchaos/demo/internal/order-service/domain"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

// OrderPO 订单持久化对象
type OrderPO struct {
    ID          string    `gorm:"column:id;primaryKey"`
    UserID      string    `gorm:"column:user_id;index;not null"`
    Items       string    `gorm:"column:items;type:jsonb"`
    TotalAmount float64   `gorm:"column:total_amount"`
    Status      string    `gorm:"column:status;index"`
    CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (OrderPO) TableName() string {
    return "orders"
}

// ToDomain 转换为领域对象
func (po *OrderPO) ToDomain() (*domain.Order, error) {
    var items []domain.OrderItem
    if err := json.Unmarshal([]byte(po.Items), &items); err != nil {
        return nil, err
    }
    
    return &domain.Order{
        ID:          po.ID,
        UserID:      po.UserID,
        Items:       items,
        TotalAmount: po.TotalAmount,
        Status:      domain.OrderStatus(po.Status),
        CreatedAt:   po.CreatedAt,
        UpdatedAt:   po.UpdatedAt,
    }, nil
}

// FromDomain 从领域对象转换
func (po *OrderPO) FromDomain(order *domain.Order) error {
    itemsJSON, err := json.Marshal(order.Items)
    if err != nil {
        return err
    }
    
    po.ID = order.ID
    po.UserID = order.UserID
    po.Items = string(itemsJSON)
    po.TotalAmount = order.TotalAmount
    po.Status = string(order.Status)
    po.CreatedAt = order.CreatedAt
    po.UpdatedAt = order.UpdatedAt
    
    return nil
}

// orderPgRepository PostgreSQL仓库实现
type orderPgRepository struct {
    db *gorm.DB
}

// NewOrderPgRepository 创建PostgreSQL仓库
func NewOrderPgRepository(db *gorm.DB) OrderRepository {
    return &orderPgRepository{db: db}
}

func (r *orderPgRepository) Create(ctx context.Context, order *domain.Order) error {
    if order.ID == "" {
        order.ID = uuid.New().String()
    }
    
    po := &OrderPO{}
    if err := po.FromDomain(order); err != nil {
        return err
    }
    
    return r.db.WithContext(ctx).Create(po).Error
}

func (r *orderPgRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
    var po OrderPO
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&po).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, domain.ErrOrderNotFound
        }
        return nil, err
    }
    return po.ToDomain()
}

func (r *orderPgRepository) ListByUserID(ctx context.Context, userID string, offset, limit int) ([]*domain.Order, error) {
    var pos []OrderPO
    err := r.db.WithContext(ctx).
        Where("user_id = ?", userID).
        Offset(offset).
        Limit(limit).
        Order("created_at DESC").
        Find(&pos).Error
    if err != nil {
        return nil, err
    }
    
    orders := make([]*domain.Order, 0, len(pos))
    for _, po := range pos {
        order, err := po.ToDomain()
        if err != nil {
            return nil, err
        }
        orders = append(orders, order)
    }
    
    return orders, nil
}

func (r *orderPgRepository) Update(ctx context.Context, order *domain.Order) error {
    po := &OrderPO{}
    if err := po.FromDomain(order); err != nil {
        return err
    }
    return r.db.WithContext(ctx).Save(po).Error
}

func (r *orderPgRepository) Delete(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Delete(&OrderPO{}, "id = ?", id).Error
}
```

**数据层容器**：

```go
// internal/order-service/data/data.go
package data

import (
    "context"
    "fmt"
    
    "gorm.io/gorm"
    "github.com/alfredchaos/demo/pkg/cache"
    "github.com/alfredchaos/demo/pkg/db"
    "github.com/alfredchaos/demo/pkg/mq"
)

// Data 数据访问层容器
type Data struct {
    pgDB        *gorm.DB
    mongoClient *db.MongoClient
    redisClient *cache.RedisClient
    mqClient    *mq.RabbitMQClient
    
    OrderRepo OrderRepository
}

// NewData 创建数据访问层
func NewData(
    pgDB *gorm.DB,
    mongoClient *db.MongoClient,
    redisClient *cache.RedisClient,
    mqClient *mq.RabbitMQClient,
) (*Data, error) {
    d := &Data{
        pgDB:        pgDB,
        mongoClient: mongoClient,
        redisClient: redisClient,
        mqClient:    mqClient,
    }
    
    if pgDB != nil {
        d.OrderRepo = NewOrderPgRepository(pgDB)
    } else {
        return nil, fmt.Errorf("no database configured")
    }
    
    return d, nil
}

func (d *Data) Close(ctx context.Context) error {
    if d.pgDB != nil {
        sqlDB, err := d.pgDB.DB()
        if err == nil {
            sqlDB.Close()
        }
    }
    if d.mongoClient != nil {
        d.mongoClient.Close(ctx)
    }
    if d.redisClient != nil {
        d.redisClient.Close()
    }
    if d.mqClient != nil {
        d.mqClient.Close()
    }
    return nil
}
```

---

### 步骤5：实现业务逻辑层

```go
// internal/order-service/biz/order_usecase.go
package biz

import (
    "context"
    
    "github.com/alfredchaos/demo/internal/order-service/data"
    "github.com/alfredchaos/demo/internal/order-service/domain"
    "github.com/alfredchaos/demo/pkg/log"
    "go.uber.org/zap"
)

// OrderUseCase 订单业务逻辑接口
type OrderUseCase interface {
    CreateOrder(ctx context.Context, userID string, items []domain.OrderItem) (*domain.Order, error)
    GetOrder(ctx context.Context, id string) (*domain.Order, error)
    ListOrders(ctx context.Context, userID string, offset, limit int) ([]*domain.Order, error)
}

// orderUseCase 订单业务逻辑实现
type orderUseCase struct {
    orderRepo data.OrderRepository
}

// NewOrderUseCase 创建订单业务逻辑
func NewOrderUseCase(orderRepo data.OrderRepository) OrderUseCase {
    return &orderUseCase{
        orderRepo: orderRepo,
    }
}

func (uc *orderUseCase) CreateOrder(ctx context.Context, userID string, items []domain.OrderItem) (*domain.Order, error) {
    log.WithContext(ctx).Info("creating order", zap.String("user_id", userID))
    
    // 创建订单
    order := domain.NewOrder(userID, items)
    
    // 验证订单
    if err := order.Validate(); err != nil {
        return nil, err
    }
    
    // 持久化
    if err := uc.orderRepo.Create(ctx, order); err != nil {
        log.WithContext(ctx).Error("failed to create order", zap.Error(err))
        return nil, err
    }
    
    log.WithContext(ctx).Info("order created", zap.String("order_id", order.ID))
    return order, nil
}

func (uc *orderUseCase) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
    return uc.orderRepo.GetByID(ctx, id)
}

func (uc *orderUseCase) ListOrders(ctx context.Context, userID string, offset, limit int) ([]*domain.Order, error) {
    return uc.orderRepo.ListByUserID(ctx, userID, offset, limit)
}
```

---

### 步骤6：实现服务层

```go
// internal/order-service/service/order_service.go
package service

import (
    "context"
    
    orderv1 "github.com/alfredchaos/demo/api/order/v1"
    "github.com/alfredchaos/demo/internal/order-service/biz"
    "github.com/alfredchaos/demo/internal/order-service/domain"
    "github.com/alfredchaos/demo/pkg/log"
    "go.uber.org/zap"
)

// OrderService gRPC服务实现
type OrderService struct {
    orderv1.UnimplementedOrderServiceServer
    useCase biz.OrderUseCase
}

// NewOrderService 创建订单服务
func NewOrderService(useCase biz.OrderUseCase) *OrderService {
    return &OrderService{
        useCase: useCase,
    }
}

func (s *OrderService) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
    log.WithContext(ctx).Info("received CreateOrder request")
    
    // 转换请求
    items := make([]domain.OrderItem, len(req.Items))
    for i, item := range req.Items {
        items[i] = domain.OrderItem{
            ProductID:   item.ProductId,
            ProductName: item.ProductName,
            Quantity:    item.Quantity,
            Price:       item.Price,
        }
    }
    
    // 调用业务逻辑
    order, err := s.useCase.CreateOrder(ctx, req.UserId, items)
    if err != nil {
        log.WithContext(ctx).Error("failed to create order", zap.Error(err))
        return nil, err
    }
    
    // 转换响应
    return &orderv1.CreateOrderResponse{
        Order: s.toProtoOrder(order),
    }, nil
}

func (s *OrderService) toProtoOrder(order *domain.Order) *orderv1.Order {
    items := make([]*orderv1.OrderItem, len(order.Items))
    for i, item := range order.Items {
        items[i] = &orderv1.OrderItem{
            ProductId:   item.ProductID,
            ProductName: item.ProductName,
            Quantity:    item.Quantity,
            Price:       item.Price,
        }
    }
    
    return &orderv1.Order{
        Id:          order.ID,
        UserId:      order.UserID,
        Items:       items,
        TotalAmount: order.TotalAmount,
        Status:      string(order.Status),
        CreatedAt:   order.CreatedAt.Unix(),
        UpdatedAt:   order.UpdatedAt.Unix(),
    }
}
```

---

### 步骤7：实现服务器层和配置

参考user-service的实现，创建 `server/grpc.go` 和 `conf/config.go`。

### 步骤8：创建主函数

```go
// cmd/order-service/main.go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/alfredchaos/demo/internal/order-service/biz"
    "github.com/alfredchaos/demo/internal/order-service/conf"
    "github.com/alfredchaos/demo/internal/order-service/data"
    "github.com/alfredchaos/demo/internal/order-service/server"
    "github.com/alfredchaos/demo/internal/order-service/service"
    "github.com/alfredchaos/demo/pkg/cache"
    "github.com/alfredchaos/demo/pkg/config"
    "github.com/alfredchaos/demo/pkg/db"
    "github.com/alfredchaos/demo/pkg/log"
    "github.com/alfredchaos/demo/pkg/mq"
    "go.uber.org/zap"
)

func main() {
    // 1. 加载配置
    var cfg conf.Config
    config.MustLoadConfig("order-service", &cfg)
    
    // 2. 初始化日志
    log.MustInitLogger(&cfg.Log, cfg.Server.Name)
    defer log.Sync()
    
    log.Info("starting order-service", zap.String("name", cfg.Server.Name))
    
    // 3. 初始化基础设施
    var pgDB *gorm.DB
    if cfg.Database.Enabled {
        pgDB = db.MustNewPostgresDB(&cfg.Database)
        defer func() {
            sqlDB, _ := pgDB.DB()
            if sqlDB != nil {
                sqlDB.Close()
            }
        }()
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
    go func() {
        if err := grpcServer.Start(); err != nil {
            log.Fatal("failed to start grpc server", zap.Error(err))
        }
    }()
    
    // 9. 优雅关闭
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Info("shutting down order-service")
    grpcServer.Stop()
    log.Info("order-service stopped")
}
```

---

### 步骤9：创建配置文件

```yaml
# configs/order-service.yaml
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

redis:
  enabled: false
  addr: localhost:6379
  password: ""
  db: 0
```

---

### 步骤10：数据库迁移

```sql
-- internal/order-service/migrations/001_create_orders_table.sql
-- +goose Up
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

-- +goose Down
DROP TABLE IF EXISTS orders;
```

**执行迁移**：

```bash
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="host=localhost port=5432 user=postgres password=password dbname=order_service sslmode=disable"
goose -dir internal/order-service/migrations up
```

---

## 测试

### 单元测试示例

```go
// internal/order-service/biz/order_usecase_test.go
package biz

import (
    "context"
    "testing"
    
    "github.com/alfredchaos/demo/internal/order-service/domain"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockOrderRepository Mock仓库
type MockOrderRepository struct {
    mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, order *domain.Order) error {
    args := m.Called(ctx, order)
    return args.Error(0)
}

func TestOrderUseCase_CreateOrder(t *testing.T) {
    mockRepo := new(MockOrderRepository)
    useCase := NewOrderUseCase(mockRepo)
    
    items := []domain.OrderItem{
        {ProductID: "p1", ProductName: "Product 1", Quantity: 2, Price: 10.0},
    }
    
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
    
    order, err := useCase.CreateOrder(context.Background(), "user1", items)
    
    assert.NoError(t, err)
    assert.NotNil(t, order)
    assert.Equal(t, "user1", order.UserID)
    assert.Equal(t, 20.0, order.TotalAmount)
    
    mockRepo.AssertExpectations(t)
}
```

---

## 运行服务

```bash
# 编译
go build -o bin/order-service ./cmd/order-service

# 运行
./bin/order-service

# 或者使用go run
go run cmd/order-service/main.go
```

---

## 注意事项

1. **遵循统一架构**：保持与其他服务相同的目录结构和分层
2. **使用依赖注入**：手动实现，保持代码透明
3. **错误处理**：使用领域错误，在service层转换为gRPC错误
4. **日志记录**：使用 `log.WithContext` 自动附加追踪信息
5. **配置管理**：使用YAML配置文件，支持环境变量覆盖
6. **数据库迁移**：使用Goose管理数据库版本
7. **测试覆盖**：编写单元测试和集成测试

---

---

## 步骤11：集成gRPC客户端（可选）

如果新服务需要调用其他内部服务，使用统一的 `pkg/grpcclient` 模块管理客户端连接。

### 1. 添加配置

在 `conf/config.go` 中添加gRPC客户端配置：

```go
// internal/order-service/conf/config.go
import "github.com/alfredchaos/demo/pkg/grpcclient"

type Config struct {
    Server      ServerConfig       `yaml:"server" mapstructure:"server"`
    Log         log.LogConfig      `yaml:"log" mapstructure:"log"`
    Database    db.Config          `yaml:"database" mapstructure:"database"`
    GRPCClients grpcclient.Config  `yaml:"grpc_clients" mapstructure:"grpc_clients"`  // 添加此行
}
```

### 2. 在配置文件中定义客户端

```yaml
# configs/order-service.yaml
server:
  name: order-service
  host: 0.0.0.0
  port: 9003

# ... 其他配置 ...

# gRPC客户端配置
grpc_clients:
  services:
    - name: user-service
      address: localhost:9001
      timeout: 5s
      retry:
        max: 3
        timeout: 10s
        backoff: 100ms
    - name: product-service
      address: localhost:9004
      timeout: 5s
```

### 3. 注册客户端工厂

在 `main.go` 的 `init()` 函数中注册：

```go
// cmd/order-service/main.go
import (
    userv1 "github.com/alfredchaos/demo/api/user/v1"
    productv1 "github.com/alfredchaos/demo/api/product/v1"
    "github.com/alfredchaos/demo/pkg/grpcclient"
    "google.golang.org/grpc"
)

func init() {
    // 注册gRPC客户端工厂
    grpcclient.GlobalRegistry.Register("user-service", func(conn *grpc.ClientConn) interface{} {
        return userv1.NewUserServiceClient(conn)
    })
    
    grpcclient.GlobalRegistry.Register("product-service", func(conn *grpc.ClientConn) interface{} {
        return productv1.NewProductServiceClient(conn)
    })
}
```

### 4. 初始化客户端管理器

在 `main()` 函数中初始化：

```go
func main() {
    // 加载配置
    var cfg conf.Config
    config.MustLoadConfig("order-service", &cfg)
    
    // 初始化日志...
    
    // 初始化gRPC客户端管理器
    var clientManager *grpcclient.Manager
    if len(cfg.GRPCClients.Services) > 0 {
        clientManager = grpcclient.NewManager()
        
        // 注册服务配置
        for _, svc := range cfg.GRPCClients.Services {
            if err := clientManager.Register(&svc); err != nil {
                log.Fatal("failed to register service", 
                    zap.String("service", svc.Name), 
                    zap.Error(err))
            }
        }
        
        // 连接所有服务
        if err := clientManager.ConnectAll(); err != nil {
            log.Fatal("failed to connect services", zap.Error(err))
        }
        defer clientManager.Close()
    }
    
    // 初始化数据层...
    dataLayer, _ := data.NewData(pgDB, mongoClient, redisClient, mqClient)
    
    // 获取gRPC客户端
    var userClient userv1.UserServiceClient
    if clientManager != nil {
        userConn, err := clientManager.GetConnection("user-service")
        if err == nil {
            userClient = userv1.NewUserServiceClient(userConn)
        }
    }
    
    // 初始化业务层（注入gRPC客户端）
    orderUseCase := biz.NewOrderUseCase(dataLayer.OrderRepo, userClient)
    
    // 后续初始化...
}
```

### 5. 在业务层中使用

```go
// internal/order-service/biz/order_usecase.go
type orderUseCase struct {
    orderRepo  data.OrderRepository
    userClient userv1.UserServiceClient  // 注入客户端
}

func NewOrderUseCase(orderRepo data.OrderRepository, userClient userv1.UserServiceClient) OrderUseCase {
    return &orderUseCase{
        orderRepo:  orderRepo,
        userClient: userClient,
    }
}

func (uc *orderUseCase) CreateOrder(ctx context.Context, userID string, items []domain.OrderItem) (*domain.Order, error) {
    // 验证用户是否存在
    if uc.userClient != nil {
        _, err := uc.userClient.GetUser(ctx, &userv1.GetUserRequest{Id: userID})
        if err != nil {
            return nil, fmt.Errorf("invalid user: %w", err)
        }
    }
    
    // 创建订单
    order := domain.NewOrder(userID, items)
    if err := uc.orderRepo.Create(ctx, order); err != nil {
        return nil, err
    }
    
    return order, nil
}
```

---

## 总结

通过遵循本指南，你可以快速创建一个结构清晰、易于维护的内部服务。关键点：

- 统一的分层架构
- 清晰的依赖注入
- 完善的错误处理
- 良好的测试覆盖
- **使用 `pkg/grpcclient` 统一管理服务间调用**

参考现有的 `user-service` 和 `book-service` 可以获得更多实践经验。
