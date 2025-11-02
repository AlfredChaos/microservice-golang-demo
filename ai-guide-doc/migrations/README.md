# 数据库迁移管理

本目录管理项目所有服务的数据库迁移文件。

## 目录结构

```
migrations/
├── migrate.go              # 迁移管理器（独立的基础设施包）
├── shared-db/              # 共享数据库的迁移文件
│   ├── 20251101000000_create_users_table.sql      # user-service
│   ├── 20251102000000_create_orders_table.sql     # order-service (示例)
│   └── 20251103000000_add_user_order_fk.sql       # 跨服务关系 (示例)
└── README.md
```

## 设计原则

### 依赖倒置原则 (DIP)

```
业务层 (user-service, order-service)
  ↓ 依赖
基础设施层 (migrations 包)
  ↑ 不依赖任何业务实现
```

- ✅ **migrations** 是独立的基础设施包
- ✅ **migrate.go** 不依赖具体的 repository 实现
- ✅ 各服务的 repository 层只使用 migrations，而不拥有它

### 共享数据库架构

多个微服务共享同一个 PostgreSQL 数据库时：

- ✅ 统一的迁移管理：所有服务的 schema 在一个地方
- ✅ 避免冲突：使用统一的版本序列
- ✅ 跨服务关系：可以定义外键等跨表关系
- ✅ 单一职责：一个工具管理整个数据库

## 快速开始

### 1. 执行迁移（升级到最新版本）

```bash
# 使用 Makefile
make migrate-up

# 或直接使用 Go 命令
go run cmd/migrate/main.go -cmd=up
```

### 2. 查看迁移状态

```bash
make migrate-status
```

输出示例：
```
Applied At                  Migration
=======================================
2025-11-01 00:00:00 UTC     20251101000000_create_users_table.sql
2025-11-02 00:00:00 UTC     20251102000000_create_orders_table.sql
```

### 3. 回滚最后一次迁移

```bash
make migrate-down
```

### 4. 迁移到指定版本

```bash
# 迁移到版本 1
make migrate-version VERSION=1

# 查看当前版本
go run cmd/migrate/main.go -cmd=version
```

### 5. 重置数据库（危险操作！）

```bash
make migrate-reset
```

## 创建新的迁移文件

### 命名规范

```
{时间戳}_{描述}.sql
```

时间戳格式为 `YYYYMMDDHHmmss`，例如：
- `20251101000000_create_users_table.sql` - 创建用户表
- `20251102000000_create_orders_table.sql` - 创建订单表
- `20251103120000_add_user_status.sql` - 添加用户状态字段

**推荐使用 goose 工具自动生成**：
```bash
# 自动生成带时间戳的迁移文件
cd migrations/shared-db
goose create add_user_status sql

# 会生成类似：20251103120530_add_user_status.sql
```

### 文件格式

每个迁移文件必须包含 `Up` 和 `Down` 两部分：

```sql
-- +goose Up
-- 在此编写升级SQL
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL
);

-- +goose Down
-- 在此编写回滚SQL
DROP TABLE IF EXISTS users;
```

### 最佳实践

#### ✅ 推荐做法

1. **每个迁移文件只做一件事**
   ```sql
   -- 好的例子：00002_add_user_status.sql
   -- +goose Up
   ALTER TABLE users ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'active';
   
   -- +goose Down
   ALTER TABLE users DROP COLUMN status;
   ```

2. **使用 IF EXISTS / IF NOT EXISTS**
   ```sql
   CREATE TABLE IF NOT EXISTS users (...);
   CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
   DROP TABLE IF EXISTS users;
   ```

3. **添加注释**
   ```sql
   COMMENT ON TABLE users IS '用户表';
   COMMENT ON COLUMN users.status IS '用户状态: active, inactive, banned';
   ```

4. **测试回滚**
   ```bash
   make migrate-up     # 执行迁移
   make migrate-down   # 测试回滚
   make migrate-up     # 再次执行
   ```

5. **使用表前缀避免冲突**
   ```sql
   -- 为不同服务的表使用前缀
   CREATE TABLE user_accounts (...);    -- user-service
   CREATE TABLE order_items (...);      -- order-service
   ```

#### ❌ 避免做法

1. **不要修改已应用的迁移文件**
   ```bash
   # ❌ 错误：修改已经执行的文件
   vim migrations/shared-db/20251101000000_create_users_table.sql
   
   # ✅ 正确：创建新的迁移文件
   cd migrations/shared-db
   goose create modify_users_table sql
   ```

2. **不要在 Up 中删除数据**
   ```sql
   -- ❌ 危险！可能导致数据丢失
   -- +goose Up
   DELETE FROM users WHERE created_at < '2024-01-01';
   
   -- ✅ 正确：使用单独的数据迁移脚本，并备份
   ```

3. **不要忽略回滚逻辑**
   ```sql
   -- ❌ 错误：没有提供回滚方法
   -- +goose Down
   -- 什么都不做
   
   -- ✅ 正确：提供完整的回滚
   -- +goose Down
   DROP TABLE IF EXISTS user_profiles;
   ```

## 在代码中使用

### 方式一：通过独立的迁移工具（推荐）

```bash
# 部署前先执行迁移
./migrate -cmd up -config configs/production.yaml

# 启动服务（服务不执行迁移）
./user-service
```

### 方式二：在代码中调用

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
    
    // 获取当前版本
    version, _ := migrations.GetCurrentVersion(sqlDB)
    log.Info("database version:", version)
}
```

## 共享数据库协调

### 迁移执行策略

对于共享数据库架构，推荐以下策略：

1. **独立的迁移服务**（推荐）
   ```bash
   # 在 CI/CD 中作为独立步骤执行
   - name: Database Migration
     run: |
       go run cmd/migrate/main.go -cmd=up
       go run cmd/migrate/main.go -cmd=status
   ```

2. **Leader Election**
   - 如果必须在服务启动时执行迁移
   - 使用分布式锁确保只有一个实例执行迁移
   - goose 本身有锁机制防止并发执行

3. **服务启动前置检查**
   ```go
   // 服务启动时只检查版本，不执行迁移
   currentVersion, _ := migrations.GetCurrentVersion(db)
   requiredVersion := int64(3)
   
   if currentVersion < requiredVersion {
       log.Fatal("database version too old, please run migration")
   }
   ```

### 跨服务 Schema 变更

添加新表：
```sql
-- 00004_create_order_items.sql (order-service 的表)
-- +goose Up
CREATE TABLE IF NOT EXISTS order_items (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS order_items;
```

## 环境配置

### 开发环境

开发环境可以使用 GORM AutoMigrate 或 Goose：

```bash
# 使用 AutoMigrate (快速迭代)
ENV=development go run cmd/user-service/main.go

# 或使用 Goose (与生产保持一致)
make migrate-up
go run cmd/user-service/main.go
```

### 生产环境

生产环境必须使用独立的迁移工具：

```bash
# 1. 备份数据库
pg_dump $DATABASE_URL > backup_$(date +%Y%m%d_%H%M%S).sql

# 2. 执行迁移
make migrate-up-prod

# 3. 验证迁移
make migrate-status

# 4. 部署服务
kubectl apply -f k8s/deployment.yaml
```

## CI/CD 集成

### GitLab CI 示例

```yaml
# .gitlab-ci.yml
deploy-production:
  stage: deploy
  script:
    # 1. 备份数据库
    - pg_dump $DATABASE_URL > backup_$(date +%Y%m%d_%H%M%S).sql
    
    # 2. 执行迁移
    - go run cmd/migrate/main.go -cmd=up -config=configs/production.yaml
    
    # 3. 验证迁移
    - go run cmd/migrate/main.go -cmd=status
    
    # 4. 部署应用
    - kubectl apply -f k8s/deployment.yaml
  only:
    - main
```

### GitHub Actions 示例

```yaml
# .github/workflows/deploy.yml
- name: Run Migrations
  run: |
    go run cmd/migrate/main.go -cmd=up -config=configs/production.yaml
    go run cmd/migrate/main.go -cmd=status
```

## 故障排查

### 问题1：迁移失败

```bash
# 查看错误信息
make migrate-status

# 检查数据库连接
psql -U postgres -h localhost -d shared_db

# 手动回滚
make migrate-down
```

### 问题2：多个服务同时执行迁移

```
Error: goose: failed to acquire lock
```

- goose 使用数据库锁防止并发执行
- 等待其他实例完成迁移即可
- 或确保通过独立的迁移服务执行

### 问题3：版本不一致

```bash
# 查看当前版本
go run cmd/migrate/main.go -cmd=version

# 强制迁移到指定版本
make migrate-version VERSION=1
```

## 参考资源

- [Goose 官方文档](https://github.com/pressly/goose)
- [PostgreSQL 文档](https://www.postgresql.org/docs/)
- [依赖倒置原则 (DIP)](https://en.wikipedia.org/wiki/Dependency_inversion_principle)

## 注意事项

⚠️ **生产环境迁移检查清单**

- [ ] 在 staging 环境测试过
- [ ] 备份了生产数据库
- [ ] 通知了相关团队
- [ ] 准备了回滚方案
- [ ] 设置了维护窗口
- [ ] 监控了数据库性能
- [ ] 验证了数据完整性
- [ ] 确保只有一个实例执行迁移
