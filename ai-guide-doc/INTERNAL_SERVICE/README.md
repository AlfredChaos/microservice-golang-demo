# 内部服务架构文档

## 概述

本文档描述了微服务项目中内部服务的统一架构设计。所有内部服务（如 `user-service`、`book-service`）都遵循相同的架构模式，以降低学习成本，让开发者专注于业务逻辑的实现。

## 核心特性

### 1. 统一架构模式
- 采用分层架构（Domain、Data、Biz、Service、Server）
- 使用依赖注入（手动实现，不使用wire）
- 遵循SOLID原则和清晰的依赖关系

### 2. 双向通信能力
- **gRPC服务端**：接收来自api-gateway的请求（南北向）
- **gRPC客户端**：调用其他内部服务（东西向）
  - 使用统一的 `pkg/grpcclient` 模块管理客户端连接
  - 配置驱动，支持重试、超时、拦截器等功能
  - 自动处理连接生命周期

### 3. 多数据源支持（可选配置）
- **PostgreSQL**：关系型数据库，使用GORM + Goose迁移
- **MongoDB**：文档数据库，存储非结构化数据
- **Redis**：缓存和会话管理

### 4. 消息队列集成
- **RabbitMQ**：支持作为生产者和消费者
- 用于异步任务和事件驱动架构

### 5. gRPC接口共享
- Proto文件统一管理在 `api/` 目录
- 避免接口变更时所有服务都需要更新

## 文档索引

| 文档 | 说明 |
|------|------|
| [ARCHITECTURE.md](./ARCHITECTURE.md) | 详细的分层架构设计 |
| [DI_AND_WIRE.md](./DI_AND_WIRE.md) | 依赖注入实现指南 |
| [DATA_STORAGE.md](./DATA_STORAGE.md) | 数据存储层设计 |
| [SERVICE_COMMUNICATION.md](./SERVICE_COMMUNICATION.md) | 服务间通信机制 |
| [DEVELOPMENT_GUIDE.md](./DEVELOPMENT_GUIDE.md) | 开发新服务指南 |

## 快速开始

### 创建新服务的步骤

1. 参考 [DEVELOPMENT_GUIDE.md](./DEVELOPMENT_GUIDE.md) 了解完整流程
2. 使用现有服务（如 `user-service`）作为模板
3. 实现业务逻辑在 `biz` 层
4. 定义数据访问在 `data` 层
5. 实现gRPC服务在 `service` 层
6. 配置依赖注入在 `main.go`

## 设计原则

### 1. 依赖倒置原则
- 高层模块不依赖低层模块，都依赖抽象
- 业务逻辑依赖接口，而非具体实现

### 2. 单一职责原则
- 每一层有明确的职责
- 代码更易维护和测试

### 3. 开闭原则
- 对扩展开放，对修改关闭
- 通过接口和依赖注入实现

### 4. 接口隔离原则
- 定义最小化的接口
- 避免依赖不需要的方法

## 示例服务结构

```
internal/user-service/
├── domain/          # 领域模型层
│   ├── user.go      # 实体定义
│   └── errors.go    # 领域错误
├── data/            # 数据访问层
│   ├── data.go      # 数据层初始化
│   ├── user_repo.go # 仓库接口
│   └── user_pg_repo.go  # PostgreSQL实现
├── biz/             # 业务逻辑层
│   └── user_usecase.go  # 用例实现
├── service/         # gRPC服务层
│   └── user_service.go  # gRPC服务实现
├── server/          # 服务器层
│   ├── grpc.go      # gRPC服务器
│   └── client.go    # gRPC客户端（可选）
└── conf/            # 配置层
    └── config.go    # 配置结构

cmd/user-service/
└── main.go          # 启动入口，依赖注入

api/user/v1/         # gRPC接口定义（共享）
└── user.proto       # Protobuf定义
```

## 下一步

- 阅读 [ARCHITECTURE.md](./ARCHITECTURE.md) 了解详细的架构设计
- 阅读 [DEVELOPMENT_GUIDE.md](./DEVELOPMENT_GUIDE.md) 开始开发新服务
