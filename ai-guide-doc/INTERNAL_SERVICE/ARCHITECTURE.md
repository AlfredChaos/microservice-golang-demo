# 内部服务分层架构设计

## 架构概览

内部服务采用清晰的分层架构，每一层都有明确的职责，通过接口进行解耦。架构自下而上分为以下几层：

```
┌─────────────────────────────────────────────────────────────┐
│                        gRPC Client                          │
│                     (api-gateway调用)                        │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      Server Layer                           │
│                    (gRPC服务器配置)                          │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                     Service Layer                           │
│              (gRPC接口实现，协议转换)                         │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      Biz Layer                              │
│                  (业务逻辑，用例实现)                         │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                     Domain Layer                            │
│                (领域模型，业务规则)                           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      Data Layer                             │
│            (数据访问，仓库模式实现)                           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│          Infrastructure (PostgreSQL/MongoDB/Redis)          │
│                     + RabbitMQ                              │
└─────────────────────────────────────────────────────────────┘
```

## 各层详细说明

### 1. Domain Layer（领域层）

**职责**：
- 定义核心业务实体和值对象
- 包含业务规则和验证逻辑
- 不依赖任何外部框架或基础设施

**文件组织**：
```
domain/
├── user.go       # 实体定义
├── errors.go     # 领域错误
└── interfaces.go # 领域接口（可选）
```

**示例**：
```go
package domain

import "time"

// User 用户领域模型
type User struct {
    ID        string
    Username  string
    Email     string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// NewUser 工厂函数，确保对象有效性
func NewUser(username, email string) *User {
    now := time.Now()
    return &User{
        Username:  username,
        Email:     email,
        CreatedAt: now,
        UpdatedAt: now,
    }
}

// Validate 业务规则验证
func (u *User) Validate() error {
    if u.Username == "" {
        return ErrInvalidUsername
    }
    if u.Email == "" {
        return ErrInvalidEmail
    }
    return nil
}
```

**关键点**：
- 领域模型应该是"贫血模型"还是"充血模型"取决于业务复杂度
- 简单CRUD可以使用贫血模型
- 复杂业务逻辑应该封装在领域对象中

---

### 2. Data Layer（数据访问层）

**职责**：
- 定义数据仓库接口（Repository Interface）
- 实现具体的数据访问逻辑
- 封装数据库、缓存、消息队列的操作细节

**文件组织**：
```
data/
├── data.go           # 数据层容器，管理所有连接
├── user_repo.go      # 仓库接口定义
├── user_pg_repo.go   # PostgreSQL实现
├── user_mongo_repo.go # MongoDB实现（可选）
└── user_cache.go     # Redis缓存（可选）
```

**仓库接口定义**：
```go
package data

import (
    "context"
    "github.com/alfredchaos/demo/internal/user-service/domain"
)

// UserRepository 用户仓库接口
// 定义数据访问的抽象，业务层依赖此接口
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id string) (*domain.User, error)
    GetByUsername(ctx context.Context, username string) (*domain.User, error)
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, offset, limit int) ([]*domain.User, error)
}
```

**PostgreSQL实现**：
```go
package data

import (
    "context"
    "github.com/alfredchaos/demo/internal/user-service/domain"
    "gorm.io/gorm"
)

// userPgRepository PostgreSQL仓库实现
type userPgRepository struct {
    db *gorm.DB
}

// NewUserPgRepository 创建PostgreSQL仓库
func NewUserPgRepository(db *gorm.DB) UserRepository {
    return &userPgRepository{db: db}
}

func (r *userPgRepository) Create(ctx context.Context, user *domain.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

func (r *userPgRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
    var user domain.User
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, domain.ErrUserNotFound
        }
        return nil, err
    }
    return &user, nil
}
```

**数据层容器**：
```go
package data

import (
    "context"
    "gorm.io/gorm"
    "github.com/alfredchaos/demo/pkg/cache"
    "github.com/alfredchaos/demo/pkg/db"
    "github.com/alfredchaos/demo/pkg/mq"
)

// Data 数据访问层容器
type Data struct {
    // 数据库和缓存客户端
    pgDB        *gorm.DB
    mongoClient *db.MongoClient
    redisClient *cache.RedisClient
    mqClient    *mq.RabbitMQClient
    
    // 仓库实例
    UserRepo UserRepository
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
    
    // 根据配置选择仓库实现
    if pgDB != nil {
        d.UserRepo = NewUserPgRepository(pgDB)
    } else if mongoClient != nil {
        d.UserRepo = NewUserMongoRepository(mongoClient)
    }
    
    return d, nil
}

func (d *Data) Close(ctx context.Context) error {
    // 关闭所有连接
    return nil
}
```

---

### 3. Biz Layer（业务逻辑层）

**职责**：
- 实现业务用例（Use Case）
- 编排数据访问和业务规则
- 不包含任何HTTP、gRPC等协议相关代码

**文件组织**：
```
biz/
├── user_usecase.go    # 用户业务逻辑
└── book_usecase.go    # 书籍业务逻辑
```

**用例接口定义**：
```go
package biz

import (
    "context"
    "github.com/alfredchaos/demo/internal/user-service/domain"
)

// UserUseCase 用户业务逻辑接口
type UserUseCase interface {
    CreateUser(ctx context.Context, username, email string) (*domain.User, error)
    GetUser(ctx context.Context, id string) (*domain.User, error)
    UpdateUser(ctx context.Context, user *domain.User) error
    DeleteUser(ctx context.Context, id string) error
    ListUsers(ctx context.Context, offset, limit int) ([]*domain.User, error)
}
```

**用例实现**：
```go
package biz

import (
    "context"
    "github.com/alfredchaos/demo/internal/user-service/data"
    "github.com/alfredchaos/demo/internal/user-service/domain"
    "github.com/alfredchaos/demo/pkg/log"
    "go.uber.org/zap"
)

// userUseCase 用户业务逻辑实现
type userUseCase struct {
    userRepo data.UserRepository
    // 可以注入其他仓库或服务
}

// NewUserUseCase 创建用户业务逻辑
func NewUserUseCase(userRepo data.UserRepository) UserUseCase {
    return &userUseCase{
        userRepo: userRepo,
    }
}

func (uc *userUseCase) CreateUser(ctx context.Context, username, email string) (*domain.User, error) {
    log.WithContext(ctx).Info("creating user", zap.String("username", username))
    
    // 1. 创建领域对象
    user := domain.NewUser(username, email)
    
    // 2. 验证业务规则
    if err := user.Validate(); err != nil {
        return nil, err
    }
    
    // 3. 检查唯一性
    existingUser, err := uc.userRepo.GetByUsername(ctx, username)
    if err != nil && err != domain.ErrUserNotFound {
        return nil, err
    }
    if existingUser != nil {
        return nil, domain.ErrUserAlreadyExists
    }
    
    // 4. 持久化
    if err := uc.userRepo.Create(ctx, user); err != nil {
        log.WithContext(ctx).Error("failed to create user", zap.Error(err))
        return nil, err
    }
    
    return user, nil
}
```

**复杂业务示例（编排多个服务）**：
```go
// ProcessHelloRequest 处理Hello请求，调用多个服务
func (uc *helloUseCase) ProcessHelloRequest(ctx context.Context, name string) (string, error) {
    // 1. 调用用户服务获取用户信息
    userMsg, err := uc.userClient.CallUserService(ctx, name)
    if err != nil {
        return "", err
    }
    
    // 2. 调用书籍服务获取推荐
    bookMsg, err := uc.bookClient.CallBookService(ctx)
    if err != nil {
        return "", err
    }
    
    // 3. 组合结果
    response := userMsg + " " + bookMsg
    
    // 4. 发送消息到队列
    go func() {
        msgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        defer cancel()
        
        message := map[string]string{
            "type":    "hello",
            "message": response,
        }
        
        msgBytes, _ := json.Marshal(message)
        uc.publisher.Publish(msgCtx, msgBytes)
    }()
    
    return response, nil
}
```

---

### 4. Service Layer（服务层）

**职责**：
- 实现gRPC接口
- 处理协议转换（Protobuf ↔ Domain Model）
- 调用业务逻辑层

**文件组织**：
```
service/
└── user_service.go  # gRPC服务实现
```

**服务实现**：
```go
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

// CreateUser 实现创建用户接口
func (s *UserService) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
    log.WithContext(ctx).Info("received CreateUser request", zap.String("username", req.Username))
    
    // 调用业务逻辑
    user, err := s.useCase.CreateUser(ctx, req.Username, req.Email)
    if err != nil {
        log.WithContext(ctx).Error("failed to create user", zap.Error(err))
        return nil, err
    }
    
    // 转换为gRPC响应
    return &userv1.CreateUserResponse{
        User: &userv1.User{
            Id:       user.ID,
            Username: user.Username,
            Email:    user.Email,
        },
    }, nil
}
```

**关键点**：
- Service层应该很薄，只做协议转换
- 不应该包含业务逻辑
- 错误处理和日志记录

---

### 5. Server Layer（服务器层）

**职责**：
- 配置和启动gRPC服务器
- 注册gRPC服务
- 配置中间件（日志、追踪、恢复等）

**文件组织**：
```
server/
├── grpc.go     # gRPC服务器
└── client.go   # gRPC客户端（用于调用其他服务）
```

**gRPC服务器实现**：
```go
package server

import (
    "fmt"
    "net"
    
    userv1 "github.com/alfredchaos/demo/api/user/v1"
    "github.com/alfredchaos/demo/internal/user-service/conf"
    "github.com/alfredchaos/demo/internal/user-service/service"
    "github.com/alfredchaos/demo/pkg/log"
    "google.golang.org/grpc"
    "go.uber.org/zap"
)

// GRPCServer gRPC服务器封装
type GRPCServer struct {
    server      *grpc.Server
    config      *conf.ServerConfig
    userService *service.UserService
}

// NewGRPCServer 创建gRPC服务器
func NewGRPCServer(cfg *conf.ServerConfig, userService *service.UserService) *GRPCServer {
    // 创建gRPC服务器，可添加中间件
    server := grpc.NewServer(
        // grpc.UnaryInterceptor(middleware.LoggingInterceptor),
    )
    
    // 注册服务
    userv1.RegisterUserServiceServer(server, userService)
    
    return &GRPCServer{
        server:      server,
        config:      cfg,
        userService: userService,
    }
}

// Start 启动gRPC服务器
func (s *GRPCServer) Start() error {
    addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
    lis, err := net.Listen("tcp", addr)
    if err != nil {
        return err
    }
    
    log.Info("gRPC server listening", zap.String("addr", addr))
    return s.server.Serve(lis)
}

// Stop 停止gRPC服务器
func (s *GRPCServer) Stop() {
    s.server.GracefulStop()
}
```

---

### 6. Conf Layer（配置层）

**职责**：
- 定义配置结构
- 加载和验证配置

**文件组织**：
```
conf/
└── config.go  # 配置定义
```

**配置定义**：
```go
package conf

import (
    "github.com/alfredchaos/demo/pkg/cache"
    "github.com/alfredchaos/demo/pkg/db"
    "github.com/alfredchaos/demo/pkg/log"
    "github.com/alfredchaos/demo/pkg/mq"
)

// Config 服务配置
type Config struct {
    Server   ServerConfig      `yaml:"server" mapstructure:"server"`
    Log      log.LogConfig     `yaml:"log" mapstructure:"log"`
    Database db.DatabaseConfig `yaml:"database" mapstructure:"database"`
    MongoDB  db.MongoConfig    `yaml:"mongodb" mapstructure:"mongodb"`
    Redis    cache.RedisConfig `yaml:"redis" mapstructure:"redis"`
    RabbitMQ mq.RabbitMQConfig `yaml:"rabbitmq" mapstructure:"rabbitmq"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
    Name string `yaml:"name" mapstructure:"name"`
    Host string `yaml:"host" mapstructure:"host"`
    Port int    `yaml:"port" mapstructure:"port"`
}
```

---

## 依赖关系图

```
main.go
  └─> Server (grpc.go)
       └─> Service (user_service.go)
            └─> Biz (user_usecase.go)
                 ├─> Domain (user.go)
                 └─> Data (user_repo.go)
                      └─> Infrastructure (PostgreSQL/MongoDB/Redis)
```

**依赖原则**：
- 外层依赖内层
- 内层不知道外层的存在
- 通过接口解耦

---

## 目录结构完整示例

```
internal/user-service/
├── domain/                   # 领域层
│   ├── user.go              # 用户实体
│   ├── errors.go            # 领域错误
│   └── interfaces.go        # 领域接口（可选）
│
├── data/                    # 数据访问层
│   ├── data.go              # 数据层容器
│   ├── user_repo.go         # 仓库接口
│   ├── user_pg_repo.go      # PostgreSQL实现
│   ├── user_mongo_repo.go   # MongoDB实现
│   └── user_cache.go        # Redis缓存
│
├── biz/                     # 业务逻辑层
│   ├── user_usecase.go      # 用户用例
│   └── hello_usecase.go     # Hello用例
│
├── service/                 # 服务层（gRPC）
│   └── user_service.go      # gRPC服务实现
│
├── server/                  # 服务器层
│   ├── grpc.go              # gRPC服务器
│   └── client.go            # gRPC客户端（可选）
│
├── conf/                    # 配置层
│   └── config.go            # 配置定义
│
└── migrations/              # 数据库迁移（可选）
    └── 001_create_users_table.sql

cmd/user-service/
└── main.go                  # 启动入口

configs/
└── user-service.yaml        # 配置文件

api/user/v1/                 # gRPC接口定义（共享）
├── user.proto               # Protobuf定义
└── user.pb.go               # 生成的代码
```

---

## 关键设计决策

### 1. 为什么不使用Wire？

虽然Wire是Google官方的依赖注入工具，但手动实现依赖注入有以下优势：
- **透明性**：代码更直观，不需要理解Wire的魔法
- **调试友好**：出错时容易定位问题
- **学习成本低**：新人可以快速理解
- **灵活性**：可以根据需要自定义初始化逻辑

参考 [DI_AND_WIRE.md](./DI_AND_WIRE.md) 了解具体实现。

### 2. 为什么分这么多层？

- **单一职责**：每一层只关注一件事
- **易于测试**：可以单独测试每一层
- **易于替换**：比如从MongoDB切换到PostgreSQL只需要修改Data层
- **团队协作**：不同的人可以并行开发不同的层

### 3. 仓库模式 vs DAO模式

我们使用仓库模式而非DAO模式：
- **仓库**：面向领域对象，提供集合语义
- **DAO**：面向数据库表，提供CRUD操作

仓库模式更符合DDD思想，抽象层次更高。

---

## 下一步

- 阅读 [DI_AND_WIRE.md](./DI_AND_WIRE.md) 了解依赖注入实现
- 阅读 [DATA_STORAGE.md](./DATA_STORAGE.md) 了解数据存储层设计
- 阅读 [SERVICE_COMMUNICATION.md](./SERVICE_COMMUNICATION.md) 了解服务间通信
