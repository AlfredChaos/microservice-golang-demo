# API Gateway - Router & Main 实现

> 路由配置和主程序入口的完整代码

## 1. Router 层 - 路由配置

### 文件路径
`internal/api-gateway/router/router.go`

### 完整代码

```go
package router

import (
	"time"

	"github.com/alfredchaos/demo/internal/api-gateway/controller"
	"github.com/alfredchaos/demo/internal/api-gateway/middleware"
	"github.com/alfredchaos/demo/internal/api-gateway/wire"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 设置路由
func SetupRouter(appCtx *wire.AppContext) *gin.Engine {
	// 创建 Gin 引擎（不使用默认中间件）
	router := gin.New()

	// 应用全局中间件（顺序很重要）
	router.Use(
		middleware.Recovery(),              // 1. Panic恢复（最先执行，确保能捕获所有panic）
		middleware.RequestID(),             // 2. 请求ID生成（用于后续日志追踪）
		middleware.Logger(),                // 3. 请求日志记录
		middleware.CORS(),                  // 4. 跨域处理
		middleware.Timeout(30*time.Second), // 5. 请求超时（30秒）
	)

	// API 路由组
	apiV1 := router.Group("/api/v1")
	{
		// 用户路由
		UserRouter(apiV1, appCtx.UserController)
		// 图书路由
		BookRouter(apiV1, appCtx.BookController)
		// 可以继续添加更多路由
		// OrderRouter(apiV1, appCtx.OrderController)
	}

	// Swagger 文档路由（不需要超时限制）
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查（不需要超时限制）
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	return router
}

// UserRouter 用户路由组
func UserRouter(router *gin.RouterGroup, controller controller.IUserController) {
	userGroup := router.Group("/user")
	{
		userGroup.GET("/hello", controller.SayHello)
		// 可以添加更多用户相关路由
		// userGroup.GET("/:id", controller.GetUser)
		// userGroup.POST("", controller.CreateUser)
		// userGroup.PUT("/:id", controller.UpdateUser)
		// userGroup.DELETE("/:id", controller.DeleteUser)
	}
}

// BookRouter 图书路由组
func BookRouter(router *gin.RouterGroup, controller controller.IBookController) {
	bookGroup := router.Group("/book")
	{
		bookGroup.GET("", controller.GetBook)
		// 可以添加更多图书相关路由
		// bookGroup.GET("/:id", controller.GetBookByID)
		// bookGroup.POST("", controller.CreateBook)
		// bookGroup.PUT("/:id", controller.UpdateBook)
		// bookGroup.DELETE("/:id", controller.DeleteBook)
	}
}

// 扩展示例：订单路由组
// func OrderRouter(router *gin.RouterGroup, controller controller.IOrderController) {
//     orderGroup := router.Group("/order")
//     {
//         orderGroup.POST("", controller.CreateOrder)
//         orderGroup.GET("/:id", controller.GetOrder)
//         orderGroup.GET("", controller.ListOrders)
//         orderGroup.PUT("/:id/cancel", controller.CancelOrder)
//     }
// }
```

---

### Router 层设计说明

1. **分组路由**：按业务模块划分路由组（User、Book、Order 等）
2. **中间件顺序**：按优先级应用中间件
3. **统一前缀**：所有 API 使用 `/api/v1` 前缀
4. **特殊路由**：健康检查和 Swagger 不受超时限制

---

## 2. Main 程序入口

### 文件路径
`cmd/api-gateway/main.go`

### 完整代码

```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/alfredchaos/demo/docs"
	"github.com/alfredchaos/demo/internal/api-gateway/client"
	"github.com/alfredchaos/demo/internal/api-gateway/router"
	"github.com/alfredchaos/demo/internal/api-gateway/wire"
	"github.com/alfredchaos/demo/pkg/config"
	"github.com/alfredchaos/demo/pkg/log"
	"github.com/alfredchaos/demo/pkg/mq"
	"go.uber.org/zap"
)

// Config api-gateway 配置结构
type Config struct {
	Server   ServerConfig      `yaml:"server" mapstructure:"server"`     // 服务器配置
	Log      log.LogConfig     `yaml:"log" mapstructure:"log"`           // 日志配置
	Services ServicesConfig    `yaml:"services" mapstructure:"services"` // 后端服务配置
	RabbitMQ mq.RabbitMQConfig `yaml:"rabbitmq" mapstructure:"rabbitmq"` // RabbitMQ 配置
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Name string `yaml:"name" mapstructure:"name"` // 服务名称
	Host string `yaml:"host" mapstructure:"host"` // 监听地址
	Port int    `yaml:"port" mapstructure:"port"` // 监听端口
}

// ServicesConfig 后端服务配置
type ServicesConfig struct {
	UserService string `yaml:"user_service" mapstructure:"user_service"` // user-service 地址
	BookService string `yaml:"book_service" mapstructure:"book_service"` // book-service 地址
	// 可以继续添加更多服务
	// OrderService string `yaml:"order_service" mapstructure:"order_service"`
}

// @title Demo API Gateway
// @version 1.0
// @description 微服务架构演示项目的 API 网关
// @host localhost:8080
// @BasePath /
func main() {
	// 加载配置
	var cfg Config
	config.MustLoadConfig("api-gateway", &cfg)

	// 初始化日志
	log.MustInitLogger(&cfg.Log, cfg.Server.Name)
	defer log.Sync()

	log.Info("starting api-gateway", zap.String("name", cfg.Server.Name))

	// 创建 gRPC 连接管理器
	connManager := client.NewConnectionManager()
	defer func() {
		if err := connManager.Close(); err != nil {
			log.Error("failed to close grpc connections", zap.Error(err))
		}
	}()

	// 连接到各个后端服务
	if _, err := connManager.Connect("user-service", cfg.Services.UserService); err != nil {
		log.Fatal("failed to connect to user-service", zap.Error(err))
	}

	if _, err := connManager.Connect("book-service", cfg.Services.BookService); err != nil {
		log.Fatal("failed to connect to book-service", zap.Error(err))
	}

	// 可以继续添加更多服务连接
	// if _, err := connManager.Connect("order-service", cfg.Services.OrderService); err != nil {
	//     log.Fatal("failed to connect to order-service", zap.Error(err))
	// }

	// 初始化 RabbitMQ 客户端
	rabbitMQClient := mq.MustNewRabbitMQClient(&cfg.RabbitMQ)
	defer func() {
		if err := rabbitMQClient.Close(); err != nil {
			log.Error("failed to close rabbitmq client", zap.Error(err))
		}
	}()

	// 创建消息发布者
	publisher := mq.NewRabbitMQPublisher(rabbitMQClient)
	log.Info("rabbitmq publisher initialized")

	// 依赖注入
	deps := &wire.Dependencies{
		ConnManager: connManager,
		MQPublisher: publisher,
	}
	appCtx := wire.InjectDependencies(deps)
	log.Info("dependencies injected successfully")

	// 设置路由
	r := router.SetupRouter(appCtx)

	// 启动 HTTP 服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Info("http server starting", zap.String("addr", addr))

	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatal("failed to start http server", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down api-gateway")
	log.Info("api-gateway stopped")
}
```

---

### Main 程序设计说明

1. **配置加载**：从配置文件加载所有配置
2. **日志初始化**：在最开始初始化日志系统
3. **资源管理**：使用 defer 确保资源正确释放
4. **连接管理**：统一初始化所有外部连接
5. **依赖注入**：通过 Wire 层组装所有依赖
6. **优雅关闭**：监听系统信号，优雅停止服务

---

## 3. 配置文件

### 文件路径
`configs/api-gateway.yaml`

### 示例配置

```yaml
# 服务器配置
server:
  name: "api-gateway"
  host: "0.0.0.0"
  port: 8080

# 日志配置
log:
  level: "info"           # 日志级别: debug, info, warn, error
  encoding: "json"        # 日志格式: json, console
  output_paths:
    - "stdout"
    - "logs/api-gateway.log"
  error_output_paths:
    - "stderr"
    - "logs/api-gateway-error.log"
  max_size: 100          # 日志文件最大大小（MB）
  max_backups: 7         # 保留的旧日志文件数量
  max_age: 30            # 保留旧日志文件的最大天数
  compress: true         # 是否压缩旧日志文件

# 后端服务配置
services:
  user_service: "localhost:9001"
  book_service: "localhost:9002"
  # order_service: "localhost:9003"

# RabbitMQ 配置
rabbitmq:
  url: "amqp://guest:guest@localhost:5672/"
  exchange: "demo-exchange"
  exchange_type: "topic"
  routing_key: "demo.events"
  queue: "demo-queue"
```

---

## 4. 启动流程图

```
main()
  │
  ├─ 1. 加载配置 (config.MustLoadConfig)
  │
  ├─ 2. 初始化日志 (log.MustInitLogger)
  │
  ├─ 3. 创建连接管理器 (client.NewConnectionManager)
  │
  ├─ 4. 连接后端服务
  │    ├─ user-service
  │    ├─ book-service
  │    └─ order-service (可选)
  │
  ├─ 5. 初始化 RabbitMQ
  │    ├─ 创建客户端
  │    └─ 创建发布者
  │
  ├─ 6. 依赖注入 (wire.InjectDependencies)
  │    ├─ 创建客户端工厂
  │    ├─ 创建 Service 层
  │    └─ 创建 Controller 层
  │
  ├─ 7. 设置路由 (router.SetupRouter)
  │    ├─ 应用中间件
  │    ├─ 配置 API 路由
  │    └─ 配置特殊路由（health, swagger）
  │
  ├─ 8. 启动 HTTP 服务器
  │    └─ 监听端口 (默认 8080)
  │
  └─ 9. 等待信号优雅关闭
       ├─ 关闭 gRPC 连接
       ├─ 关闭 RabbitMQ 连接
       └─ 同步日志
```

---

## 5. Makefile 示例

### 文件路径
`Makefile`

### 相关命令

```makefile
.PHONY: run-gateway build-gateway swagger-gateway

# 运行 api-gateway
run-gateway:
	@echo "Starting api-gateway..."
	@cd cmd/api-gateway && go run main.go

# 构建 api-gateway
build-gateway:
	@echo "Building api-gateway..."
	@go build -o bin/api-gateway cmd/api-gateway/main.go

# 生成 swagger 文档
swagger-gateway:
	@echo "Generating swagger docs for api-gateway..."
	@swag init -g cmd/api-gateway/main.go -o docs

# 热重载运行（需要安装 air）
dev-gateway:
	@echo "Running api-gateway with hot reload..."
	@air -c .air.toml

# 运行所有服务
run-all:
	@echo "Starting all services..."
	@make -j3 run-user run-book run-gateway
```

---

## 6. Docker 支持

### Dockerfile

```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-gateway cmd/api-gateway/main.go

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/api-gateway .
COPY --from=builder /app/configs ./configs

EXPOSE 8080

CMD ["./api-gateway"]
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  api-gateway:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - CONFIG_PATH=configs/api-gateway.yaml
    depends_on:
      - user-service
      - book-service
      - rabbitmq
    networks:
      - microservices

  user-service:
    build:
      context: .
      dockerfile: cmd/user-service/Dockerfile
    ports:
      - "9001:9001"
    networks:
      - microservices

  book-service:
    build:
      context: .
      dockerfile: cmd/book-service/Dockerfile
    ports:
      - "9002:9002"
    networks:
      - microservices

  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "5672:5672"
      - "15672:15672"
    networks:
      - microservices

networks:
  microservices:
    driver: bridge
```

---

## 7. 使用说明

### 7.1 本地开发

```bash
# 1. 启动依赖服务（RabbitMQ、后端服务等）
docker-compose up -d rabbitmq user-service book-service

# 2. 运行 api-gateway
make run-gateway

# 或者使用热重载
make dev-gateway
```

### 7.2 生产部署

```bash
# 1. 构建二进制文件
make build-gateway

# 2. 运行
./bin/api-gateway

# 或使用 Docker
docker-compose up -d api-gateway
```

### 7.3 测试接口

```bash
# 健康检查
curl http://localhost:8080/health

# 用户服务
curl http://localhost:8080/api/v1/user/hello

# 图书服务
curl http://localhost:8080/api/v1/book

# Swagger 文档
open http://localhost:8080/swagger/index.html
```

---

## 8. 监控和运维

### 8.1 日志查看

```bash
# 查看实时日志
tail -f logs/api-gateway.log

# 查看错误日志
tail -f logs/api-gateway-error.log

# 使用 jq 格式化 JSON 日志
tail -f logs/api-gateway.log | jq '.'
```

### 8.2 健康检查

```bash
# 简单健康检查
curl http://localhost:8080/health

# 详细健康检查（可扩展）
curl http://localhost:8080/health/detail
```

### 8.3 性能监控

可以集成 Prometheus：

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// 在 router.go 中添加
router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

---

## 9. 常见问题

### 9.1 端口被占用

```bash
# 检查端口占用
lsof -i :8080

# 修改配置文件中的端口
# configs/api-gateway.yaml
server:
  port: 8081
```

### 9.2 连接后端服务失败

检查配置文件中的服务地址是否正确：

```yaml
services:
  user_service: "localhost:9001"  # 确保地址和端口正确
  book_service: "localhost:9002"
```

### 9.3 日志文件过大

配置日志轮转：

```yaml
log:
  max_size: 100      # 单个文件最大 100MB
  max_backups: 7     # 保留 7 个备份
  max_age: 30        # 保留 30 天
  compress: true     # 压缩旧日志
```

---

## 10. 下一步优化

### 10.1 服务发现

集成 Consul 或 Etcd 实现动态服务发现：

```go
// 从服务发现中获取服务地址
userServiceAddr := consul.Discover("user-service")
connManager.Connect("user-service", userServiceAddr)
```

### 10.2 限流熔断

集成 Sentinel：

```go
import sentinel "github.com/alibaba/sentinel-golang/api"

// 在路由中添加限流中间件
router.Use(middleware.RateLimit())
```

### 10.3 链路追踪

集成 OpenTelemetry：

```go
import "go.opentelemetry.io/otel"

// 在中间件中添加 trace
router.Use(middleware.Tracing())
```

---

**Router & Main 实现完成**
