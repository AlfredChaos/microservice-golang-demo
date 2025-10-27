# Demo 微服务项目

这是一个基于 Golang 的微服务架构演示项目,展示了服务间的同步(gRPC)和异步(RabbitMQ)通信。

## 项目架构

```
backend/
├── .gitignore
├── go.mod
├── go.sum
├── Makefile  # 增加 'make swagger' 命令
├── README.md
|
├── api/                    # 存放所有服务的 API 定义 (Protobuf)，定义服务间同步通信契约
│   ├── user/v1/
│   └── order/v1/
|
├── build/                  # 编译生成的二进制文件
|
├── cmd/                    # 所有服务的启动入口 (main.go)
│   ├── api-gateway/
│   │   └── main.go     # 导入 swagger docs, 初始化 swagger UI 路由
│   └── user-service/
│       └── main.go
|
├── configs/                # 集中管理所有服务的配置文件
│   ├── api-gateway.yaml
│   └── user-service.yaml   # 增加 mongodb 和 redis 的配置
|
├── docs/                   # 存放自动生成的 API 文档文件
│   └── swagger.json
|
├── deployments/            # 部署相关文件
|
├── internal/               # 各服务的内部实现
│   ├── api-gateway/        # API 网关 (BFF)
│   │   ├── client/
│   │   ├── controller/     # 在此处为每个 HTTP Handler 添加 swagger 注释
│   │   ├── dto/            # 在此处为 DTO 结构体添加 swagger 注释
│   │   ├── middleware/ 
│   │   └── router/         # 在此处添加 /swagger/* 的路由
│   │
│   └── user-service/       # 用户服务的内部代码
│       ├── biz/            # 业务逻辑层 (Business)
│       ├── consumer/       # 消息队列消费者
│       ├── data/           # 数据访问层 (Data Access)
│       │   ├── data.go             # 负责初始化所有数据连接和 Repository
│       │   ├── user_repo.go        # 定义用户数据仓库的接口 (Interface)
│       │   ├── user_mongo_repo.go  # 用户仓库的 MongoDB 实现
│       │   └── user_redis_cache.go # 用户缓存的 Redis 实现
│       ├── domain/         # 领域模型 (Domain Model)
│       ├── server/         # gRPC 服务器实现
│       ├── service/        # gRPC Service 实现 (胶水层)
│       └── conf/           # 服务自身的配置加载与结构体
|
├── migrations/             # 数据库迁移脚本
|
├── pkg/                    # 跨服务共享的公共库代码
│   ├── config/             # 共享的配置加载逻辑
│   ├── log/                # 共享的日志初始化逻辑
│   ├── errors/             # 共享的自定义错误类型
│   ├── db/                 # 共享的数据库连接
│   │   └── mongo.go        # 共享的 MongoDB 连接和客户端管理
│   ├── cache/              # 共享的缓存库
│   │   └── redis.go        # 共享的 Redis 连接和客户端管理
│   ├── middleware/         # 共享的 gRPC Interceptors
│   ├── transport/          # 共享的传输层工具
│   ├── discovery/          # 共享的服务发现与注册逻辑
│   └── mq/                 # 共享的消息队列(Message Queue)工具包
│       ├── publisher.go
│       ├── consumer.go
│       └── rabbitmq.go
|
├── scripts/                # 通用脚本
│   └── gen-proto.sh
│   └── gen-swagger.sh      # 用于扫描代码并生成 swagger.json 的脚本
|
└── third_party/            # 第三方工具或代码
```

## 业务流程

1. 客户端发送 POST 请求到 `api-gateway`
2. `api-gateway` 并发调用 `user-service` 和 `book-service` 的 gRPC 接口
3. `user-service` 返回 "Hello", `book-service` 返回 "World"
4. `api-gateway` 组合响应为 "Hello World" 并返回给客户端
5. 同时 `api-gateway` 发送消息到 RabbitMQ
6. `nice-service` 消费消息并打印 "Nice"

## 技术栈

- **框架**: Gin (HTTP), gRPC
- **数据库**: MongoDB
- **缓存**: Redis
- **消息队列**: RabbitMQ
- **配置管理**: Viper
- **日志**: Zap
- **API文档**: Swagger

## 快速开始

### 前置要求

- Go 1.21+
- MongoDB
- Redis
- RabbitMQ
- protoc (Protocol Buffers 编译器)

### 安装依赖

```bash
go mod download
```

### 生成 Protobuf 代码

```bash
make proto
```

### 生成 Swagger 文档

```bash
make swagger
```

### 编译所有服务

```bash
make build
```

### 运行服务

```bash
# 启动 user-service
./build/user-service

# 启动 book-service
./build/book-service

# 启动 nice-service
./build/nice-service

# 启动 api-gateway
./build/api-gateway
```

### 测试接口

```bash
curl -X POST http://localhost:8080/api/v1/hello \
  -H "Content-Type: application/json" \
  -d '{}'
```

预期响应:
```json
{
  "code": 0,
  "message": "success",
  "data": "Hello World"
}
```

### 查看 Swagger 文档

访问: http://localhost:8080/swagger/index.html

## 项目结构说明

- `api/`: Protobuf API 定义
- `cmd/`: 各服务的启动入口
- `configs/`: 配置文件
- `docs/`: Swagger 文档
- `internal/`: 各服务的内部实现
- `pkg/`: 跨服务共享的公共库
- `scripts/`: 工具脚本

## Makefile 命令

```bash
make proto      # 生成 protobuf 代码
make swagger    # 生成 swagger 文档
make build      # 编译所有服务
make clean      # 清理编译产物
```

## 配置说明

各服务的配置文件位于 `configs/` 目录:

- `api-gateway.yaml`: API 网关配置
- `user-service.yaml`: 用户服务配置
- `book-service.yaml`: 图书服务配置
- `nice-service.yaml`: 消息消费服务配置

## 开发规范

- 遵循 Go 语言最佳实践
- 使用依赖注入解耦组件
- 遵循单一职责原则
- 保持高内聚低耦合
- 为公开接口提供清晰注释

## License

MIT
