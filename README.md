# Demo 微服务项目

这是一个基于 Golang 的微服务架构演示项目,展示了服务间的同步(gRPC)和异步(RabbitMQ)通信。

## 项目架构

```
├── api-gateway      # HTTP 网关服务,接收客户端请求
├── user-service     # 用户服务 (gRPC)
├── book-service     # 图书服务 (gRPC)
└── nice-service     # 消息消费服务 (RabbitMQ Consumer)
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
