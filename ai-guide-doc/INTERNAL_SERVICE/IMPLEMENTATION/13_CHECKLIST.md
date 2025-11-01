# 内部服务实施检查清单

> 完整的实施验证清单，确保所有必需组件都已正确实现

## 使用说明

在每个阶段完成后，使用此清单验证实施的完整性。标记完成的项目：
- ✅ 已完成
- ⚠️ 部分完成或可选
- ❌ 未完成

---

## 阶段1：项目基础（必需）

### 1.1 gRPC接口定义
- [ ] Proto文件已创建（`api/<service>/v1/<service>.proto`）
- [ ] 定义了所有必需的service和message
- [ ] go_package选项正确配置
- [ ] 运行`./scripts/gen-proto.sh`成功生成Go代码
- [ ] 生成的`.pb.go`和`_grpc.pb.go`文件无错误

### 1.2 目录结构
- [ ] `internal/<service>/domain/`目录已创建
- [ ] `internal/<service>/data/`目录已创建
- [ ] `internal/<service>/biz/`目录已创建
- [ ] `internal/<service>/service/`目录已创建
- [ ] `internal/<service>/server/`目录已创建
- [ ] `internal/<service>/conf/`目录已创建
- [ ] `internal/<service>/migrations/`目录已创建
- [ ] `cmd/<service>/`目录已创建
- [ ] `configs/`目录包含配置文件

### 1.3 配置文件
- [ ] 创建了`configs/<service>.yaml`
- [ ] 配置了server部分（name, host, port）
- [ ] 配置了log部分
- [ ] 配置了database部分（如需要）
- [ ] 配置了redis部分（如需要）
- [ ] 配置了mongodb部分（如需要）
- [ ] 配置了rabbitmq部分（如需要）
- [ ] 配置了grpc_clients部分（如需调用其他服务）

---

## 阶段2：领域层实现（必需）

### 2.1 领域错误
- [ ] 创建了`domain/errors.go`
- [ ] 定义了所有业务相关错误
- [ ] 错误命名清晰，遵循`Err`前缀
- [ ] 错误按类别组织

### 2.2 值对象
- [ ] 定义了值对象（如OrderItem）
- [ ] 实现了Validate方法
- [ ] 实现了业务计算方法
- [ ] 提供了String方法（可选）

### 2.3 实体
- [ ] 定义了状态枚举（如OrderStatus）
- [ ] 状态枚举提供了IsValid方法
- [ ] 状态枚举提供了CanTransitionTo方法（如需要）
- [ ] 定义了主实体（如Order）
- [ ] 实体包含所有必要字段
- [ ] 提供了工厂函数（如NewOrder）
- [ ] 实现了Validate方法
- [ ] 实现了业务方法（如Confirm, Cancel）
- [ ] 实现了查询方法（如IsPending）

### 2.4 单元测试
- [ ] 测试了工厂函数
- [ ] 测试了Validate方法
- [ ] 测试了业务方法
- [ ] 测试了状态转换
- [ ] 测试了边界条件

---

## 阶段3：数据访问层实现（必需）

### 3.1 仓库接口
- [ ] 创建了`data/<entity>_repo.go`
- [ ] 定义了Repository接口
- [ ] 接口包含所有必需的CRUD方法
- [ ] 方法签名正确（接收context.Context）

### 3.2 PostgreSQL实现
- [ ] 创建了`data/<entity>_pg_repo.go`
- [ ] 定义了持久化对象（PO）
- [ ] PO正确映射到数据库表（gorm标签）
- [ ] 实现了ToDomain方法
- [ ] 实现了FromDomain方法
- [ ] 实现了仓库结构体
- [ ] 实现了所有Repository接口方法
- [ ] 正确处理了gorm.ErrRecordNotFound
- [ ] 转换为领域错误

### 3.3 MongoDB实现（可选）
- [ ] 创建了`data/<entity>_mongo_repo.go`
- [ ] 定义了MongoDB持久化对象
- [ ] 实现了Repository接口
- [ ] 创建了必要的索引

### 3.4 Redis缓存（可选）
- [ ] 创建了`data/<entity>_cache.go`
- [ ] 实现了Get/Set/Delete方法
- [ ] 正确设置了TTL
- [ ] 创建了`data/<entity>_cached_repo.go`
- [ ] 实现了缓存装饰器

### 3.5 数据层容器
- [ ] 创建了`data/data.go`
- [ ] 定义了Data结构体
- [ ] Data结构包含所有数据源字段
- [ ] Data结构导出所有Repository
- [ ] 实现了NewData构造函数
- [ ] NewData根据配置创建正确的仓库实现
- [ ] 实现了Close方法
- [ ] Close方法关闭所有连接

---

## 阶段4：业务逻辑层实现（必需）

### 4.1 UseCase接口
- [ ] 创建了`biz/<entity>_usecase.go`
- [ ] 定义了UseCase接口
- [ ] 接口包含所有业务方法
- [ ] 方法签名正确

### 4.2 UseCase实现
- [ ] 定义了useCase结构体
- [ ] 结构体包含必要的依赖（Repository、gRPC客户端等）
- [ ] 实现了构造函数
- [ ] 实现了所有接口方法
- [ ] 业务方法调用领域对象的业务方法
- [ ] 正确处理事务（如需要）
- [ ] 调用其他服务（如需要）
- [ ] 发布消息到队列（如需要）
- [ ] 记录了适当的日志

### 4.3 单元测试
- [ ] 创建了Mock Repository
- [ ] 测试了所有UseCase方法
- [ ] 使用Mock隔离依赖
- [ ] 测试了成功场景
- [ ] 测试了失败场景

---

## 阶段5：服务层实现（必需）

### 5.1 gRPC服务
- [ ] 创建了`service/<entity>_service.go`
- [ ] 定义了Service结构体
- [ ] Service嵌入了UnimplementedXXXServiceServer
- [ ] Service包含UseCase依赖
- [ ] 实现了构造函数
- [ ] 实现了所有gRPC接口方法

### 5.2 协议转换
- [ ] 实现了Proto到Domain的转换
- [ ] 实现了Domain到Proto的转换
- [ ] 创建了辅助转换方法
- [ ] 正确处理了嵌套对象
- [ ] 正确处理了时间戳

### 5.3 错误处理
- [ ] 将领域错误转换为gRPC错误
- [ ] 使用了正确的gRPC状态码
- [ ] 提供了清晰的错误消息
- [ ] 记录了错误日志

---

## 阶段6：服务器层实现（必需）

### 6.1 gRPC服务器
- [ ] 创建了`server/grpc.go`
- [ ] 定义了GRPCServer结构体
- [ ] 实现了NewGRPCServer构造函数
- [ ] 注册了gRPC服务
- [ ] 添加了拦截器（日志、恢复等）
- [ ] 注册了反射服务（开发环境）
- [ ] 实现了Start方法
- [ ] 实现了Stop方法（优雅关闭）

---

## 阶段7：配置实现（必需）

### 7.1 配置结构
- [ ] 创建了`conf/config.go`
- [ ] 定义了Config结构体
- [ ] 包含了ServerConfig
- [ ] 包含了LogConfig
- [ ] 包含了DatabaseConfig（如需要）
- [ ] 包含了RedisConfig（如需要）
- [ ] 包含了MongoConfig（如需要）
- [ ] 包含了RabbitMQConfig（如需要）
- [ ] 包含了GRPCClientsConfig（如需要）
- [ ] 所有字段都有yaml和mapstructure标签

---

## 阶段8：主函数实现（必需）

### 8.1 main.go基础
- [ ] 创建了`cmd/<service>/main.go`
- [ ] 导入了所有必需的包
- [ ] 实现了init函数（如需注册gRPC客户端）
- [ ] 实现了main函数

### 8.2 依赖注入流程
- [ ] 加载配置（config.MustLoadConfig）
- [ ] 初始化日志（log.MustInitLogger）
- [ ] 初始化PostgreSQL（如需要）
- [ ] 初始化MongoDB（如需要）
- [ ] 初始化Redis（如需要）
- [ ] 初始化RabbitMQ（如需要）
- [ ] 初始化gRPC客户端管理器（如需要）
- [ ] 注册并连接gRPC客户端（如需要）
- [ ] 初始化Data层
- [ ] 初始化Biz层
- [ ] 初始化Service层
- [ ] 初始化Server层
- [ ] 启动消费者（如需要）
- [ ] 启动gRPC服务器
- [ ] 实现优雅关闭

### 8.3 资源管理
- [ ] 所有资源都有defer关闭
- [ ] 关闭顺序正确
- [ ] 监听SIGINT和SIGTERM信号
- [ ] 优雅关闭服务器

---

## 阶段9：数据库迁移（必需）

### 9.1 迁移文件
- [ ] 创建了迁移SQL文件
- [ ] 文件命名正确（001_*.sql）
- [ ] 包含了Up迁移
- [ ] 包含了Down迁移
- [ ] 创建了所有必需的表
- [ ] 创建了索引
- [ ] 设置了外键（如需要）

### 9.2 迁移执行
- [ ] 安装了Goose工具
- [ ] 配置了数据库连接字符串
- [ ] 成功执行了迁移
- [ ] 验证了表结构
- [ ] 测试了回滚

---

## 阶段10：gRPC客户端集成（可选）

### 10.1 配置
- [ ] 在Config中添加了GRPCClients字段
- [ ] 在配置文件中添加了grpc_clients配置

### 10.2 客户端注册
- [ ] 在init函数中注册了客户端工厂
- [ ] 注册了所有需要的服务

### 10.3 客户端初始化
- [ ] 创建了clientManager
- [ ] 注册了服务配置
- [ ] 调用了ConnectAll
- [ ] 添加了defer Close
- [ ] 获取了gRPC客户端连接
- [ ] 创建了客户端实例

### 10.4 业务层集成
- [ ] UseCase构造函数接收gRPC客户端
- [ ] 业务方法中调用了其他服务
- [ ] 正确处理了调用错误
- [ ] 添加了nil检查（客户端可选）

---

## 阶段11：RabbitMQ集成（可选）

### 11.1 发布者
- [ ] 初始化了RabbitMQ客户端
- [ ] 创建了Publisher
- [ ] 在UseCase中注入Publisher
- [ ] 发布消息时使用goroutine（异步）
- [ ] 正确序列化消息（JSON）
- [ ] 记录发布失败日志

### 11.2 消费者
- [ ] 创建了`consumer/<entity>_consumer.go`
- [ ] 定义了Consumer结构体
- [ ] 实现了Start方法
- [ ] 订阅了正确的队列
- [ ] 实现了消息处理逻辑
- [ ] 正确反序列化消息
- [ ] 调用了业务逻辑
- [ ] 确认了消息
- [ ] 在main函数中启动了Consumer

---

## 阶段12：测试（必需）

### 12.1 单元测试
- [ ] Domain层有单元测试
- [ ] Biz层有单元测试
- [ ] 使用了Mock隔离依赖
- [ ] 测试覆盖率 > 70%

### 12.2 集成测试（可选）
- [ ] 创建了集成测试文件
- [ ] 使用测试容器（testcontainers）
- [ ] 测试了完整的业务流程
- [ ] 测试了gRPC接口

### 12.3 手动测试
- [ ] 使用grpcurl测试了所有接口
- [ ] 验证了创建功能
- [ ] 验证了查询功能
- [ ] 验证了更新功能
- [ ] 验证了删除功能
- [ ] 验证了错误处理

---

## 阶段13：文档和代码质量

### 13.1 代码注释
- [ ] 所有公共函数都有注释
- [ ] 注释遵循Go Doc规范
- [ ] 注释使用中文（根据项目要求）
- [ ] 复杂逻辑有详细注释

### 13.2 代码风格
- [ ] 运行了`gofmt`
- [ ] 运行了`goimports`
- [ ] 运行了`golint`（无警告）
- [ ] 变量命名清晰
- [ ] 函数职责单一

### 13.3 错误处理
- [ ] 所有错误都被正确处理
- [ ] 错误被包装（使用fmt.Errorf %w）
- [ ] 错误日志包含足够的上下文
- [ ] 没有忽略错误（no `_ = err`）

### 13.4 日志记录
- [ ] 关键操作有日志
- [ ] 日志级别正确（Info/Warn/Error）
- [ ] 使用了WithContext获取logger
- [ ] 日志包含必要的字段（zap.String等）
- [ ] 敏感信息已脱敏

---

## 阶段14：部署准备

### 14.1 配置文件
- [ ] 创建了生产环境配置
- [ ] 生产配置使用环境变量（敏感信息）
- [ ] 配置了合理的连接池大小
- [ ] 配置了合理的超时时间
- [ ] 日志级别设置为info或warn

### 14.2 Docker化（可选）
- [ ] 创建了Dockerfile
- [ ] 创建了.dockerignore
- [ ] 构建了Docker镜像
- [ ] 测试了容器运行
- [ ] 创建了docker-compose.yml（开发环境）

### 14.3 健康检查
- [ ] 实现了健康检查接口（可选）
- [ ] gRPC服务正常启动
- [ ] 可以连接到数据库
- [ ] 可以连接到Redis（如使用）
- [ ] 可以连接到RabbitMQ（如使用）

---

## 最终验证

### 功能验证
- [ ] 所有gRPC接口正常工作
- [ ] 数据正确持久化到数据库
- [ ] 缓存正常工作（如使用）
- [ ] 服务间调用正常（如使用）
- [ ] 消息队列正常（如使用）

### 性能验证
- [ ] 服务正常启动（< 10秒）
- [ ] 接口响应时间合理（< 1秒）
- [ ] 无内存泄漏
- [ ] 无goroutine泄漏
- [ ] 数据库连接数正常

### 稳定性验证
- [ ] 可以优雅关闭
- [ ] 重启后数据一致
- [ ] 处理并发请求正常
- [ ] 错误恢复正常
- [ ] 日志正常输出

---

## 常见问题检查

### 编译问题
- [ ] `go mod tidy`已运行
- [ ] 所有依赖已下载
- [ ] 没有import循环依赖
- [ ] 生成的Proto代码是最新的

### 运行时问题
- [ ] 配置文件路径正确
- [ ] 数据库连接字符串正确
- [ ] 端口没有被占用
- [ ] 文件权限正确

### 数据问题
- [ ] 数据库迁移已执行
- [ ] 表结构正确
- [ ] 索引已创建
- [ ] 数据正确序列化/反序列化

---

## 下一步

所有检查项完成后：

1. ✅ 提交代码到版本控制
2. ✅ 创建Pull Request
3. ✅ 进行Code Review
4. ✅ 运行CI/CD流程
5. ✅ 部署到测试环境
6. ✅ 进行集成测试
7. ✅ 部署到生产环境

---

## 参考文档

- [OVERVIEW.md](./00_OVERVIEW.md) - 编码方案总览
- [PROJECT_STRUCTURE.md](./01_PROJECT_STRUCTURE.md) - 项目结构
- 各层实现文档（02-12）

---

**提示**：
- 逐项检查，不要跳过
- 发现问题及时修复
- 保持代码质量
- 遵循项目规范
