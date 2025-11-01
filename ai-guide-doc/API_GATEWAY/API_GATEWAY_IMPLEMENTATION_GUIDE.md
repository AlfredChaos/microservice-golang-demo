# API Gateway 实施指南

> 完整的实施步骤和文件清单

## 文档索引

1. [架构重构方案](./API_GATEWAY_DI_REFACTOR.md) - 整体架构设计和原则
2. [Client 层实现](./API_GATEWAY_CLIENT_LAYER.md) - gRPC 客户端管理
3. [Domain & Service 层实现](./API_GATEWAY_DOMAIN_SERVICE_LAYER.md) - 领域接口和服务实现
4. [Controller & Wire 层实现](./API_GATEWAY_CONTROLLER_WIRE_LAYER.md) - 控制器和依赖注入
5. [Router & Main 实现](./API_GATEWAY_ROUTER_MAIN.md) - 路由配置和主程序

---

## 实施步骤

### 第一阶段：创建新文件

#### 1. Client 层（gRPC 客户端管理）

```bash
# 创建目录
mkdir -p internal/api-gateway/client

# 创建文件
touch internal/api-gateway/client/connection_manager.go
touch internal/api-gateway/client/client_factory.go
```

复制以下内容：
- `connection_manager.go` - 参考 [Client 层实现](./API_GATEWAY_CLIENT_LAYER.md#1-connectionmanager---连接管理器)
- `client_factory.go` - 参考 [Client 层实现](./API_GATEWAY_CLIENT_LAYER.md#2-clientfactory---客户端工厂)

---

#### 2. Domain 层（领域接口）

```bash
# 创建文件
touch internal/api-gateway/domain/user_service.go
touch internal/api-gateway/domain/book_service.go
```

复制以下内容：
- `user_service.go` - 参考 [Domain & Service 层实现](./API_GATEWAY_DOMAIN_SERVICE_LAYER.md#11-用户服务接口)
- `book_service.go` - 参考 [Domain & Service 层实现](./API_GATEWAY_DOMAIN_SERVICE_LAYER.md#12-图书服务接口)

---

#### 3. Service 层（服务实现）

```bash
# 创建目录
mkdir -p internal/api-gateway/service

# 创建文件
touch internal/api-gateway/service/user_service.go
touch internal/api-gateway/service/book_service.go
```

复制以下内容：
- `user_service.go` - 参考 [Domain & Service 层实现](./API_GATEWAY_DOMAIN_SERVICE_LAYER.md#21-用户服务实现)
- `book_service.go` - 参考 [Domain & Service 层实现](./API_GATEWAY_DOMAIN_SERVICE_LAYER.md#22-图书服务实现)

---

#### 4. Controller 层（控制器）

```bash
# 创建文件
touch internal/api-gateway/controller/user_controller.go
touch internal/api-gateway/controller/book_controller.go
```

复制以下内容：
- `user_controller.go` - 参考 [Controller & Wire 层实现](./API_GATEWAY_CONTROLLER_WIRE_LAYER.md#11-用户控制器)
- `book_controller.go` - 参考 [Controller & Wire 层实现](./API_GATEWAY_CONTROLLER_WIRE_LAYER.md#12-图书控制器)

---

#### 5. Wire 层（依赖注入）

```bash
# 创建目录
mkdir -p internal/api-gateway/wire

# 创建文件
touch internal/api-gateway/wire/wire.go
```

复制以下内容：
- `wire.go` - 参考 [Controller & Wire 层实现](./API_GATEWAY_CONTROLLER_WIRE_LAYER.md#2-wire-层---依赖注入)

---

### 第二阶段：修改现有文件

#### 1. 修改 Router

编辑文件：`internal/api-gateway/router/router.go`

参考：[Router & Main 实现](./API_GATEWAY_ROUTER_MAIN.md#1-router-层---路由配置)

主要修改：
- 修改 `SetupRouter` 函数签名，接收 `*wire.AppContext`
- 添加 `UserRouter` 和 `BookRouter` 函数
- 删除或注释旧的路由配置

---

#### 2. 修改 Main

编辑文件：`cmd/api-gateway/main.go`

参考：[Router & Main 实现](./API_GATEWAY_ROUTER_MAIN.md#2-main-程序入口)

主要修改：
- 删除旧的 `client.NewGRPCClients` 调用
- 创建 `ConnectionManager`
- 使用 `wire.InjectDependencies` 进行依赖注入
- 传递 `AppContext` 给 `router.SetupRouter`

---

### 第三阶段：清理旧文件

```bash
# 备份旧文件（可选）
mv internal/api-gateway/client/grpc_client.go internal/api-gateway/client/grpc_client.go.bak
mv internal/api-gateway/controller/hello_controller.go internal/api-gateway/controller/hello_controller.go.bak

# 或直接删除
rm internal/api-gateway/client/grpc_client.go
rm internal/api-gateway/controller/hello_controller.go
```

---

### 第四阶段：验证和测试

#### 1. 编译检查

```bash
# 进入项目根目录
cd /home/shixuan/code/microservice-golang-demo

# 编译检查
go build ./cmd/api-gateway/main.go

# 或使用 Makefile
make build-gateway
```

#### 2. 运行测试

```bash
# 运行单元测试
go test ./internal/api-gateway/...

# 运行集成测试
go test -tags=integration ./internal/api-gateway/...
```

#### 3. 启动服务

```bash
# 启动后端服务
make run-user &
make run-book &

# 启动 api-gateway
make run-gateway
```

#### 4. 测试接口

```bash
# 健康检查
curl http://localhost:8080/health

# 测试用户服务
curl http://localhost:8080/api/v1/user/hello

# 测试图书服务
curl http://localhost:8080/api/v1/book
```

---

## 文件清单

### 新建文件

```
internal/api-gateway/
├── client/
│   ├── connection_manager.go     ✅ 新建
│   └── client_factory.go         ✅ 新建
├── domain/
│   ├── user_service.go            ✅ 新建
│   └── book_service.go            ✅ 新建
├── service/
│   ├── user_service.go            ✅ 新建
│   └── book_service.go            ✅ 新建
├── controller/
│   ├── user_controller.go         ✅ 新建
│   └── book_controller.go         ✅ 新建
└── wire/
    └── wire.go                    ✅ 新建
```

### 修改文件

```
internal/api-gateway/
└── router/
    └── router.go                  📝 修改

cmd/api-gateway/
└── main.go                        📝 修改
```

### 删除/备份文件

```
internal/api-gateway/
├── client/
│   └── grpc_client.go             ❌ 删除（或备份）
└── controller/
    └── hello_controller.go        ❌ 删除（或备份）
```

---

## 目录结构对比

### 重构前

```
internal/api-gateway/
├── client/
│   └── grpc_client.go
├── controller/
│   └── hello_controller.go
├── dto/
│   └── response.go
├── middleware/
│   └── ...
└── router/
    └── router.go
```

### 重构后

```
internal/api-gateway/
├── client/                        # gRPC 客户端管理
│   ├── connection_manager.go     # 连接管理器
│   └── client_factory.go         # 客户端工厂
├── domain/                        # 领域接口
│   ├── user_service.go
│   └── book_service.go
├── service/                       # 服务实现
│   ├── user_service.go
│   └── book_service.go
├── controller/                    # 控制器
│   ├── user_controller.go
│   └── book_controller.go
├── wire/                          # 依赖注入
│   └── wire.go
├── dto/                           # 数据传输对象
│   └── response.go
├── middleware/                    # 中间件
│   └── ...
└── router/                        # 路由配置
    └── router.go
```

---

## 依赖关系图

```
main.go
  │
  ├─ ConnectionManager ────────┐
  ├─ MQPublisher               │
  │                            │
  └─ wire.InjectDependencies   │
        │                      │
        ├─ ClientFactory ◄─────┘
        │     │
        │     ├─ UserClient
        │     └─ BookClient
        │
        ├─ Service Layer
        │     ├─ UserService (实现 IUserService)
        │     └─ BookService (实现 IBookService)
        │
        └─ Controller Layer
              ├─ UserController (依赖 IUserService)
              └─ BookController (依赖 IBookService)
```

---

## 常见问题处理

### 1. 编译错误：找不到包

```bash
# 确保所有依赖已下载
go mod tidy
go mod download
```

### 2. 运行错误：连接后端服务失败

检查配置文件 `configs/api-gateway.yaml`：

```yaml
services:
  user_service: "localhost:9001"  # 确保地址正确
  book_service: "localhost:9002"
```

确保后端服务已启动：

```bash
# 检查端口
lsof -i :9001
lsof -i :9002
```

### 3. 导入路径错误

确保所有导入路径使用正确的模块名：

```go
import (
    "github.com/alfredchaos/demo/internal/api-gateway/domain"
    "github.com/alfredchaos/demo/internal/api-gateway/service"
    // ...
)
```

检查 `go.mod` 中的模块名是否正确。

---

## 测试清单

- [ ] 编译通过
- [ ] 单元测试通过
- [ ] 健康检查接口正常
- [ ] 用户服务接口正常
- [ ] 图书服务接口正常
- [ ] Swagger 文档正常
- [ ] 日志输出正常
- [ ] gRPC 连接正常
- [ ] 优雅关闭正常

---

## 后续优化建议

### 1. 添加单元测试

为每个层编写单元测试：
- Service 层：使用 Mock gRPC 客户端
- Controller 层：使用 Mock Service
- 测试覆盖率目标：80%+

### 2. 添加集成测试

编写端到端测试：
- 启动测试服务器
- 模拟真实请求
- 验证响应正确性

### 3. 性能优化

- 添加连接池
- 实现请求缓存
- 优化日志性能

### 4. 监控和告警

- 集成 Prometheus 指标
- 添加自定义业务指标
- 配置告警规则

### 5. 文档完善

- 生成 API 文档（Swagger）
- 编写运维文档
- 添加架构决策记录（ADR）

---

## 参考文档

- [架构重构方案](./API_GATEWAY_DI_REFACTOR.md)
- [Client 层实现](./API_GATEWAY_CLIENT_LAYER.md)
- [Domain & Service 层实现](./API_GATEWAY_DOMAIN_SERVICE_LAYER.md)
- [Controller & Wire 层实现](./API_GATEWAY_CONTROLLER_WIRE_LAYER.md)
- [Router & Main 实现](./API_GATEWAY_ROUTER_MAIN.md)

---

**准备开始实施吗？按照上述步骤逐步进行，祝您成功！**
