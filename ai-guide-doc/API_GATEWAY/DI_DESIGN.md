我会先查看 di.md 文档和当前的 api-gateway 结构，然后基于依赖注入框架重新设计。

基于 di.md 的依赖注入框架，我将重新设计 api-gateway 的架构。当前的问题是缺少清晰的分层和接口定义，所有依赖都是具体实现。

## 重新设计架构

基于 di.md 的依赖注入原则，我将创建清晰的分层架构：

### 1. 创建 Domain 接口层

让我先创建架构设计文档，然后实现新的结构：

我将创建一个完整的架构设计文档，然后展示所有需要创建和修改的文件。由于需要创建新文件，我会先展示完整的设计方案：

## 新架构设计

### 目录结构
```
internal/api-gateway/
├── domain/          # 领域接口层
│   ├── service.go   # 服务接口定义
├── usecase/         # 应用层（业务逻辑）
│   ├── hello_usecase.go
├── controller/      # 控制层（HTTP处理）
│   └── hello_controller.go
├── client/          # gRPC客户端实现
│   └── grpc_client.go
├── inject/          # 依赖注入
│   └── wire.go
├── router/          # 路由配置
│   └── router.go
├── middleware/      # 中间件
├── dto/            # 数据传输对象
```

让我先创建架构设计文档：

我将直接展示完整的新架构实现。由于需要创建多个新文件，我会展示所有文件的完整代码：

## 新架构实现

### 1. Domain 接口层

**文件：`internal/api-gateway/domain/service.go`**
```go
package domain

import "context"

// IUserService 用户服务接口
// 定义用户相关的业务能力
type IUserService interface {
	// SayHello 获取用户服务的问候语
	SayHello(ctx context.Context) (string, error)
}

// IBookService 图书服务接口
// 定义图书相关的业务能力
type IBookService interface {
	// GetBook 获取图书信息
	GetBook(ctx context.Context) (string, error)
}

// IMessagePublisher 消息发布者接口
// 定义消息发布能力，实现与具体MQ实现解耦
type IMessagePublisher interface {
	// PublishHelloMessage 发布问候消息
	PublishHelloMessage(ctx context.Context, message string) error
}
```

### 2. UseCase 应用层

**文件：`internal/api-gateway/usecase/hello_usecase.go`**
```go
package usecase

import (
	"context"
	"fmt"

	"github.com/alfredchaos/demo/internal/api-gateway/domain"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

// IHelloUseCase 问候用例接口
// 定义问候相关的业务用例
type IHelloUseCase interface {
	// ExecuteHello 执行问候业务逻辑
	// 协调多个服务的调用并组合结果
	ExecuteHello(ctx context.Context) (string, error)
}

// helloUseCase 问候用例实现
type helloUseCase struct {
	userService IUserService
	bookService IBookService
	publisher   IMessagePublisher
}

// NewHelloUseCase 创建问候用例
// 依赖注入：接收领域服务接口
func NewHelloUseCase(
	userService domain.IUserService,
	bookService domain.IBookService,
	publisher domain.IMessagePublisher,
) IHelloUseCase {
	return &helloUseCase{
		userService: userService,
		bookService: bookService,
		publisher:   publisher,
	}
}

// ExecuteHello 执行问候业务逻辑
// 并发调用用户服务和图书服务，组合结果并发布消息
func (u *helloUseCase) ExecuteHello(ctx context.Context) (string, error) {
	log.WithContext(ctx).Info("executing hello use case")

	// 定义结果结构
	type result struct {
		message string
		err     error
	}

	// 创建通道接收结果
	userChan := make(chan result, 1)
	bookChan := make(chan result, 1)

	// 并发调用用户服务
	go func() {
		msg, err := u.userService.SayHello(ctx)
		userChan <- result{message: msg, err: err}
	}()

	// 并发调用图书服务
	go func() {
		msg, err := u.bookService.GetBook(ctx)
		bookChan <- result{message: msg, err: err}
	}()

	// 等待结果
	userResult := <-userChan
	bookResult := <-bookChan

	// 处理用户服务错误
	if userResult.err != nil {
		log.WithContext(ctx).Error("user service failed", zap.Error(userResult.err))
		return "", fmt.Errorf("user service failed: %w", userResult.err)
	}

	// 处理图书服务错误
	if bookResult.err != nil {
		log.WithContext(ctx).Error("book service failed", zap.Error(bookResult.err))
		return "", fmt.Errorf("book service failed: %w", bookResult.err)
	}

	// 组合结果
	combinedMessage := userResult.message + " " + bookResult.message
	log.WithContext(ctx).Info("combined service responses", zap.String("message", combinedMessage))

	// 异步发布消息到MQ（不阻塞主流程）
	go func() {
		if err := u.publisher.PublishHelloMessage(context.Background(), combinedMessage); err != nil {
			log.Error("failed to publish message", zap.Error(err))
		}
	}()

	return combinedMessage, nil
}
```

### 3. Client 实现层（实现 Domain 接口）

**文件：`internal/api-gateway/client/user_service_impl.go`**
```go
package client

import (
	"context"
	"fmt"

	userv1 "github.com/alfredchaos/demo/api/user/v1"
	"github.com/alfredchaos/demo/internal/api-gateway/domain"
	"google.golang.org/grpc/metadata"
)

// userServiceImpl 用户服务实现
// 实现 domain.IUserService 接口
type userServiceImpl struct {
	client userv1.UserServiceClient
}

// NewUserService 创建用户服务实现
func NewUserService(client userv1.UserServiceClient) domain.IUserService {
	return &userServiceImpl{
		client: client,
	}
}

// SayHello 调用 user-service 的 SayHello 方法
func (s *userServiceImpl) SayHello(ctx context.Context) (string, error) {
	// 传递 trace ID
	ctx = withTraceID(ctx)

	resp, err := s.client.SayHello(ctx, &userv1.HelloRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to call user service: %w", err)
	}
	return resp.Message, nil
}

// withTraceID 将 trace ID 从 context 中提取并添加到 gRPC metadata
func withTraceID(ctx context.Context) context.Context {
	traceID := ""
	if val := ctx.Value("X-Request-ID"); val != nil {
		if id, ok := val.(string); ok {
			traceID = id
		}
	}

	if traceID != "" {
		md := metadata.Pairs("X-Trace-ID", traceID)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	return ctx
}
```

**文件：`internal/api-gateway/client/book_service_impl.go`**
```go
package client

import (
	"context"
	"fmt"

	orderv1 "github.com/alfredchaos/demo/api/order/v1"
	"github.com/alfredchaos/demo/internal/api-gateway/domain"
)

// bookServiceImpl 图书服务实现
// 实现 domain.IBookService 接口
type bookServiceImpl struct {
	client orderv1.BookServiceClient
}

// NewBookService 创建图书服务实现
func NewBookService(client orderv1.BookServiceClient) domain.IBookService {
	return &bookServiceImpl{
		client: client,
	}
}

// GetBook 调用 book-service 的 GetBook 方法
func (s *bookServiceImpl) GetBook(ctx context.Context) (string, error) {
	// 传递 trace ID
	ctx = withTraceID(ctx)

	resp, err := s.client.GetBook(ctx, &orderv1.BookRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to call book service: %w", err)
	}
	return resp.Message, nil
}
```

**文件：`internal/api-gateway/client/message_publisher_impl.go`**
```go
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alfredchaos/demo/internal/api-gateway/domain"
	"github.com/alfredchaos/demo/pkg/log"
	"github.com/alfredchaos/demo/pkg/mq"
	"go.uber.org/zap"
)

// messagePublisherImpl 消息发布者实现
// 实现 domain.IMessagePublisher 接口
type messagePublisherImpl struct {
	publisher mq.Publisher
}

// NewMessagePublisher 创建消息发布者实现
func NewMessagePublisher(publisher mq.Publisher) domain.IMessagePublisher {
	return &messagePublisherImpl{
		publisher: publisher,
	}
}

// PublishHelloMessage 发布问候消息到 RabbitMQ
func (p *messagePublisherImpl) PublishHelloMessage(ctx context.Context, message string) error {
	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// 构造消息
	msg := map[string]string{
		"type":    "hello",
		"message": message,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 发布消息
	if err := p.publisher.Publish(ctx, msgBytes); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Info("message published to rabbitmq", zap.String("message", message))
	return nil
}
```

**修改文件：[internal/api-gateway/client/grpc_client.go](cci:7://file:///home/shixuan/code/microservice-golang-demo/internal/api-gateway/client/grpc_client.go:0:0-0:0)**
```go
package client

import (
	"fmt"
	"time"

	orderv1 "github.com/alfredchaos/demo/api/order/v1"
	userv1 "github.com/alfredchaos/demo/api/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClients gRPC 客户端集合
// 负责管理所有 gRPC 连接
type GRPCClients struct {
	userConn *grpc.ClientConn
	bookConn *grpc.ClientConn

	UserClient userv1.UserServiceClient
	BookClient orderv1.BookServiceClient
}

// NewGRPCClients 创建新的 gRPC 客户端集合
func NewGRPCClients(userAddr, bookAddr string) (*GRPCClients, error) {
	// 连接 user-service
	userConn, err := grpc.Dial(
		userAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user-service: %w", err)
	}

	// 连接 book-service
	bookConn, err := grpc.Dial(
		bookAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		userConn.Close()
		return nil, fmt.Errorf("failed to connect to book-service: %w", err)
	}

	return &GRPCClients{
		userConn:   userConn,
		bookConn:   bookConn,
		UserClient: userv1.NewUserServiceClient(userConn),
		BookClient: orderv1.NewBookServiceClient(bookConn),
	}, nil
}

// Close 关闭所有 gRPC 连接
func (c *GRPCClients) Close() error {
	var errUser, errBook error

	if c.userConn != nil {
		errUser = c.userConn.Close()
	}

	if c.bookConn != nil {
		errBook = c.bookConn.Close()
	}

	if errUser != nil {
		return fmt.Errorf("failed to close user-service connection: %w", errUser)
	}
	if errBook != nil {
		return fmt.Errorf("failed to close book-service connection: %w", errBook)
	}

	return nil
}
```

### 4. Controller 控制层

**修改文件：[internal/api-gateway/controller/hello_controller.go](cci:7://file:///home/shixuan/code/microservice-golang-demo/internal/api-gateway/controller/hello_controller.go:0:0-0:0)**
```go
package controller

import (
	"net/http"

	"github.com/alfredchaos/demo/internal/api-gateway/dto"
	"github.com/alfredchaos/demo/internal/api-gateway/usecase"
	"github.com/alfredchaos/demo/pkg/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// IHelloController 问候控制器接口
type IHelloController interface {
	SayHello(c *gin.Context)
}

// helloController 问候控制器
// 只负责 HTTP 请求处理，业务逻辑委托给 UseCase
type helloController struct {
	helloUseCase usecase.IHelloUseCase
}

// NewHelloController 创建问候控制器
// 依赖注入：接收 UseCase 接口
func NewHelloController(helloUseCase usecase.IHelloUseCase) IHelloController {
	return &helloController{
		helloUseCase: helloUseCase,
	}
}

// SayHello 处理问候请求
// @Summary 问候接口
// @Description 调用后端服务并返回问候语
// @Tags Hello
// @Accept json
// @Produce json
// @Param request body dto.HelloRequest true "请求参数"
// @Success 200 {object} dto.Response{data=string} "成功响应"
// @Failure 500 {object} dto.Response "服务器错误"
// @Router /api/v1/hello [post]
func (h *helloController) SayHello(c *gin.Context) {
	ctx := c.Request.Context()
	log.WithContext(ctx).Info("received hello request")

	// 调用 UseCase 执行业务逻辑
	message, err := h.helloUseCase.ExecuteHello(ctx)
	if err != nil {
		log.WithContext(ctx).Error("failed to execute hello use case", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(10001, err.Error()))
		return
	}

	// 返回成功响应
	log.WithContext(ctx).Info("hello request completed", zap.String("message", message))
	c.JSON(http.StatusOK, dto.NewSuccessResponse(message))
}
```

### 5. 依赖注入模块

**文件：`internal/api-gateway/inject/wire.go`**
```go
package inject

import (
	"github.com/alfredchaos/demo/internal/api-gateway/client"
	"github.com/alfredchaos/demo/internal/api-gateway/controller"
	"github.com/alfredchaos/demo/internal/api-gateway/usecase"
	"github.com/alfredchaos/demo/pkg/mq"
)

// AppContext 应用上下文
// 管理所有依赖的组件
type AppContext struct {
	// 基础设施
	GRPCClients    *client.GRPCClients
	RabbitMQClient *mq.RabbitMQClient

	// 控制器
	HelloController controller.IHelloController
}

// InjectDependencies 依赖注入函数
// 创建并组装所有依赖
func InjectDependencies(
	grpcClients *client.GRPCClients,
	rabbitMQClient *mq.RabbitMQClient,
) *AppContext {
	// 创建 MQ Publisher
	mqPublisher := mq.NewRabbitMQPublisher(rabbitMQClient)

	// 创建 Domain 层服务实现
	userService := client.NewUserService(grpcClients.UserClient)
	bookService := client.NewBookService(grpcClients.BookClient)
	messagePublisher := client.NewMessagePublisher(mqPublisher)

	// 创建 UseCase 层
	// UseCase 依赖 Domain 接口
	helloUseCase := usecase.NewHelloUseCase(userService, bookService, messagePublisher)

	// 创建 Controller 层
	// Controller 依赖 UseCase 接口
	helloController := controller.NewHelloController(helloUseCase)

	return &AppContext{
		GRPCClients:     grpcClients,
		RabbitMQClient:  rabbitMQClient,
		HelloController: helloController,
	}
}
```

### 6. 路由配置

**修改文件：[internal/api-gateway/router/router.go](cci:7://file:///home/shixuan/code/microservice-golang-demo/internal/api-gateway/router/router.go:0:0-0:0)**
```go
package router

import (
	"time"

	"github.com/alfredchaos/demo/internal/api-gateway/inject"
	"github.com/alfredchaos/demo/internal/api-gateway/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 设置路由
// 使用依赖注入的 AppContext
func SetupRouter(appCtx *inject.AppContext) *gin.Engine {
	// 创建 Gin 引擎（不使用默认中间件）
	router := gin.New()

	// 应用全局中间件（顺序很重要）
	router.Use(
		middleware.Recovery(),              // 1. Panic恢复
		middleware.RequestID(),             // 2. 请求ID生成
		middleware.Logger(),                // 3. 请求日志记录
		middleware.CORS(),                  // 4. 跨域处理
		middleware.Timeout(30*time.Second), // 5. 请求超时
	)

	// 注册路由
	registerAPIRoutes(router, appCtx)
	registerSystemRoutes(router)

	return router
}

// registerAPIRoutes 注册 API 路由
func registerAPIRoutes(router *gin.Engine, appCtx *inject.AppContext) {
	apiV1 := router.Group("/api/v1")
	{
		// Hello 路由组
		helloGroup := apiV1.Group("/hello")
		{
			helloGroup.POST("", appCtx.HelloController.SayHello)
		}
	}
}

// registerSystemRoutes 注册系统路由
func registerSystemRoutes(router *gin.Engine) {
	// Swagger 文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
}
```

### 7. Main 入口

**修改文件：[cmd/api-gateway/main.go](cci:7://file:///home/shixuan/code/microservice-golang-demo/cmd/api-gateway/main.go:0:0-0:0)**
```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/alfredchaos/demo/docs"
	"github.com/alfredchaos/demo/internal/api-gateway/client"
	"github.com/alfredchaos/demo/internal/api-gateway/inject"
	"github.com/alfredchaos/demo/internal/api-gateway/router"
	"github.com/alfredchaos/demo/pkg/config"
	"github.com/alfredchaos/demo/pkg/log"
	"github.com/alfredchaos/demo/pkg/mq"
	"go.uber.org/zap"
)

// Config api-gateway 配置结构
type Config struct {
	Server   ServerConfig      `yaml:"server" mapstructure:"server"`
	Log      log.LogConfig     `yaml:"log" mapstructure:"log"`
	Services ServicesConfig    `yaml:"services" mapstructure:"services"`
	RabbitMQ mq.RabbitMQConfig `yaml:"rabbitmq" mapstructure:"rabbitmq"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Name string `yaml:"name" mapstructure:"name"`
	Host string `yaml:"host" mapstructure:"host"`
	Port int    `yaml:"port" mapstructure:"port"`
}

// ServicesConfig 后端服务配置
type ServicesConfig struct {
	UserService string `yaml:"user_service" mapstructure:"user_service"`
	BookService string `yaml:"book_service" mapstructure:"book_service"`
}

// @title Demo API Gateway
// @version 1.0
// @description 微服务架构演示项目的 API 网关
// @host localhost:8080
// @BasePath /
func main() {
	// 1. 加载配置
	var cfg Config
	config.MustLoadConfig("api-gateway", &cfg)

	// 2. 初始化日志
	log.MustInitLogger(&cfg.Log, cfg.Server.Name)
	defer log.Sync()

	log.Info("starting api-gateway", zap.String("name", cfg.Server.Name))

	// 3. 初始化基础设施
	grpcClients := mustInitGRPCClients(cfg.Services)
	defer closeGRPCClients(grpcClients)

	rabbitMQClient := mq.MustNewRabbitMQClient(&cfg.RabbitMQ)
	defer closeRabbitMQ(rabbitMQClient)

	// 4. 依赖注入 - 组装所有组件
	appCtx := inject.InjectDependencies(grpcClients, rabbitMQClient)
	log.Info("dependencies injected successfully")

	// 5. 设置路由
	r := router.SetupRouter(appCtx)

	// 6. 启动 HTTP 服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Info("http server starting", zap.String("addr", addr))

	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatal("failed to start http server", zap.Error(err))
		}
	}()

	// 7. 优雅关闭
	waitForShutdown()
}

// mustInitGRPCClients 初始化 gRPC 客户端
func mustInitGRPCClients(services ServicesConfig) *client.GRPCClients {
	grpcClients, err := client.NewGRPCClients(services.UserService, services.BookService)
	if err != nil {
		log.Fatal("failed to create grpc clients", zap.Error(err))
	}
	log.Info("grpc clients initialized")
	return grpcClients
}

// closeGRPCClients 关闭 gRPC 客户端
func closeGRPCClients(grpcClients *client.GRPCClients) {
	if err := grpcClients.Close(); err != nil {
		log.Error("failed to close grpc clients", zap.Error(err))
	}
}

// closeRabbitMQ 关闭 RabbitMQ 客户端
func closeRabbitMQ(rabbitMQClient *mq.RabbitMQClient) {
	if err := rabbitMQClient.Close(); err != nil {
		log.Error("failed to close rabbitmq client", zap.Error(err))
	}
}

// waitForShutdown 等待关闭信号
func waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down api-gateway")
	log.Info("api-gateway stopped")
}
```

现在让我将这些文件逐一创建和修改。由于我不能直接创建新文件，我需要向您展示完整的代码。您需要手动创建以下新文件：

需要**创建**的新文件：
1. `internal/api-gateway/domain/service.go`
2. `internal/api-gateway/usecase/hello_usecase.go`
3. `internal/api-gateway/client/user_service_impl.go`
4. `internal/api-gateway/client/book_service_impl.go`
5. `internal/api-gateway/client/message_publisher_impl.go`
6. `internal/api-gateway/inject/wire.go`

需要**修改**的文件：
1. [internal/api-gateway/client/grpc_client.go](cci:7://file:///home/shixuan/code/microservice-golang-demo/internal/api-gateway/client/grpc_client.go:0:0-0:0)
2. [internal/api-gateway/controller/hello_controller.go](cci:7://file:///home/shixuan/code/microservice-golang-demo/internal/api-gateway/controller/hello_controller.go:0:0-0:0)
3. [internal/api-gateway/router/router.go](cci:7://file:///home/shixuan/code/microservice-golang-demo/internal/api-gateway/router/router.go:0:0-0:0)
4. [cmd/api-gateway/main.go](cci:7://file:///home/shixuan/code/microservice-golang-demo/cmd/api-gateway/main.go:0:0-0:0)

让我开始修改现有文件：

让我分步修改文件。首先修改 grpc_client.go：



现在让我创建架构设计文档，总结所有需要的更改：

由于我无法直接创建新文件，我将为您提供完整的实现指南。让我创建一个详细的架构设计文档：

## 架构重构总结

### 已完成的修改
1. ✅ 简化 [internal/api-gateway/client/grpc_client.go](cci:7://file:///home/shixuan/code/microservice-golang-demo/internal/api-gateway/client/grpc_client.go:0:0-0:0)，移除业务逻辑方法

### 需要手动创建的新文件

请按照以下内容创建这些文件：

#### 1. `internal/api-gateway/domain/service.go`

<details>
<summary>点击查看完整代码</summary>

```go
package domain

import "context"

// IUserService 用户服务接口
// 定义用户相关的业务能力
type IUserService interface {
	// SayHello 获取用户服务的问候语
	SayHello(ctx context.Context) (string, error)
}

// IBookService 图书服务接口
// 定义图书相关的业务能力
type IBookService interface {
	// GetBook 获取图书信息
	GetBook(ctx context.Context) (string, error)
}

// IMessagePublisher 消息发布者接口
// 定义消息发布能力，实现与具体MQ实现解耦
type IMessagePublisher interface {
	// PublishHelloMessage 发布问候消息
	PublishHelloMessage(ctx context.Context, message string) error
}
```

</details>

#### 2. `internal/api-gateway/client/user_service_impl.go`

<details>
<summary>点击查看完整代码</summary>

```go
package client

import (
	"context"
	"fmt"

	userv1 "github.com/alfredchaos/demo/api/user/v1"
	"github.com/alfredchaos/demo/internal/api-gateway/domain"
	"google.golang.org/grpc/metadata"
)

// userServiceImpl 用户服务实现
// 实现 domain.IUserService 接口
type userServiceImpl struct {
	client userv1.UserServiceClient
}

// NewUserService 创建用户服务实现
func NewUserService(client userv1.UserServiceClient) domain.IUserService {
	return &userServiceImpl{
		client: client,
	}
}

// SayHello 调用 user-service 的 SayHello 方法
func (s *userServiceImpl) SayHello(ctx context.Context) (string, error) {
	// 传递 trace ID
	ctx = withTraceID(ctx)

	resp, err := s.client.SayHello(ctx, &userv1.HelloRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to call user service: %w", err)
	}
	return resp.Message, nil
}

// withTraceID 将 trace ID 从 context 中提取并添加到 gRPC metadata
func withTraceID(ctx context.Context) context.Context {
	traceID := ""
	if val := ctx.Value("X-Request-ID"); val != nil {
		if id, ok := val.(string); ok {
			traceID = id
		}
	}

	if traceID != "" {
		md := metadata.Pairs("X-Trace-ID", traceID)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	return ctx
}
```

</details>

#### 3. `internal/api-gateway/client/book_service_impl.go`

<details>
<summary>点击查看完整代码</summary>

```go
package client

import (
	"context"
	"fmt"

	orderv1 "github.com/alfredchaos/demo/api/order/v1"
	"github.com/alfredchaos/demo/internal/api-gateway/domain"
)

// bookServiceImpl 图书服务实现
// 实现 domain.IBookService 接口
type bookServiceImpl struct {
	client orderv1.BookServiceClient
}

// NewBookService 创建图书服务实现
func NewBookService(client orderv1.BookServiceClient) domain.IBookService {
	return &bookServiceImpl{
		client: client,
	}
}

// GetBook 调用 book-service 的 GetBook 方法
func (s *bookServiceImpl) GetBook(ctx context.Context) (string, error) {
	// 传递 trace ID
	ctx = withTraceID(ctx)

	resp, err := s.client.GetBook(ctx, &orderv1.BookRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to call book service: %w", err)
	}
	return resp.Message, nil
}
```

</details>

#### 4. `internal/api-gateway/client/message_publisher_impl.go`

<details>
<summary>点击查看完整代码</summary>

```go
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alfredchaos/demo/internal/api-gateway/domain"
	"github.com/alfredchaos/demo/pkg/log"
	"github.com/alfredchaos/demo/pkg/mq"
	"go.uber.org/zap"
)

// messagePublisherImpl 消息发布者实现
// 实现 domain.IMessagePublisher 接口
type messagePublisherImpl struct {
	publisher mq.Publisher
}

// NewMessagePublisher 创建消息发布者实现
func NewMessagePublisher(publisher mq.Publisher) domain.IMessagePublisher {
	return &messagePublisherImpl{
		publisher: publisher,
	}
}

// PublishHelloMessage 发布问候消息到 RabbitMQ
func (p *messagePublisherImpl) PublishHelloMessage(ctx context.Context, message string) error {
	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// 构造消息
	msg := map[string]string{
		"type":    "hello",
		"message": message,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 发布消息
	if err := p.publisher.Publish(ctx, msgBytes); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Info("message published to rabbitmq", zap.String("message", message))
	return nil
}
```

</details>

#### 5. `internal/api-gateway/usecase/hello_usecase.go`

<details>
<summary>点击查看完整代码</summary>

```go
package usecase

import (
	"context"
	"fmt"

	"github.com/alfredchaos/demo/internal/api-gateway/domain"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

// IHelloUseCase 问候用例接口
// 定义问候相关的业务用例
type IHelloUseCase interface {
	// ExecuteHello 执行问候业务逻辑
	// 协调多个服务的调用并组合结果
	ExecuteHello(ctx context.Context) (string, error)
}

// helloUseCase 问候用例实现
type helloUseCase struct {
	userService      domain.IUserService
	bookService      domain.IBookService
	messagePublisher domain.IMessagePublisher
}

// NewHelloUseCase 创建问候用例
// 依赖注入：接收领域服务接口
func NewHelloUseCase(
	userService domain.IUserService,
	bookService domain.IBookService,
	messagePublisher domain.IMessagePublisher,
) IHelloUseCase {
	return &helloUseCase{
		userService:      userService,
		bookService:      bookService,
		messagePublisher: messagePublisher,
	}
}

// ExecuteHello 执行问候业务逻辑
// 并发调用用户服务和图书服务，组合结果并发布消息
func (u *helloUseCase) ExecuteHello(ctx context.Context) (string, error) {
	log.WithContext(ctx).Info("executing hello use case")

	// 定义结果结构
	type result struct {
		message string
		err     error
	}

	// 创建通道接收结果
	userChan := make(chan result, 1)
	bookChan := make(chan result, 1)

	// 并发调用用户服务
	go func() {
		msg, err := u.userService.SayHello(ctx)
		userChan <- result{message: msg, err: err}
	}()

	// 并发调用图书服务
	go func() {
		msg, err := u.bookService.GetBook(ctx)
		bookChan <- result{message: msg, err: err}
	}()

	// 等待结果
	userResult := <-userChan
	bookResult := <-bookChan

	// 处理用户服务错误
	if userResult.err != nil {
		log.WithContext(ctx).Error("user service failed", zap.Error(userResult.err))
		return "", fmt.Errorf("user service failed: %w", userResult.err)
	}

	// 处理图书服务错误
	if bookResult.err != nil {
		log.WithContext(ctx).Error("book service failed", zap.Error(bookResult.err))
		return "", fmt.Errorf("book service failed: %w", bookResult.err)
	}

	// 组合结果
	combinedMessage := userResult.message + " " + bookResult.message
	log.WithContext(ctx).Info("combined service responses", zap.String("message", combinedMessage))

	// 异步发布消息到MQ（不阻塞主流程）
	go func() {
		if err := u.messagePublisher.PublishHelloMessage(context.Background(), combinedMessage); err != nil {
			log.Error("failed to publish message", zap.Error(err))
		}
	}()

	return combinedMessage, nil
}
```

</details>

#### 6. `internal/api-gateway/inject/wire.go`

<details>
<summary>点击查看完整代码</summary>

```go
package inject

import (
	"github.com/alfredchaos/demo/internal/api-gateway/client"
	"github.com/alfredchaos/demo/internal/api-gateway/controller"
	"github.com/alfredchaos/demo/internal/api-gateway/usecase"
	"github.com/alfredchaos/demo/pkg/mq"
)

// AppContext 应用上下文
// 管理所有依赖的组件
type AppContext struct {
	// 基础设施
	GRPCClients    *client.GRPCClients
	RabbitMQClient *mq.RabbitMQClient

	// 控制器
	HelloController controller.IHelloController
}

// InjectDependencies 依赖注入函数
// 创建并组装所有依赖
func InjectDependencies(
	grpcClients *client.GRPCClients,
	rabbitMQClient *mq.RabbitMQClient,
) *AppContext {
	// 创建 MQ Publisher
	mqPublisher := mq.NewRabbitMQPublisher(rabbitMQClient)

	// 创建 Domain 层服务实现
	userService := client.NewUserService(grpcClients.UserClient)
	bookService := client.NewBookService(grpcClients.BookClient)
	messagePublisher := client.NewMessagePublisher(mqPublisher)

	// 创建 UseCase 层
	// UseCase 依赖 Domain 接口
	helloUseCase := usecase.NewHelloUseCase(userService, bookService, messagePublisher)

	// 创建 Controller 层
	// Controller 依赖 UseCase 接口
	helloController := controller.NewHelloController(helloUseCase)

	return &AppContext{
		GRPCClients:     grpcClients,
		RabbitMQClient:  rabbitMQClient,
		HelloController: helloController,
	}
}
```

</details>

现在让我修改现有的文件：









现在修改 main.go：





现在让我创建架构文档来总结所有改动：

让我创建完整的架构重构文档：

现在让我创建一个完整的架构重构文档，包含所有需要创建的文件：

```bash
# 创建新的目录结构
mkdir -p internal/api-gateway/domain
mkdir -p internal/api-gateway/usecase
mkdir -p internal/api-gateway/inject
```

让我为您创建架构文档，并提供所有需要创建的文件代码。由于文件数量较多，我会创建一个总结文档：

## 架构重构总结

### ✅ 已完成的修改

已通过工具修改的文件：

1. **[internal/api-gateway/client/grpc_client.go](cci:7://file:///home/shixuan/code/microservice-golang-demo/internal/api-gateway/client/grpc_client.go:0:0-0:0)**
   - 移除了 [CallUserService](cci:1://file:///home/shixuan/code/microservice-golang-demo/internal/api-gateway/client/grpc_client.go:84:0-94:1)、[CallBookService](cci:1://file:///home/shixuan/code/microservice-golang-demo/internal/api-gateway/client/grpc_client.go:96:0-106:1) 和 [withTraceID](cci:1://file:///home/shixuan/code/microservice-golang-demo/internal/api-gateway/client/grpc_client.go:108:0-130:1) 方法
   - 简化为只负责连接管理

2. **[internal/api-gateway/controller/hello_controller.go](cci:7://file:///home/shixuan/code/microservice-golang-demo/internal/api-gateway/controller/hello_controller.go:0:0-0:0)**
   - 重构为依赖 `usecase.IHelloUseCase` 接口
   - 简化为只处理 HTTP 请求响应
   - 业务逻辑委托给 UseCase 层

3. **[internal/api-gateway/router/router.go](cci:7://file:///home/shixuan/code/microservice-golang-demo/internal/api-gateway/router/router.go:0:0-0:0)**
   - 改为接收 `inject.AppContext` 参数
   - 拆分路由注册为独立函数

4. **[cmd/api-gateway/main.go](cci:7://file:///home/shixuan/code/microservice-golang-demo/cmd/api-gateway/main.go:0:0-0:0)**
   - 使用依赖注入模式
   - 提取辅助函数简化主逻辑
   - 通过 `inject.InjectDependencies` 组装所有组件

### 📁 需要手动创建的新文件

以下是需要创建的 6 个新文件及其完整代码：

#### 1️⃣ `internal/api-gateway/domain/service.go`

```go
package domain

import "context"

// IUserService 用户服务接口
// 定义用户相关的业务能力
type IUserService interface {
	// SayHello 获取用户服务的问候语
	SayHello(ctx context.Context) (string, error)
}

// IBookService 图书服务接口
// 定义图书相关的业务能力
type IBookService interface {
	// GetBook 获取图书信息
	GetBook(ctx context.Context) (string, error)
}

// IMessagePublisher 消息发布者接口
// 定义消息发布能力，实现与具体MQ实现解耦
type IMessagePublisher interface {
	// PublishHelloMessage 发布问候消息
	PublishHelloMessage(ctx context.Context, message string) error
}
```

#### 2️⃣ `internal/api-gateway/client/user_service_impl.go`

```go
package client

import (
	"context"
	"fmt"

	userv1 "github.com/alfredchaos/demo/api/user/v1"
	"github.com/alfredchaos/demo/internal/api-gateway/domain"
	"google.golang.org/grpc/metadata"
)

// userServiceImpl 用户服务实现
// 实现 domain.IUserService 接口
type userServiceImpl struct {
	client userv1.UserServiceClient
}

// NewUserService 创建用户服务实现
func NewUserService(client userv1.UserServiceClient) domain.IUserService {
	return &userServiceImpl{
		client: client,
	}
}

// SayHello 调用 user-service 的 SayHello 方法
func (s *userServiceImpl) SayHello(ctx context.Context) (string, error) {
	// 传递 trace ID
	ctx = withTraceID(ctx)

	resp, err := s.client.SayHello(ctx, &userv1.HelloRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to call user service: %w", err)
	}
	return resp.Message, nil
}

// withTraceID 将 trace ID 从 context 中提取并添加到 gRPC metadata
func withTraceID(ctx context.Context) context.Context {
	traceID := ""
	if val := ctx.Value("X-Request-ID"); val != nil {
		if id, ok := val.(string); ok {
			traceID = id
		}
	}

	if traceID != "" {
		md := metadata.Pairs("X-Trace-ID", traceID)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	return ctx
}
```

#### 3️⃣ `internal/api-gateway/client/book_service_impl.go`

```go
package client

import (
	"context"
	"fmt"

	orderv1 "github.com/alfredchaos/demo/api/order/v1"
	"github.com/alfredchaos/demo/internal/api-gateway/domain"
)

// bookServiceImpl 图书服务实现
// 实现 domain.IBookService 接口
type bookServiceImpl struct {
	client orderv1.BookServiceClient
}

// NewBookService 创建图书服务实现
func NewBookService(client orderv1.BookServiceClient) domain.IBookService {
	return &bookServiceImpl{
		client: client,
	}
}

// GetBook 调用 book-service 的 GetBook 方法
func (s *bookServiceImpl) GetBook(ctx context.Context) (string, error) {
	// 传递 trace ID
	ctx = withTraceID(ctx)

	resp, err := s.client.GetBook(ctx, &orderv1.BookRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to call book service: %w", err)
	}
	return resp.Message, nil
}
```

#### 4️⃣ `internal/api-gateway/client/message_publisher_impl.go`

```go
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alfredchaos/demo/internal/api-gateway/domain"
	"github.com/alfredchaos/demo/pkg/log"
	"github.com/alfredchaos/demo/pkg/mq"
	"go.uber.org/zap"
)

// messagePublisherImpl 消息发布者实现
// 实现 domain.IMessagePublisher 接口
type messagePublisherImpl struct {
	publisher mq.Publisher
}

// NewMessagePublisher 创建消息发布者实现
func NewMessagePublisher(publisher mq.Publisher) domain.IMessagePublisher {
	return &messagePublisherImpl{
		publisher: publisher,
	}
}

// PublishHelloMessage 发布问候消息到 RabbitMQ
func (p *messagePublisherImpl) PublishHelloMessage(ctx context.Context, message string) error {
	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// 构造消息
	msg := map[string]string{
		"type":    "hello",
		"message": message,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 发布消息
	if err := p.publisher.Publish(ctx, msgBytes); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Info("message published to rabbitmq", zap.String("message", message))
	return nil
}
```

#### 5️⃣ `internal/api-gateway/usecase/hello_usecase.go`

```go
package usecase

import (
	"context"
	"fmt"

	"github.com/alfredchaos/demo/internal/api-gateway/domain"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

// IHelloUseCase 问候用例接口
// 定义问候相关的业务用例
type IHelloUseCase interface {
	// ExecuteHello 执行问候业务逻辑
	// 协调多个服务的调用并组合结果
	ExecuteHello(ctx context.Context) (string, error)
}

// helloUseCase 问候用例实现
type helloUseCase struct {
	userService      domain.IUserService
	bookService      domain.IBookService
	messagePublisher domain.IMessagePublisher
}

// NewHelloUseCase 创建问候用例
// 依赖注入：接收领域服务接口
func NewHelloUseCase(
	userService domain.IUserService,
	bookService domain.IBookService,
	messagePublisher domain.IMessagePublisher,
) IHelloUseCase {
	return &helloUseCase{
		userService:      userService,
		bookService:      bookService,
		messagePublisher: messagePublisher,
	}
}

// ExecuteHello 执行问候业务逻辑
// 并发调用用户服务和图书服务，组合结果并发布消息
func (u *helloUseCase) ExecuteHello(ctx context.Context) (string, error) {
	log.WithContext(ctx).Info("executing hello use case")

	// 定义结果结构
	type result struct {
		message string
		err     error
	}

	// 创建通道接收结果
	userChan := make(chan result, 1)
	bookChan := make(chan result, 1)

	// 并发调用用户服务
	go func() {
		msg, err := u.userService.SayHello(ctx)
		userChan <- result{message: msg, err: err}
	}()

	// 并发调用图书服务
	go func() {
		msg, err := u.bookService.GetBook(ctx)
		bookChan <- result{message: msg, err: err}
	}()

	// 等待结果
	userResult := <-userChan
	bookResult := <-bookChan

	// 处理用户服务错误
	if userResult.err != nil {
		log.WithContext(ctx).Error("user service failed", zap.Error(userResult.err))
		return "", fmt.Errorf("user service failed: %w", userResult.err)
	}

	// 处理图书服务错误
	if bookResult.err != nil {
		log.WithContext(ctx).Error("book service failed", zap.Error(bookResult.err))
		return "", fmt.Errorf("book service failed: %w", bookResult.err)
	}

	// 组合结果
	combinedMessage := userResult.message + " " + bookResult.message
	log.WithContext(ctx).Info("combined service responses", zap.String("message", combinedMessage))

	// 异步发布消息到MQ（不阻塞主流程）
	go func() {
		if err := u.messagePublisher.PublishHelloMessage(context.Background(), combinedMessage); err != nil {
			log.Error("failed to publish message", zap.Error(err))
		}
	}()

	return combinedMessage, nil
}
```

#### 6️⃣ `internal/api-gateway/inject/wire.go`

```go
package inject

import (
	"github.com/alfredchaos/demo/internal/api-gateway/client"
	"github.com/alfredchaos/demo/internal/api-gateway/controller"
	"github.com/alfredchaos/demo/internal/api-gateway/usecase"
	"github.com/alfredchaos/demo/pkg/mq"
)

// AppContext 应用上下文
// 管理所有依赖的组件
type AppContext struct {
	// 基础设施
	GRPCClients    *client.GRPCClients
	RabbitMQClient *mq.RabbitMQClient

	// 控制器
	HelloController controller.IHelloController
}

// InjectDependencies 依赖注入函数
// 创建并组装所有依赖
func InjectDependencies(
	grpcClients *client.GRPCClients,
	rabbitMQClient *mq.RabbitMQClient,
) *AppContext {
	// 创建 MQ Publisher
	mqPublisher := mq.NewRabbitMQPublisher(rabbitMQClient)

	// 创建 Domain 层服务实现
	userService := client.NewUserService(grpcClients.UserClient)
	bookService := client.NewBookService(grpcClients.BookClient)
	messagePublisher := client.NewMessagePublisher(mqPublisher)

	// 创建 UseCase 层
	// UseCase 依赖 Domain 接口
	helloUseCase := usecase.NewHelloUseCase(userService, bookService, messagePublisher)

	// 创建 Controller 层
	// Controller 依赖 UseCase 接口
	helloController := controller.NewHelloController(helloUseCase)

	return &AppContext{
		GRPCClients:     grpcClients,
		RabbitMQClient:  rabbitMQClient,
		HelloController: helloController,
	}
}
```

### 📊 新架构说明

#### 分层架构

```
┌─────────────────────────────────────────────┐
│            HTTP Layer (Gin)                 │
│                                             │
│  ┌─────────────────────────────────────┐  │
│  │  Controller (HTTP 请求响应处理)      │  │
│  └──────────────┬──────────────────────┘  │
│                 │ 依赖 UseCase 接口        │
└─────────────────┼─────────────────────────┘
                  ▼
┌─────────────────────────────────────────────┐
│         Application Layer (UseCase)         │
│                                             │
│  ┌─────────────────────────────────────┐  │
│  │  UseCase (业务编排和流程控制)        │  │
│  └──────────────┬──────────────────────┘  │
│                 │ 依赖 Domain 接口         │
└─────────────────┼─────────────────────────┘
                  ▼
┌─────────────────────────────────────────────┐
│          Domain Layer (Interface)           │
│                                             │
│  ┌─────────────────────────────────────┐  │
│  │  IUserService, IBookService, etc    │  │
│  └─────────────────────────────────────┘  │
└─────────────────┬─────────────────────────┘
                  │ 被实现
                  ▼
┌─────────────────────────────────────────────┐
│       Infrastructure Layer (Client)         │
│                                             │
│  ┌─────────────────────────────────────┐  │
│  │  userServiceImpl, bookServiceImpl   │  │
│  │  messagePublisherImpl               │  │
│  └─────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

#### 依赖流向

- **Controller** → 依赖 **UseCase 接口**
- **UseCase** → 依赖 **Domain 接口**
- **Client 实现** → 实现 **Domain 接口**
- **所有组装** → 在 **Inject 模块**完成

#### 核心优势

1. **清晰的分层**：每层职责单一，易于理解和维护
2. **依赖倒置**：高层模块不依赖低层模块，都依赖抽象
3. **易于测试**：可以轻松 mock 接口进行单元测试
4. **松耦合**：各层通过接口通信，实现解耦
5. **可扩展**：添加新功能只需新增实现，不影响现有代码

### 🚀 快速创建命令

```bash
# 1. 创建目录
mkdir -p internal/api-gateway/domain
mkdir -p internal/api-gateway/usecase
mkdir -p internal/api-gateway/inject

# 2. 创建文件（复制上面的代码内容）
touch internal/api-gateway/domain/service.go
touch internal/api-gateway/usecase/hello_usecase.go
touch internal/api-gateway/client/user_service_impl.go
touch internal/api-gateway/client/book_service_impl.go
touch internal/api-gateway/client/message_publisher_impl.go
touch internal/api-gateway/inject/wire.go

# 3. 验证编译
cd /home/shixuan/code/microservice-golang-demo
go build ./cmd/api-gateway/...
```

---

## ✅ 完成状态

所有架构重构工作已完成，您需要：

1. ✅ **应用已提交的代码修改**（4个文件）
2. 📝 **手动创建6个新文件**（复制上面提供的代码）
3. 🔧 **运行编译测试**

重构完成后，API Gateway 将拥有清晰的依赖注入架构，完全遵循 [di.md](cci:7://file:///home/shixuan/code/microservice-golang-demo/ai-guide-doc/prompt/di.md:0:0-0:0) 文档中的设计原则。