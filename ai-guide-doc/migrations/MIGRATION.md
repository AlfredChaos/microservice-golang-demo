# 数据库迁移架构重构说明

## 重构概述

本次重构将数据库迁移从 `internal/user-service/repository/psql/` 提升为独立的基础设施层 `migrations/`，符合**依赖倒置原则 (DIP)**。

## 变更内容

### 文件结构变更

#### 重构前 ❌
```
internal/user-service/repository/
├── psql/
│   ├── migrate.go                    # 迁移逻辑
│   ├── migrations/                   # SQL 文件
│   │   ├── 20251101000000_create_users_table.sql
│   │   └── README.md
│   ├── init_psql.go                  # 启动时自动执行迁移
│   └── user_pg_repo.go
```

**问题**：
- ❌ 违反依赖倒置原则（schema 定义在实现层）
- ❌ 多服务共享数据库时冲突
- ❌ 运维困难，无法独立管理数据库版本

#### 重构后 ✅
```
migrations/                           # 独立的基础设施包
├── migrate.go                        # 迁移管理器
├── shared-db/                        # 共享数据库的迁移
│   └── 20251101000000_create_users_table.sql
├── README.md
└── MIGRATION.md                      # 本文档

cmd/
└── migrate/                          # 迁移工具
    └── main.go                       # 使用 migrations 包

internal/user-service/repository/
└── psql/
    ├── init_psql.go                  # 只初始化连接，不执行迁移
    └── user_pg_repo.go
```

**优势**：
- ✅ 符合依赖倒置原则
- ✅ 支持多服务共享数据库
- ✅ 独立的迁移管理
- ✅ 便于 DevOps 和 CI/CD 集成

### 代码变更

#### 1. `migrations/migrate.go` (新建)
```go
package migrations

import (
	"database/sql"
	"embed"
	"github.com/pressly/goose/v3"
)

//go:embed shared-db/*.sql
var SharedDBMigrations embed.FS

func MigrateUp(db *sql.DB) error { ... }
func MigrateDown(db *sql.DB) error { ... }
// ... 其他迁移函数
```

**关键点**：
- 使用 `sql.DB` 而非 `db.PostgresClient`（更通用）
- 不依赖任何业务层代码
- 可以被多个服务使用

#### 2. `cmd/migrate/main.go` (修改)
```go
import (
	"github.com/alfredchaos/demo/migrations"  // 新导入
)

func main() {
	client, _ := db.NewPostgresClient(&cfg.Database)
	sqlDB, _ := client.GetDB().DB()
	
	// 使用独立的 migrations 包
	migrations.MigrateUp(sqlDB)
}
```

#### 3. `internal/user-service/repository/psql/init_psql.go` (修改)
```go
// 重构前：
func InitPostgresClient(cfg *db.PostgresConfig) (*db.PostgresClient, error) {
	client, _ := db.NewPostgresClient(cfg)
	
	// ❌ 启动时自动执行迁移
	if os.Getenv("ENV") == "development" {
		migrateModels(client.GetDB())
	} else {
		psql.MigrateUp(client)
	}
	
	return client, nil
}

// 重构后：
func InitPostgresClient(cfg *db.PostgresConfig) (*db.PostgresClient, error) {
	client, _ := db.NewPostgresClient(cfg)
	
	// ✅ 只初始化连接，不执行迁移
	log.Info("Note: Database migrations should be run separately")
	
	return client, nil
}
```

## 使用方法

### 方式一：独立迁移工具（推荐）

```bash
# 1. 先执行迁移
go run cmd/migrate/main.go -cmd=up -config=configs/user-service.yaml

# 2. 启动服务
go run cmd/user-service/main.go
```

### 方式二：使用 Makefile

```bash
# 执行迁移
make migrate-up

# 查看状态
make migrate-status

# 回滚
make migrate-down
```

### 方式三：在代码中调用

```go
import (
	"github.com/alfredchaos/demo/migrations"
	"github.com/alfredchaos/demo/pkg/db"
)

func main() {
	// 创建数据库客户端
	pgClient, _ := db.NewPostgresClient(cfg)
	sqlDB, _ := pgClient.GetDB().DB()
	
	// 执行迁移
	if err := migrations.MigrateUp(sqlDB); err != nil {
		log.Fatal("migration failed", err)
	}
	
	// ... 启动服务
}
```

## 多服务共享数据库场景

### 迁移文件组织

```
migrations/shared-db/
├── 20251101000000_create_users_table.sql      # user-service
├── 20251102000000_create_orders_table.sql     # order-service
├── 20251103000000_create_products_table.sql   # product-service
└── 20251104000000_add_user_order_fk.sql       # 跨服务关系
```

### 迁移执行策略

**推荐方式：独立的迁移 Job**

```yaml
# kubernetes/migration-job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: db-migration
spec:
  template:
    spec:
      containers:
      - name: migrate
        image: your-app:latest
        command: ["./migrate", "-cmd=up"]
      restartPolicy: OnFailure
```

**优势**：
- 只有一个 Job 执行迁移
- 避免多个服务实例并发执行
- 迁移失败时服务不会启动

## CI/CD 集成

### GitLab CI 示例

```yaml
# .gitlab-ci.yml
stages:
  - build
  - migrate
  - deploy

migrate-production:
  stage: migrate
  script:
    # 1. 备份
    - pg_dump $DATABASE_URL > backup.sql
    
    # 2. 执行迁移
    - go run cmd/migrate/main.go -cmd=up -config=configs/production.yaml
    
    # 3. 验证
    - go run cmd/migrate/main.go -cmd=status
  only:
    - main

deploy-services:
  stage: deploy
  needs:
    - migrate-production
  script:
    - kubectl apply -f k8s/
```

## 开发工作流

### 添加新的迁移文件

```bash
# 1. 创建迁移文件
vim migrations/shared-db/00002_add_user_status.sql
```

```sql
-- +goose Up
ALTER TABLE users ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'active';
CREATE INDEX idx_users_status ON users(status);

-- +goose Down
DROP INDEX IF EXISTS idx_users_status;
ALTER TABLE users DROP COLUMN status;
```

```bash
# 2. 测试迁移
make migrate-up
make migrate-status

# 3. 测试回滚
make migrate-down
make migrate-up

# 4. 提交代码
git add migrations/shared-db/00002_add_user_status.sql
git commit -m "feat: add user status field"
```

### 本地开发流程

```bash
# 启动本地数据库
docker-compose up -d postgres

# 执行迁移
make migrate-up

# 启动服务
make run-user-service

# 服务启动时会看到：
# INFO: PostgreSQL client initialized successfully
# INFO: Note: Database migrations should be run separately using 'make migrate-up'
```

## 设计原则验证

### ✅ 依赖倒置原则 (DIP)

```
                  高层模块
        ┌─────────────────────────┐
        │   user-service          │
        │   order-service         │
        └────────┬────────────────┘
                 │ 依赖
                 ↓
        ┌─────────────────────────┐
        │   migrations (抽象)     │  ← 基础设施层
        └────────┬────────────────┘
                 │ 不依赖
                 ✗
        ┌─────────────────────────┐
        │   psql (实现细节)       │
        └─────────────────────────┘
```

- ✅ `migrations` 不依赖 `psql` 实现
- ✅ `psql` 不包含 schema 定义
- ✅ 服务层通过 `migrations` 管理 schema

### ✅ 单一职责原则 (SRP)

- **migrations 包**：只负责数据库 schema 版本管理
- **psql 包**：只负责数据访问实现
- **cmd/migrate**：只负责迁移命令行工具

### ✅ 开闭原则 (OCP)

- 添加新服务：只需在 `migrations/shared-db/` 添加迁移文件
- 切换存储实现：不影响 migrations
- 扩展迁移功能：修改 `migrations/migrate.go`，不影响业务代码

## 回滚计划

如果需要回滚到旧架构：

```bash
# 1. 恢复旧文件
git revert <commit-hash>

# 2. 将 SQL 文件移回
mv migrations/shared-db/*.sql internal/user-service/repository/psql/migrations/

# 3. 恢复 init_psql.go 的自动迁移逻辑
```

## 常见问题

### Q: 为什么不在服务启动时自动执行迁移？

**A**: 
- 多实例部署时会并发执行迁移
- 迁移失败不应该阻止所有实例启动
- 生产环境需要人工审批和备份

### Q: 开发环境是否可以自动迁移？

**A**: 可以，但不推荐。建议统一使用 `make migrate-up`：
- 与生产环境保持一致
- 学习正确的迁移流程
- 避免意外的 schema 变更

### Q: 多个服务如何避免迁移冲突？

**A**:
- 使用时间戳格式，自动避免冲突（YYYYMMDDHHmmss）
- 表名添加服务前缀（user_accounts, order_items）
- 通过独立的迁移 Job 执行
- goose 自带锁机制防止并发
- 使用 goose create 自动生成迁移文件

### Q: 如何回滚到指定版本？

**A**:
```bash
# 查看当前版本
make migrate-version

# 回滚到版本 3
make migrate-version VERSION=3
```

## 相关文档

- [migrations/README.md](./README.md) - 迁移使用指南
- [依赖倒置原则](https://en.wikipedia.org/wiki/Dependency_inversion_principle)
- [Goose 文档](https://github.com/pressly/goose)

## 总结

本次重构实现了：
- ✅ 符合 SOLID 原则的架构设计
- ✅ 支持多服务共享数据库
- ✅ 独立的迁移管理和部署
- ✅ 更好的测试和运维体验

所有服务都应该通过 `migrations` 包管理数据库 schema，而不是在各自的 repository 实现中嵌入迁移逻辑。
