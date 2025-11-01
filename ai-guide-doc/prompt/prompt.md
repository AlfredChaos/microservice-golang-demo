### 项目主旨
本项目的目的是构建一个Golang后端微服务架构，无须关注具体业务的实现，只需跑通项目结构即可。
因此：**请你遵循我所提供的项目结构和要求，为我实现项目demo**

### 项目结构
以下是提供参考的目录结构
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

## 项目要求
1. 根据项目结构，实现Golang后端微服务架构
2. 实现api-gateway作为网关，接受来自客户端的http请求，并转发给后端服务（客户端无须实现）。请你构造一个POST请求，返回值为“Hello World”
3. 实现一个user-service微服务，通过grpc和api-gateway连接，实现一个grpc同步接口，返回值为“Hello”
4. 实现一个book-service微服务，通过grpc和api-gateway连接，实现一个grpc同步接口，返回值为“World”
5. 实现一个nice-service微服务，作为Rabbitmq消费者接受来自api-gateway的消息，收到消息后，打印"Nice"，此服务代表异步通信
6. 业务逻辑：
    - api-gateway接收到客户端的http请求后，将请求转发给user-service和book-service
    - user-service和book-service接收到请求后，分别返回"Hello"和"World"
    - api-gateway接收到user-service和book-service的响应后，发送一个hello消息到消息队列，通知nice-service服务，并组合响应返回给客户端
    - nice-service接收到请求后，打印"Nice"


