# 数据库迁移指南

本项目使用 [Goose](https://github.com/pressly/goose) 进行数据库版本迁移管理。

## 目录结构

```
internal/user-service/
├── repository/
│   └── psql/
│       ├── migrations/              # 迁移文件目录
│       │   ├── 00001_create_users_table.sql
│       │   ├── 00002_add_user_status.sql
│       │   └── README.md
│       ├── migrate.go               # 迁移管理器
│       ├── init_psql.go            # 初始化逻辑
│       └── user_pg_repo.go         # Repository 实现
└── cmd/
    └── migrate/
        └── main.go                  # 迁移 CLI 工具
```

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
2024-11-01 18:30:15 UTC     00001_create_users_table.sql
2024-11-01 18:35:20 UTC     00002_add_user_status.sql
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
{序号}_{描述}.sql
```

示例：
- `00001_create_users_table.sql` - 创建用户表
- `00002_add_user_status.sql` - 添加用户状态字段
- `00003_create_user_profiles.sql` - 创建用户资料表
- `00004_migrate_legacy_data.sql` - 迁移旧数据

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

#### ❌ 避免做法

1. **不要修改已应用的迁移文件**
   ```bash
   # ❌ 错误：修改已经执行的文件
   vim migrations/00001_create_users_table.sql
   
   # ✅ 正确：创建新的迁移文件
   touch migrations/00003_modify_users_table.sql
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

## 环境配置

### 开发环境 (ENV=development)

- 自动使用 **GORM AutoMigrate**
- 快速迭代，方便开发
- 启动服务时自动执行

```bash
ENV=development go run cmd/user-service/main.go
```

### 生产/测试环境 (ENV=production 或其他)

- 使用 **Goose 版本化迁移**
- 手动执行，安全可控
- 需要显式运行迁移命令

```bash
# 部署前执行迁移
make migrate-up-prod

# 或使用指定配置
go run cmd/migrate/main.go -cmd=up -config=configs/user-service.prod.yaml

# 启动服务
ENV=production go run cmd/user-service/main.go
```

## 常见场景

### 场景1：添加新字段

```sql
-- migrations/00002_add_user_phone.sql
-- +goose Up
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
CREATE INDEX idx_users_phone ON users(phone);

COMMENT ON COLUMN users.phone IS '用户手机号';

-- +goose Down
DROP INDEX IF EXISTS idx_users_phone;
ALTER TABLE users DROP COLUMN IF EXISTS phone;
```

### 场景2：修改字段类型

```sql
-- migrations/00003_modify_email_length.sql
-- +goose Up
ALTER TABLE users ALTER COLUMN email TYPE VARCHAR(500);

-- +goose Down
ALTER TABLE users ALTER COLUMN email TYPE VARCHAR(255);
```

### 场景3：创建新表

```sql
-- migrations/00004_create_user_profiles.sql
-- +goose Up
CREATE TABLE IF NOT EXISTS user_profiles (
    user_id VARCHAR(36) PRIMARY KEY,
    bio TEXT,
    avatar_url VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_profiles_created_at ON user_profiles(created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS user_profiles;
```

### 场景4：数据迁移

```sql
-- migrations/00005_migrate_user_status.sql
-- +goose Up
-- 为现有用户设置默认状态
UPDATE users 
SET status = 'active' 
WHERE status IS NULL;

-- 将旧的 is_active 字段迁移到新的 status 字段
UPDATE users 
SET status = CASE 
    WHEN is_active = true THEN 'active'
    ELSE 'inactive'
END;

-- 删除旧字段
ALTER TABLE users DROP COLUMN IF EXISTS is_active;

-- +goose Down
-- 警告：数据迁移的回滚通常很复杂
-- 确保在执行前备份数据
ALTER TABLE users ADD COLUMN is_active BOOLEAN DEFAULT true;

UPDATE users 
SET is_active = (status = 'active');
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
    - make migrate-up-prod
    
    # 3. 验证迁移
    - make migrate-status
    
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
    # 执行迁移
    make migrate-up-prod
    
    # 检查状态
    make migrate-status
```

## 故障排查

### 问题1：迁移失败

```bash
# 查看错误信息
make migrate-status

# 检查数据库连接
psql -U postgres -h localhost -d user_service

# 手动回滚
make migrate-down
```

### 问题2：版本不一致

```bash
# 查看当前版本
go run cmd/migrate/main.go -cmd=version

# 强制迁移到指定版本
make migrate-version VERSION=1
```

### 问题3：embed 文件找不到

确保迁移文件在正确的位置：
```
internal/user-service/repository/psql/migrations/*.sql
```

## 高级用法

### 使用自定义配置文件

```bash
go run cmd/migrate/main.go \
    -cmd=up \
    -config=configs/user-service.staging.yaml
```

### 在代码中调用

```go
import "github.com/alfredchaos/demo/internal/user-service/repository/psql"

// 执行迁移
if err := psql.MigrateUp(pgClient); err != nil {
    log.Fatal("migration failed", zap.Error(err))
}

// 获取当前版本
version, err := psql.GetCurrentVersion(pgClient)
```

## 参考资源

- [Goose 官方文档](https://github.com/pressly/goose)
- [PostgreSQL 文档](https://www.postgresql.org/docs/)
- [数据库迁移最佳实践](https://en.wikipedia.org/wiki/Schema_migration)

## 注意事项

⚠️ **生产环境迁移检查清单**

- [ ] 在staging环境测试过
- [ ] 备份了生产数据库
- [ ] 通知了相关团队
- [ ] 准备了回滚方案
- [ ] 设置了维护窗口
- [ ] 监控了数据库性能
- [ ] 验证了数据完整性
