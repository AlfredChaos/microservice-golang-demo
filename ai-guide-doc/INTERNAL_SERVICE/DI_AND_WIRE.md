# 依赖注入设计与实现

## 概述

本项目采用手动依赖注入方式，不使用Wire等自动化工具。虽然Wire可以减少样板代码，但手动注入更透明、易于理解和调试。

## 为什么手动实现依赖注入？

### 优势

1. **代码透明性**
   - 所有依赖关系一目了然
   - 不需要理解Wire的编译时代码生成

2. **调试友好**
   - 出错时可以直接看到初始化逻辑
   - 不需要查看生成的代码

3. **学习成本低**
   - 新人可以快速理解依赖关系
   - 不需要学习Wire的语法和规则

4. **灵活性高**
   - 可以在初始化时执行复杂逻辑
   - 可以根据配置动态选择实现

### 劣势

- 需要手动维护依赖链
- 代码量稍多

对于微服务项目，依赖关系相对简单，手动注入的成本可以接受。

---

## 依赖注入原则

### 1. 依赖倒置原则（DIP）

> 高层模块不应该依赖低层模块，两者都应该依赖抽象

**示例**：

```go
// ❌ 错误：直接依赖具体实现
type UserUseCase struct {
    repo *UserPgRepository  // 依赖具体的PostgreSQL实现
}

// ✅ 正确：依赖接口
type UserUseCase struct {
    repo UserRepository  // 依赖抽象接口
}
```

### 2. 接口隔离原则（ISP）

> 客户端不应该依赖它不需要的接口

**示例**：

```go
// ❌ 错误：接口过大
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id string) (*domain.User, error)
    // ... 30个方法
}

// ✅ 正确：按需定义接口
type UserCreator interface {
    Create(ctx context.Context, user *domain.User) error
}

type UserGetter interface {
    GetByID(ctx context.Context, id string) (*domain.User, error)
}

// 使用组合
type UserRepository interface {
    UserCreator
    UserGetter
    // 其他接口...
}
```

---

## 依赖注入实现模式

### 模式1：构造函数注入（推荐）

通过构造函数传入依赖，这是最推荐的方式。

**示例**：

```go
// 定义接口
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id string) (*domain.User, error)
}

// 定义用例结构
type userUseCase struct {
    userRepo UserRepository  // 依赖接口
}

// 构造函数注入
func NewUserUseCase(userRepo UserRepository) UserUseCase {
    return &userUseCase{
        userRepo: userRepo,  // 注入依赖
    }
}

// 使用
func (uc *userUseCase) CreateUser(ctx context.Context, username, email string) (*domain.User, error) {
    user := domain.NewUser(username, email)
    if err := uc.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }
    return user, nil
}
```

**优点**：
- 依赖关系明确
- 强制提供依赖（编译时检查）
- 不可变性（依赖在创建后不可更改）

---

### 模式2：选项模式（可选）

对于有大量可选依赖的情况，可以使用选项模式。

**示例**：

```go
// 选项函数类型
type DataOption func(*Data)

// 选项函数
func WithPostgreSQL(db *gorm.DB) DataOption {
    return func(d *Data) {
        d.pgDB = db
        d.UserRepo = NewUserPgRepository(db)
    }
}

func WithMongoDB(client *db.MongoClient) DataOption {
    return func(d *Data) {
        d.mongoClient = client
        d.UserRepo = NewUserMongoRepository(client)
    }
}

func WithRedis(client *cache.RedisClient) DataOption {
    return func(d *Data) {
        d.redisClient = client
    }
}

// 构造函数
func NewData(opts ...DataOption) *Data {
    d := &Data{}
    for _, opt := range opts {
        opt(d)
    }
    return d
}

// 使用
data := NewData(
    WithPostgreSQL(pgDB),
    WithRedis(redisClient),
)
```

**适用场景**：
- 依赖是可选的
- 配置项很多
- 需要向后兼容

---

## 完整的依赖注入示例

### 场景：User Service

以下是一个完整的user-service依赖注入示例。

### 1. 定义接口

```go
// internal/user-service/data/user_repo.go
package data

import (
    "context"
    "github.com/alfredchaos/demo/internal/user-service/domain"
)

// UserRepository 用户仓库接口
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id string) (*domain.User, error)
    GetByUsername(ctx context.Context, username string) (*domain.User, error)
}
```

```go
// internal/user-service/biz/user_usecase.go
package biz

import (
    "context"
    "github.com/alfredchaos/demo/internal/user-service/domain"
)

// UserUseCase 用户用例接口
type UserUseCase interface {
    CreateUser(ctx context.Context, username, email string) (*domain.User, error)
    GetUser(ctx context.Context, id string) (*domain.User, error)
}
```

### 2. 实现接口

```go
// internal/user-service/data/user_pg_repo.go
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

```go
// internal/user-service/biz/user_usecase.go
package biz

import (
    "context"
    "github.com/alfredchaos/demo/internal/user-service/data"
    "github.com/alfredchaos/demo/internal/user-service/domain"
)

// userUseCase 用户用例实现
type userUseCase struct {
    userRepo data.UserRepository  // 依赖接口
}

// NewUserUseCase 创建用户用例
func NewUserUseCase(userRepo data.UserRepository) UserUseCase {
    return &userUseCase{
        userRepo: userRepo,
    }
}

func (uc *userUseCase) CreateUser(ctx context.Context, username, email string) (*domain.User, error) {
    user := domain.NewUser(username, email)
    if err := user.Validate(); err != nil {
        return nil, err
    }
    if err := uc.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }
    return user, nil
}
```

### 3. 数据层容器

```go
// internal/user-service/data/data.go
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
    // 基础设施客户端
    pgDB        *gorm.DB
    mongoClient *db.MongoClient
    redisClient *cache.RedisClient
    mqClient    *mq.RabbitMQClient
    
    // 仓库实例（导出，供外部使用）
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
    
    // 根据可用的数据源选择仓库实现
    if pgDB != nil {
        d.UserRepo = NewUserPgRepository(pgDB)
    } else if mongoClient != nil {
        d.UserRepo = NewUserMongoRepository(mongoClient)
    } else {
        return nil, fmt.Errorf("no database configured")
    }
    
    return d, nil
}

// Close 关闭所有连接
func (d *Data) Close(ctx context.Context) error {
    // 关闭数据库连接
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

### 4. 主函数中的依赖注入

```go
// cmd/user-service/main.go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/alfredchaos/demo/internal/user-service/biz"
    "github.com/alfredchaos/demo/internal/user-service/conf"
    "github.com/alfredchaos/demo/internal/user-service/data"
    "github.com/alfredchaos/demo/internal/user-service/server"
    "github.com/alfredchaos/demo/internal/user-service/service"
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
    config.MustLoadConfig("user-service", &cfg)
    
    // 2. 初始化日志
    log.MustInitLogger(&cfg.Log, cfg.Server.Name)
    defer log.Sync()
    
    log.Info("starting user-service", zap.String("name", cfg.Server.Name))
    
    // 3. 初始化基础设施
    // PostgreSQL（可选）
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
    
    // MongoDB（可选）
    var mongoClient *db.MongoClient
    if cfg.MongoDB.Enabled {
        mongoClient = db.MustNewMongoClient(&cfg.MongoDB)
        defer mongoClient.Close(context.Background())
    }
    
    // Redis（可选）
    var redisClient *cache.RedisClient
    if cfg.Redis.Enabled {
        redisClient = cache.MustNewRedisClient(&cfg.Redis)
        defer redisClient.Close()
    }
    
    // RabbitMQ（可选）
    var mqClient *mq.RabbitMQClient
    if cfg.RabbitMQ.Enabled {
        mqClient = mq.MustNewRabbitMQClient(&cfg.RabbitMQ)
        defer mqClient.Close()
    }
    
    // 4. 初始化数据访问层（依赖注入：基础设施 -> Data）
    dataLayer, err := data.NewData(pgDB, mongoClient, redisClient, mqClient)
    if err != nil {
        log.Fatal("failed to initialize data layer", zap.Error(err))
    }
    defer dataLayer.Close(context.Background())
    
    // 5. 初始化业务逻辑层（依赖注入：Data -> Biz）
    userUseCase := biz.NewUserUseCase(dataLayer.UserRepo)
    
    // 6. 初始化服务层（依赖注入：Biz -> Service）
    userService := service.NewUserService(userUseCase)
    
    // 7. 初始化gRPC服务器（依赖注入：Service -> Server）
    grpcServer := server.NewGRPCServer(&cfg.Server, userService)
    
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
    
    log.Info("shutting down user-service")
    grpcServer.Stop()
    log.Info("user-service stopped")
}
```

---

## 依赖注入链路

```
main.go
  │
  ├─> 初始化配置 (config.MustLoadConfig)
  │
  ├─> 初始化日志 (log.MustInitLogger)
  │
  ├─> 初始化基础设施
  │   ├─> PostgreSQL (db.MustNewPostgresDB)
  │   ├─> MongoDB (db.MustNewMongoClient)
  │   ├─> Redis (cache.MustNewRedisClient)
  │   └─> RabbitMQ (mq.MustNewRabbitMQClient)
  │
  ├─> 初始化数据层 (data.NewData)
  │   └─> 注入: PostgreSQL, MongoDB, Redis, RabbitMQ
  │       └─> 创建: UserRepo, BookRepo, etc.
  │
  ├─> 初始化业务层 (biz.NewUserUseCase)
  │   └─> 注入: UserRepo
  │
  ├─> 初始化服务层 (service.NewUserService)
  │   └─> 注入: UserUseCase
  │
  └─> 初始化服务器 (server.NewGRPCServer)
      └─> 注入: UserService
```

---

## 复杂场景：跨服务调用

> **重要更新**：现在使用统一的 `pkg/grpcclient` 模块管理跨服务调用。

当一个服务需要调用其他服务时（东西向通信），通过依赖注入gRPC客户端实现。

**示例**：user-service需要调用book-service

### 1. 添加配置结构

首先，在服务的配置结构中添加gRPC客户端配置：

```go
// internal/user-service/conf/config.go
package conf

import (
    "github.com/alfredchaos/demo/pkg/cache"
    "github.com/alfredchaos/demo/pkg/db"
    "github.com/alfredchaos/demo/pkg/grpcclient"
    "github.com/alfredchaos/demo/pkg/log"
    "github.com/alfredchaos/demo/pkg/mq"
)

type Config struct {
    Server      ServerConfig       `yaml:"server" mapstructure:"server"`
    Log         log.LogConfig      `yaml:"log" mapstructure:"log"`
    Database    db.Config          `yaml:"database" mapstructure:"database"`
    MongoDB     db.MongoConfig     `yaml:"mongodb" mapstructure:"mongodb"`
    Redis       cache.RedisConfig  `yaml:"redis" mapstructure:"redis"`
    RabbitMQ    mq.RabbitMQConfig  `yaml:"rabbitmq" mapstructure:"rabbitmq"`
    GRPCClients grpcclient.Config  `yaml:"grpc_clients" mapstructure:"grpc_clients"` // gRPC客户端配置
}
```

### 2. 注册客户端工厂

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

### 3. 在UseCase中注入客户端

直接使用生成的gRPC客户端类型，无需额外封装：

```go
// internal/user-service/biz/user_usecase.go
package biz

import (
    "context"
    bookv1 "github.com/alfredchaos/demo/api/book/v1"
    "github.com/alfredchaos/demo/internal/user-service/data"
    "github.com/alfredchaos/demo/internal/user-service/domain"
)

// userUseCase 用户用例实现
type userUseCase struct {
    userRepo   data.UserRepository
    bookClient bookv1.BookServiceClient  // 直接使用生成的客户端
}

// NewUserUseCase 创建用户用例
func NewUserUseCase(
    userRepo data.UserRepository,
    bookClient bookv1.BookServiceClient,  // 依赖注入
) UserUseCase {
    return &userUseCase{
        userRepo:   userRepo,
        bookClient: bookClient,
    }
}

// GetUserWithRecommendation 获取用户及推荐书籍
func (uc *userUseCase) GetUserWithRecommendation(ctx context.Context, id string) (*domain.UserWithBook, error) {
    // 1. 获取用户
    user, err := uc.userRepo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // 2. 调用书籍服务获取推荐
    var recommendation string
    if uc.bookClient != nil {
        resp, err := uc.bookClient.GetRecommendation(ctx, &bookv1.GetRecommendationRequest{
            UserId: id,
        })
        if err != nil {
            // 处理错误，可以返回默认值或失败
            recommendation = "No recommendation available"
        } else {
            recommendation = resp.Recommendation
        }
    } else {
        recommendation = "Book service not configured"
    }
    
    return &domain.UserWithBook{
        User:           user,
        Recommendation: recommendation,
    }, nil
}
```

### 4. 在main.go中初始化客户端

使用 `pkg/grpcclient` 模块统一管理客户端连接：

```go
// cmd/user-service/main.go
func main() {
    // 1. 加载配置
    var cfg conf.Config
    config.MustLoadConfig("user-service", &cfg)
    
    // 2. 初始化日志
    log.MustInitLogger(&cfg.Log, cfg.Server.Name)
    defer log.Sync()
    
    // 3. 初始化gRPC客户端管理器（如果需要调用其他服务）
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
    
    // 4. 初始化基础设施（PostgreSQL, MongoDB, Redis, RabbitMQ）
    // ...
    
    // 5. 初始化数据访问层
    dataLayer, err := data.NewData(pgDB, mongoClient, redisClient, mqClient)
    if err != nil {
        log.Fatal("failed to initialize data layer", zap.Error(err))
    }
    defer dataLayer.Close(context.Background())
    
    // 6. 获取gRPC客户端
    var bookClient bookv1.BookServiceClient
    if clientManager != nil {
        bookConn, err := clientManager.GetConnection("book-service")
        if err != nil {
            log.Warn("book-service not configured", zap.Error(err))
        } else {
            bookClient = bookv1.NewBookServiceClient(bookConn)
        }
    }
    
    // 7. 初始化业务逻辑层（注入gRPC客户端）
    userUseCase := biz.NewUserUseCase(
        dataLayer.UserRepo,
        bookClient,  // 注入客户端
    )
    
    // 8. 初始化服务层
    userService := service.NewUserService(userUseCase)
    
    // 9. 初始化gRPC服务器
    grpcServer := server.NewGRPCServer(&cfg.Server, userService)
    
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
    
    log.Info("shutting down user-service")
    grpcServer.Stop()
    log.Info("user-service stopped")
}
```

### 5. 配置文件示例

```yaml
# configs/user-service.yaml
server:
  name: user-service
  host: 0.0.0.0
  port: 9001

log:
  level: debug

# ... 其他配置 ...

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
```

### 6. 依赖注入链路图

```
main.go
  │
  ├─> 初始化配置
  │
  ├─> 初始化gRPC客户端管理器 (grpcclient.Manager)
  │   └─> 注册服务配置
  │   └─> 连接所有服务
  │
  ├─> 初始化基础设施 (PostgreSQL, MongoDB, Redis, RabbitMQ)
  │
  ├─> 初始化数据层 (data.NewData)
  │   └─> 注入: PostgreSQL, MongoDB, Redis, RabbitMQ
  │
  ├─> 获取gRPC客户端 (clientManager.GetConnection)
  │   └─> bookClient := bookv1.NewBookServiceClient(conn)
  │
  ├─> 初始化业务层 (biz.NewUserUseCase)
  │   └─> 注入: UserRepo, BookClient
  │
  ├─> 初始化服务层 (service.NewUserService)
  │   └─> 注入: UserUseCase
  │
  └─> 初始化服务器 (server.NewGRPCServer)
      └─> 注入: UserService
```

---

## 测试中的依赖注入

手动依赖注入的一个巨大优势是易于测试。

### Mock实现

```go
// internal/user-service/data/user_repo_mock.go
package data

import (
    "context"
    "github.com/alfredchaos/demo/internal/user-service/domain"
)

// MockUserRepository Mock仓库实现
type MockUserRepository struct {
    CreateFunc      func(ctx context.Context, user *domain.User) error
    GetByIDFunc     func(ctx context.Context, id string) (*domain.User, error)
    GetByUsernameFunc func(ctx context.Context, username string) (*domain.User, error)
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
    if m.CreateFunc != nil {
        return m.CreateFunc(ctx, user)
    }
    return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
    if m.GetByIDFunc != nil {
        return m.GetByIDFunc(ctx, id)
    }
    return nil, domain.ErrUserNotFound
}
```

### 单元测试

```go
// internal/user-service/biz/user_usecase_test.go
package biz

import (
    "context"
    "testing"
    
    "github.com/alfredchaos/demo/internal/user-service/data"
    "github.com/alfredchaos/demo/internal/user-service/domain"
    "github.com/stretchr/testify/assert"
)

func TestUserUseCase_CreateUser(t *testing.T) {
    // 创建Mock仓库
    mockRepo := &data.MockUserRepository{
        GetByUsernameFunc: func(ctx context.Context, username string) (*domain.User, error) {
            return nil, domain.ErrUserNotFound  // 模拟用户不存在
        },
        CreateFunc: func(ctx context.Context, user *domain.User) error {
            return nil  // 模拟成功创建
        },
    }
    
    // 注入Mock仓库
    useCase := NewUserUseCase(mockRepo)
    
    // 测试
    user, err := useCase.CreateUser(context.Background(), "testuser", "test@example.com")
    
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "testuser", user.Username)
}
```

---

## 最佳实践

### 1. 接口定义位置

**规则**：接口应该定义在使用者一侧，而不是实现者一侧

```go
// ✅ 正确：接口定义在biz层（使用者）
// internal/user-service/biz/user_usecase.go
package biz

type UserRepository interface {  // 定义在使用者包中
    Create(ctx context.Context, user *domain.User) error
}

// ❌ 错误：接口定义在data层（实现者）
// internal/user-service/data/user_repo.go
package data

type UserRepository interface {  // 不应该在实现者包中
    Create(ctx context.Context, user *domain.User) error
}
```

**原因**：遵循依赖倒置原则，使用者定义需要的接口，实现者去满足接口。

**实际应用**：为了方便管理，我们将所有仓库接口统一放在`data`包中，但从设计上它们属于业务层的需求。

### 2. 避免循环依赖

```go
// ❌ 错误：循环依赖
// package A imports package B
// package B imports package A

// ✅ 正确：通过接口解耦
// package A 定义接口，package B 实现接口
```

### 3. 最小化公开接口

```go
// ✅ 正确：只公开必要的方法
type UserReader interface {
    GetByID(ctx context.Context, id string) (*domain.User, error)
}

// ❌ 错误：公开所有方法
type UserRepository interface {
    Create(...)
    Update(...)
    Delete(...)
    GetByID(...)
    // ... 20个方法
}
```

### 4. 使用构造函数

```go
// ✅ 正确：使用构造函数
func NewUserUseCase(repo UserRepository) UserUseCase {
    return &userUseCase{repo: repo}
}

// ❌ 错误：直接创建
uc := &userUseCase{}
uc.repo = repo
```

---

## 总结

手动依赖注入虽然需要编写更多代码，但在微服务场景下具有以下优势：

1. **透明性**：依赖关系一目了然
2. **可调试性**：容易定位问题
3. **可测试性**：容易编写单元测试
4. **灵活性**：可以根据配置动态选择实现

通过遵循依赖倒置原则和接口隔离原则，我们可以构建松耦合、易测试、易维护的代码。

---

## 参考资料

- [ai-guide-doc/prompt/di.md](../prompt/di.md) - 依赖注入示例
- [ARCHITECTURE.md](./ARCHITECTURE.md) - 分层架构设计
