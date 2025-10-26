# 部署指南

## 前置要求

### 1. 安装 Go 1.21+
```bash
go version
```

### 2. 安装 Protocol Buffers 编译器
```bash
# macOS
brew install protobuf

# 验证安装
protoc --version
```

### 3. 安装 Go 工具
```bash
# 安装 protoc-gen-go 和 protoc-gen-go-grpc
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 安装 swag (Swagger 文档生成工具)
go install github.com/swaggo/swag/cmd/swag@latest

# 确保 $GOPATH/bin 在 PATH 中
export PATH=$PATH:$(go env GOPATH)/bin
```

### 4. 安装并启动依赖服务

#### MongoDB
```bash
# macOS
brew install mongodb-community
brew services start mongodb-community

# 验证
mongosh
```

#### Redis
```bash
# macOS
brew install redis
brew services start redis

# 验证
redis-cli ping
```

#### RabbitMQ
```bash
# macOS
brew install rabbitmq
brew services start rabbitmq

# 访问管理界面: http://localhost:15672
# 默认用户名/密码: guest/guest
```

## 构建步骤

### 1. 下载依赖
```bash
cd /Users/alfredchaos/home/company/demo
go mod download
go mod tidy
```

### 2. 生成 Protobuf 代码
```bash
make proto
```

### 3. 生成 Swagger 文档
```bash
make swagger
```

### 4. 编译所有服务
```bash
make build
```

## 运行服务

### 方式一: 使用 Makefile (推荐在不同终端运行)

```bash
# 终端 1: 启动 user-service
make run-user

# 终端 2: 启动 book-service
make run-book

# 终端 3: 启动 nice-service
make run-nice

# 终端 4: 启动 api-gateway
make run-gateway
```

### 方式二: 直接运行编译后的二进制文件

```bash
# 终端 1
./build/user-service

# 终端 2
./build/book-service

# 终端 3
./build/nice-service

# 终端 4
./build/api-gateway
```

### 方式三: 使用 go run

```bash
# 终端 1
go run cmd/user-service/main.go

# 终端 2
go run cmd/book-service/main.go

# 终端 3
go run cmd/nice-service/main.go

# 终端 4
go run cmd/api-gateway/main.go
```

## 测试接口

### 1. 健康检查
```bash
curl http://localhost:8080/health
```

### 2. 调用 Hello 接口
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

### 3. 查看 Swagger 文档
打开浏览器访问: http://localhost:8080/swagger/index.html

### 4. 验证消息队列
查看 nice-service 的日志输出,应该能看到 "Nice" 的打印信息。

## 服务端口

- **api-gateway**: 8080 (HTTP)
- **user-service**: 9001 (gRPC)
- **book-service**: 9002 (gRPC)
- **nice-service**: 无端口 (RabbitMQ Consumer)

## 故障排查

### 问题 1: protoc 命令找不到
```bash
# 确保 protoc 已安装
which protoc

# 如果没有,重新安装
brew install protobuf
```

### 问题 2: protoc-gen-go 找不到
```bash
# 确保 Go bin 目录在 PATH 中
export PATH=$PATH:$(go env GOPATH)/bin

# 重新安装工具
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 问题 3: MongoDB 连接失败
```bash
# 检查 MongoDB 是否运行
brew services list | grep mongodb

# 启动 MongoDB
brew services start mongodb-community

# 检查连接
mongosh
```

### 问题 4: RabbitMQ 连接失败
```bash
# 检查 RabbitMQ 是否运行
brew services list | grep rabbitmq

# 启动 RabbitMQ
brew services start rabbitmq

# 访问管理界面
open http://localhost:15672
```

### 问题 5: 端口被占用
```bash
# 查看端口占用
lsof -i :8080
lsof -i :9001
lsof -i :9002

# 杀死占用端口的进程
kill -9 <PID>
```

## 停止服务

在每个终端按 `Ctrl+C` 停止对应的服务。

## 清理

```bash
# 清理编译产物
make clean

# 停止所有依赖服务
brew services stop mongodb-community
brew services stop redis
brew services stop rabbitmq
```

## 开发建议

1. **修改代码后**: 重新编译对应的服务
2. **修改 proto 文件后**: 运行 `make proto` 重新生成代码
3. **修改 Swagger 注释后**: 运行 `make swagger` 重新生成文档
4. **查看日志**: 所有服务的日志都输出到控制台,便于调试

## 项目结构说明

- `api/`: Protobuf API 定义
- `cmd/`: 各服务的启动入口
- `configs/`: 配置文件
- `docs/`: Swagger 文档 (自动生成)
- `internal/`: 各服务的内部实现
- `pkg/`: 跨服务共享的公共库
- `scripts/`: 工具脚本
- `build/`: 编译产物 (自动生成)
