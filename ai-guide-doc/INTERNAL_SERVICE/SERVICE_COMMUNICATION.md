# 服务间通信设计

## 概述

内部服务支持两种通信模式：
1. **南北向通信**：作为gRPC服务端，接收来自api-gateway的请求
2. **东西向通信**：作为gRPC客户端，调用其他内部服务

所有通信都基于gRPC协议，使用Protobuf进行序列化。

---

## gRPC接口管理

### 1. 统一的Proto文件管理

所有gRPC接口定义统一放在项目根目录的 `api/` 目录下，按服务和版本组织。

**目录结构**：

```
api/
├── user/
│   └── v1/
│       ├── user.proto          # Proto定义
│       ├── user.pb.go          # 生成的Go代码
│       └── user_grpc.pb.go     # 生成的gRPC代码
├── book/
│   └── v1/
│       ├── book.proto
│       ├── book.pb.go
│       └── book_grpc.pb.go
└── order/
    └── v1/
        └── order.proto
```

**优势**：
- 所有服务共享相同的接口定义
- 避免接口变更时所有服务都需要更新
- 方便版本管理和向后兼容

### 2. Proto文件示例

```protobuf
// api/user/v1/user.proto
syntax = "proto3";

package user.v1;

option go_package = "github.com/alfredchaos/demo/api/user/v1;userv1";

// UserService 用户服务接口
service UserService {
  // SayHello 问候接口
  rpc SayHello(HelloRequest) returns (HelloResponse);
  
  // CreateUser 创建用户
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  
  // GetUser 获取用户
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}

// HelloRequest 问候请求
message HelloRequest {
  string name = 1;
}

// HelloResponse 问候响应
message HelloResponse {
  string message = 1;
}

// CreateUserRequest 创建用户请求
message CreateUserRequest {
  string username = 1;
  string email = 2;
}

// CreateUserResponse 创建用户响应
message CreateUserResponse {
  User user = 1;
}

// GetUserRequest 获取用户请求
message GetUserRequest {
  string id = 1;
}

// GetUserResponse 获取用户响应
message GetUserResponse {
  User user = 1;
}

// User 用户信息
message User {
  string id = 1;
  string username = 2;
  string email = 3;
  int64 created_at = 4;
  int64 updated_at = 5;
}
```

### 3. 生成Go代码

```bash
# scripts/gen-proto.sh
#!/bin/bash

set -e

# 安装依赖（首次执行）
# go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
# go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 生成所有proto文件
find api -name "*.proto" | while read proto_file; do
    echo "Generating $proto_file"
    protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           "$proto_file"
done

echo "Proto generation completed"
```

---

## 南北向通信：作为gRPC服务端

### 1. 服务实现

```go
// internal/user-service/service/user_service.go
package service

import (
    "context"
    
    userv1 "github.com/alfredchaos/demo/api/user/v1"
    "github.com/alfredchaos/demo/internal/user-service/biz"
    "github.com/alfredchaos/demo/pkg/log"
    "go.uber.org/zap"
)

// UserService gRPC服务实现
type UserService struct {
    userv1.UnimplementedUserServiceServer
    useCase biz.UserUseCase
}

// NewUserService 创建用户服务
func NewUserService(useCase biz.UserUseCase) *UserService {
    return &UserService{
        useCase: useCase,
    }
}

// SayHello 实现SayHello接口
func (s *UserService) SayHello(ctx context.Context, req *userv1.HelloRequest) (*userv1.HelloResponse, error) {
    log.WithContext(ctx).Info("received SayHello request", zap.String("name", req.Name))
    
    message, err := s.useCase.SayHello(ctx)
    if err != nil {
        log.WithContext(ctx).Error("failed to say hello", zap.Error(err))
        return nil, err
    }
    
    return &userv1.HelloResponse{
        Message: message,
    }, nil
}

// CreateUser 实现CreateUser接口
func (s *UserService) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
    log.WithContext(ctx).Info("received CreateUser request", zap.String("username", req.Username))
    
    user, err := s.useCase.CreateUser(ctx, req.Username, req.Email)
    if err != nil {
        log.WithContext(ctx).Error("failed to create user", zap.Error(err))
        return nil, err
    }
    
    return &userv1.CreateUserResponse{
        User: &userv1.User{
            Id:       user.ID,
            Username: user.Username,
            Email:    user.Email,
        },
    }, nil
}
```

### 2. gRPC服务器配置

```go
// internal/user-service/server/grpc.go
package server

import (
    "fmt"
    "net"
    
    userv1 "github.com/alfredchaos/demo/api/user/v1"
    "github.com/alfredchaos/demo/internal/user-service/conf"
    "github.com/alfredchaos/demo/internal/user-service/service"
    "github.com/alfredchaos/demo/pkg/log"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
    "go.uber.org/zap"
)

// GRPCServer gRPC服务器
type GRPCServer struct {
    server      *grpc.Server
    config      *conf.ServerConfig
    userService *service.UserService
}

// NewGRPCServer 创建gRPC服务器
func NewGRPCServer(cfg *conf.ServerConfig, userService *service.UserService) *GRPCServer {
    // 创建gRPC服务器，可添加拦截器
    server := grpc.NewServer(
        grpc.ChainUnaryInterceptor(
            // 可以添加日志、认证、限流等拦截器
        ),
    )
    
    // 注册服务
    userv1.RegisterUserServiceServer(server, userService)
    
    // 注册反射服务（用于grpcurl等工具）
    reflection.Register(server)
    
    return &GRPCServer{
        server:      server,
        config:      cfg,
        userService: userService,
    }
}

// Start 启动服务器
func (s *GRPCServer) Start() error {
    addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
    lis, err := net.Listen("tcp", addr)
    if err != nil {
        return fmt.Errorf("failed to listen: %w", err)
    }
    
    log.Info("gRPC server listening", zap.String("addr", addr))
    return s.server.Serve(lis)
}

// Stop 停止服务器
func (s *GRPCServer) Stop() {
    log.Info("stopping gRPC server")
    s.server.GracefulStop()
}
```

---

## 东西向通信：作为gRPC客户端

> **重要更新**：现在使用统一的 `pkg/grpcclient` 公共模块管理gRPC客户端连接。

### 1. 使用公共gRPC客户端管理模块

内部服务调用其他服务时，使用 `pkg/grpcclient` 模块统一管理连接，享受以下优势：
- ✅ **统一管理**：与api-gateway使用相同的连接管理逻辑
- ✅ **连接复用**：避免重复创建连接，提高性能
- ✅ **配置驱动**：通过YAML配置文件管理服务连接
- ✅ **拦截器支持**：统一的日志、追踪、重试功能
- ✅ **生命周期管理**：自动处理连接建立和关闭

### 2. 配置文件

在服务配置文件中添加 `grpc_clients` 配置段：

```yaml
# configs/user-service.yaml
server:
  name: user-service
  host: 0.0.0.0
  port: 9001

log:
  level: debug

# 数据库配置...

# gRPC客户端配置（仅当需要调用其他服务时）
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

### 3. 注册客户端工厂

在 `main.go` 的 `init()` 函数中注册需要的客户端工厂：

```go
// cmd/user-service/main.go
package main

import (
    bookv1 "github.com/alfredchaos/demo/api/book/v1"
    orderv1 "github.com/alfredchaos/demo/api/order/v1"
    "github.com/alfredchaos/demo/pkg/grpcclient"
    "google.golang.org/grpc"
)

func init() {
    // 注册gRPC客户端工厂
    grpcclient.GlobalRegistry.Register("book-service", func(conn *grpc.ClientConn) interface{} {
        return bookv1.NewBookServiceClient(conn)
    })
    
    grpcclient.GlobalRegistry.Register("order-service", func(conn *grpc.ClientConn) interface{} {
        return orderv1.NewOrderServiceClient(conn)
    })
}
```

### 4. 初始化客户端管理器

在 `main()` 函数中初始化gRPC客户端管理器：

```go
func main() {
    // 加载配置
    var cfg conf.Config
    config.MustLoadConfig("user-service", &cfg)
    
    // 初始化日志...
    
    // 初始化gRPC客户端管理器（如果需要调用其他服务）
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
    var bookClient bookv1.BookServiceClient
    if clientManager != nil {
        bookConn, _ := clientManager.GetConnection("book-service")
        bookClient = bookv1.NewBookServiceClient(bookConn)
    }
    
    // 初始化业务层（注入gRPC客户端）
    userUseCase := biz.NewUserUseCase(dataLayer.UserRepo, bookClient)
    
    // 后续初始化...
}
```

### 5. 在业务逻辑中使用

```go
// internal/user-service/biz/hello_usecase.go
package biz

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    bookv1 "github.com/alfredchaos/demo/api/book/v1"
    "github.com/alfredchaos/demo/pkg/log"
    "github.com/alfredchaos/demo/pkg/mq"
    "go.uber.org/zap"
)

// HelloUseCase Hello业务逻辑接口
type HelloUseCase interface {
    ProcessHello(ctx context.Context, name string) (string, error)
}

// helloUseCase Hello业务逻辑实现
type helloUseCase struct {
    bookClient bookv1.BookServiceClient  // 直接使用生成的gRPC客户端
    publisher  mq.Publisher
}

// NewHelloUseCase 创建Hello业务逻辑
func NewHelloUseCase(bookClient bookv1.BookServiceClient, publisher mq.Publisher) HelloUseCase {
    return &helloUseCase{
        bookClient: bookClient,
        publisher:  publisher,
    }
}

func (uc *helloUseCase) ProcessHello(ctx context.Context, name string) (string, error) {
    log.WithContext(ctx).Info("processing hello request", zap.String("name", name))
    
    // 调用其他服务（如果客户端可用）
    var recommendation string
    if uc.bookClient != nil {
        resp, err := uc.bookClient.GetRecommendation(ctx, &bookv1.GetRecommendationRequest{
            UserId: name,
        })
        if err != nil {
            log.WithContext(ctx).Warn("failed to get recommendation", zap.Error(err))
            recommendation = "No recommendation"
        } else {
            recommendation = resp.Recommendation
        }
    } else {
        recommendation = "Book service not available"
    }
    
    response := fmt.Sprintf("Hello %s! %s", name, recommendation)
    
    // 发送消息到队列
    go func() {
        msgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        defer cancel()
        
        message := map[string]string{
            "type":    "hello",
            "name":    name,
            "message": response,
        }
        
        msgBytes, _ := json.Marshal(message)
        if err := uc.publisher.Publish(msgCtx, msgBytes); err != nil {
            log.Error("failed to publish message", zap.Error(err))
        }
    }()
    
    return response, nil
}
```

---

## RabbitMQ集成

### 1. 配置

```yaml
# configs/user-service.yaml
rabbitmq:
  enabled: true
  url: amqp://guest:guest@localhost:5672/
  exchange: demo_exchange
  exchange_type: topic
  queue: user_service_queue
  routing_key: user.#
  durable: true
  auto_delete: false
```

### 2. 发布者实现

```go
// pkg/mq/publisher.go
package mq

import (
    "context"
    
    amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher 消息发布者接口
type Publisher interface {
    Publish(ctx context.Context, body []byte) error
}

// rabbitmqPublisher RabbitMQ发布者实现
type rabbitmqPublisher struct {
    client *RabbitMQClient
}

// NewPublisher 创建发布者
func NewPublisher(client *RabbitMQClient) Publisher {
    return &rabbitmqPublisher{client: client}
}

func (p *rabbitmqPublisher) Publish(ctx context.Context, body []byte) error {
    return p.client.GetChannel().PublishWithContext(
        ctx,
        p.client.config.Exchange,
        p.client.config.RoutingKey,
        false,
        false,
        amqp.Publishing{
            ContentType: "application/json",
            Body:        body,
        },
    )
}
```

### 3. 消费者实现

```go
// internal/user-service/consumer/message_consumer.go
package consumer

import (
    "context"
    "encoding/json"
    
    "github.com/alfredchaos/demo/pkg/log"
    "github.com/alfredchaos/demo/pkg/mq"
    amqp "github.com/rabbitmq/amqp091-go"
    "go.uber.org/zap"
)

// MessageConsumer 消息消费者
type MessageConsumer struct {
    mqClient *mq.RabbitMQClient
}

// NewMessageConsumer 创建消息消费者
func NewMessageConsumer(mqClient *mq.RabbitMQClient) *MessageConsumer {
    return &MessageConsumer{
        mqClient: mqClient,
    }
}

// Start 开始消费消息
func (c *MessageConsumer) Start(ctx context.Context) error {
    msgs, err := c.mqClient.GetChannel().Consume(
        c.mqClient.config.Queue,
        "",
        true,
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        return err
    }
    
    log.Info("message consumer started")
    
    go func() {
        for {
            select {
            case msg := <-msgs:
                c.handleMessage(ctx, msg)
            case <-ctx.Done():
                log.Info("message consumer stopped")
                return
            }
        }
    }()
    
    return nil
}

func (c *MessageConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
    log.WithContext(ctx).Info("received message", zap.String("body", string(msg.Body)))
    
    var message map[string]interface{}
    if err := json.Unmarshal(msg.Body, &message); err != nil {
        log.Error("failed to unmarshal message", zap.Error(err))
        return
    }
    
    // 处理消息
    messageType, _ := message["type"].(string)
    switch messageType {
    case "hello":
        c.handleHelloMessage(ctx, message)
    default:
        log.Warn("unknown message type", zap.String("type", messageType))
    }
}

func (c *MessageConsumer) handleHelloMessage(ctx context.Context, message map[string]interface{}) {
    log.WithContext(ctx).Info("handling hello message", zap.Any("message", message))
    // 业务逻辑处理
}
```

---

## 最佳实践

### 1. 错误处理

使用gRPC的标准错误码：

```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func (s *UserService) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
    user, err := s.useCase.GetUser(ctx, req.Id)
    if err != nil {
        if errors.Is(err, domain.ErrUserNotFound) {
            return nil, status.Error(codes.NotFound, "user not found")
        }
        return nil, status.Error(codes.Internal, "internal error")
    }
    return &userv1.GetUserResponse{User: user}, nil
}
```

### 2. 超时控制

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resp, err := client.GetUser(ctx, req)
```

### 3. 重试机制

使用 `pkg/grpcclient` 模块时，重试机制通过配置文件设置：

```yaml
grpc_clients:
  services:
    - name: book-service
      address: localhost:9002
      timeout: 5s
      retry:  # 重试配置
        max: 3                # 最大重试次数
        timeout: 10s          # 重试总超时
        backoff: 100ms        # 退避时间
```

拦截器会自动处理重试逻辑，无需手动编码。

### 4. 服务发现（可选）

使用Consul或etcd进行服务发现，避免硬编码服务地址。

---

## 迁移指南

### 从旧版客户端管理迁移到 pkg/grpcclient

如果你的服务正在使用旧的客户端管理方式，可以按以下步骤迁移：

#### 步骤1：删除旧的客户端管理代码

删除 `internal/xxx-service/server/client.go` 和 `internal/xxx-service/client/` 中的自定义客户端封装。

#### 步骤2：更新配置结构

```go
// 在 conf/config.go 中添加
import "github.com/alfredchaos/demo/pkg/grpcclient"

type Config struct {
    // ... 其他配置 ...
    GRPCClients grpcclient.Config `yaml:"grpc_clients" mapstructure:"grpc_clients"`
}
```

#### 步骤3：更新配置文件

```yaml
# configs/xxx-service.yaml
# 添加以下配置
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

#### 步骤4：注册客户端工厂

在 `main.go` 添加 `init()` 函数：

```go
func init() {
    grpcclient.GlobalRegistry.Register("book-service", func(conn *grpc.ClientConn) interface{} {
        return bookv1.NewBookServiceClient(conn)
    })
}
```

#### 步骤5：更新main.go中的初始化逻辑

```go
// 旧代码（删除）
// bookConn, err := grpc.Dial("localhost:9002", ...)
// bookClient := bookv1.NewBookServiceClient(bookConn)

// 新代码
var clientManager *grpcclient.Manager
if len(cfg.GRPCClients.Services) > 0 {
    clientManager = grpcclient.NewManager()
    for _, svc := range cfg.GRPCClients.Services {
        clientManager.Register(&svc)
    }
    clientManager.ConnectAll()
    defer clientManager.Close()
}

var bookClient bookv1.BookServiceClient
if clientManager != nil {
    bookConn, _ := clientManager.GetConnection("book-service")
    bookClient = bookv1.NewBookServiceClient(bookConn)
}
```

#### 步骤6：更新业务层

```go
// 旧代码（如果有自定义接口封装）
// type BookClient interface { ... }

// 新代码：直接使用生成的客户端类型
type userUseCase struct {
    bookClient bookv1.BookServiceClient  // 直接使用
}
```

#### 迁移优势

- ✅ 减少代码量（删除自定义客户端封装）
- ✅ 统一管理（与api-gateway相同的管理方式）
- ✅ 配置驱动（更灵活的配置）
- ✅ 开箱即用（日志、追踪、重试等功能）

---

## 总结

通过统一的gRPC接口管理和 `pkg/grpcclient` 模块，内部服务可以：

1. **作为服务端**响应来自api-gateway的请求
2. **作为客户端**调用其他内部服务（使用统一的连接管理）
3. 通过RabbitMQ实现**异步通信和事件驱动**
4. 保持接口定义的**一致性和可维护性**
5. **享受统一的日志、追踪、重试等功能**
