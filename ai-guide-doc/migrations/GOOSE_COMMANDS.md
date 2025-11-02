# Goose 命令对照表

本项目已完整实现 goose 的所有主要迁移命令。

## 命令对照

| Goose 原生命令 | 项目实现 | Makefile 命令 | 说明 |
|---------------|---------|--------------|------|
| `goose up` | ✅ `migrations.MigrateUp()` | `make migrate-up` | 迁移到最新版本 |
| `goose up-to VERSION` | ✅ `migrations.MigrateUpTo()` | `make migrate-up-to VERSION=N` | 迁移到指定版本（升级） |
| `goose down` | ✅ `migrations.MigrateDown()` | `make migrate-down` | 回滚最后一次迁移 |
| `goose down-to VERSION` | ✅ `migrations.MigrateDownTo()` | `make migrate-down-to VERSION=N` | 回滚到指定版本（降级） |
| `goose status` | ✅ `migrations.MigrateStatus()` | `make migrate-status` | 查看迁移状态 |
| `goose version` | ✅ `migrations.GetCurrentVersion()` | `make migrate-version` | 查看当前数据库版本 |
| `goose reset` | ✅ `migrations.MigrateReset()` | `make migrate-reset` | 重置数据库（回滚所有迁移） |
| `goose create NAME sql` | ⚠️ 需要 goose CLI | 使用 goose 工具 | 创建新的迁移文件 |

## 使用示例

### 1. up - 迁移到最新版本

```bash
# 使用 Makefile
make migrate-up

# 或直接使用编译的二进制
./build/migrate -cmd=up

# 或使用 go run
go run cmd/migrate/main.go -cmd=up
```

**对应 goose 命令**：
```bash
goose -dir migrations/shared-db postgres "CONNECTION_STRING" up
```

### 2. up-to - 迁移到指定版本（升级）

```bash
# 使用 Makefile（推荐）
make migrate-up-to VERSION=20251101000000

# 或直接使用
./build/migrate -cmd=up-to -version=20251101000000

# 或使用 go run
go run cmd/migrate/main.go -cmd=up-to -version=20251101000000
```

**对应 goose 命令**：
```bash
goose -dir migrations/shared-db postgres "CONNECTION_STRING" up-to 20251101000000
```

**使用场景**：
- 想要迁移到特定版本而不是最新版本
- 在测试环境中逐步验证迁移
- 回滚后重新迁移到某个安全的版本

### 3. down - 回滚最后一次迁移

```bash
# 使用 Makefile
make migrate-down

# 或直接使用
./build/migrate -cmd=down
```

**对应 goose 命令**：
```bash
goose -dir migrations/shared-db postgres "CONNECTION_STRING" down
```

**注意**：只回滚一次，如果需要回滚多次，需要多次执行。

### 4. down-to - 回滚到指定版本（降级）

```bash
# 使用 Makefile（推荐）
make migrate-down-to VERSION=20251101000000

# 或直接使用
./build/migrate -cmd=down-to -version=20251101000000
```

**对应 goose 命令**：
```bash
goose -dir migrations/shared-db postgres "CONNECTION_STRING" down-to 20251101000000
```

**使用场景**：
- 需要回滚到某个已知稳定的版本
- 修复迁移错误，回滚到问题之前的版本
- 在生产环境中谨慎地回滚多个迁移

### 5. status - 查看迁移状态

```bash
# 使用 Makefile
make migrate-status

# 或直接使用
./build/migrate -cmd=status
```

**输出示例**：
```
Applied At                  Migration
=======================================
2025-11-01 00:00:00 UTC     20251101000000_create_users_table.sql
2025-11-02 00:00:00 UTC     20251102000000_create_orders_table.sql
Pending                     20251103000000_add_user_status.sql
```

**对应 goose 命令**：
```bash
goose -dir migrations/shared-db postgres "CONNECTION_STRING" status
```

### 6. version - 查看当前版本

```bash
# 使用 Makefile
make migrate-version

# 或直接使用
./build/migrate -cmd=version
```

**输出示例**：
```
2025-11-02 12:00:00 INFO Current database version {"version": 20251102000000}
```

**对应 goose 命令**：
```bash
goose -dir migrations/shared-db postgres "CONNECTION_STRING" version
```

### 7. reset - 重置数据库

```bash
# 使用 Makefile（会要求确认）
make migrate-reset

# 或直接使用
./build/migrate -cmd=reset
```

**⚠️ 警告**：此命令会删除所有数据，请谨慎使用！

**对应 goose 命令**：
```bash
goose -dir migrations/shared-db postgres "CONNECTION_STRING" reset
```

### 8. create - 创建新的迁移文件

**需要使用 goose CLI**：

```bash
# 安装 goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# 创建迁移文件
cd migrations/shared-db
goose create add_user_status sql

# 会生成类似：20251103120530_add_user_status.sql
```

**为什么不在项目中实现 create 命令？**

1. **复杂度**：create 命令需要处理文件系统操作，时间戳生成等
2. **标准化**：使用 goose 官方工具可以保证文件格式的标准化
3. **简单性**：开发者可以直接使用 goose CLI，无需重复实现

## 完整的工作流程示例

### 场景 1：开发新功能

```bash
# 1. 创建迁移文件
cd migrations/shared-db
goose create add_user_profile sql

# 2. 编辑生成的文件
vim 20251103120530_add_user_profile.sql

# 3. 执行迁移
cd ../..
make migrate-up

# 4. 查看状态
make migrate-status

# 5. 如果有问题，回滚
make migrate-down

# 6. 修复后重新迁移
make migrate-up
```

### 场景 2：生产环境部署

```bash
# 1. 查看当前版本
make migrate-version

# 2. 查看待执行的迁移
make migrate-status

# 3. 备份数据库
pg_dump $DATABASE_URL > backup_$(date +%Y%m%d).sql

# 4. 执行迁移
make migrate-up-prod

# 5. 验证迁移结果
make migrate-status

# 6. 如果有问题，回滚到之前的版本
make migrate-down-to VERSION=20251102000000
```

### 场景 3：测试环境验证

```bash
# 1. 迁移到特定版本进行测试
make migrate-up-to VERSION=20251103000000

# 2. 运行测试
go test ./...

# 3. 继续迁移到下一个版本
make migrate-up-to VERSION=20251104000000

# 4. 运行测试
go test ./...

# 5. 如果某个版本有问题，回滚
make migrate-down-to VERSION=20251103000000
```

## 命令参数说明

### 通用参数

| 参数 | 说明 | 必需 | 默认值 |
|------|------|------|--------|
| `-cmd` | 迁移命令 | 是 | `up` |
| `-config` | 配置文件路径 | 否 | `configs/user-service.yaml` |

### 版本相关参数

| 参数 | 适用命令 | 说明 | 示例 |
|------|---------|------|------|
| `-version` | `up-to`, `down-to` | 目标版本号（时间戳格式） | `20251101000000` |

## 与 goose CLI 的区别

| 特性 | 项目实现 | goose CLI |
|------|---------|-----------|
| 迁移执行 | ✅ 支持 | ✅ 支持 |
| 配置管理 | ✅ 使用项目配置文件 | ❌ 需要命令行参数 |
| 日志集成 | ✅ 集成项目日志系统 | ❌ 标准输出 |
| 创建迁移 | ❌ 使用 goose CLI | ✅ 支持 |
| 嵌入式部署 | ✅ 编译为二进制 | ❌ 需要安装工具 |
| CI/CD 集成 | ✅ 无需额外工具 | ⚠️ 需要安装 goose |

## 最佳实践

### 1. 开发环境

```bash
# 使用 goose create 创建迁移文件
cd migrations/shared-db
goose create add_feature sql

# 使用 make 命令执行迁移
make migrate-up
make migrate-status
```

### 2. CI/CD 环境

```yaml
# .gitlab-ci.yml
migrate:
  script:
    - make build-migrate  # 编译 migrate 工具
    - ./build/migrate -cmd=up -config=configs/production.yaml
    - ./build/migrate -cmd=status
```

### 3. 生产环境

```bash
# 1. 先查看状态
./build/migrate -cmd=status -config=configs/production.yaml

# 2. 执行迁移
./build/migrate -cmd=up -config=configs/production.yaml

# 3. 验证版本
./build/migrate -cmd=version -config=configs/production.yaml
```

### 4. 回滚策略

```bash
# 推荐：使用 down-to 回滚到已知稳定版本
make migrate-down-to VERSION=20251102000000

# 不推荐：多次使用 down 逐个回滚（容易出错）
# make migrate-down
# make migrate-down
# make migrate-down
```

## 常见问题

### Q: 为什么没有实现 create 命令？

**A**: create 命令已由 goose CLI 完美实现，直接使用官方工具可以：
- 保证文件格式标准化
- 自动生成正确的时间戳
- 避免重复造轮子

### Q: up-to 和 version 有什么区别？

**A**:
- `up-to`：迁移**到**指定版本（执行操作）
- `version`：查看当前版本（只读操作）

### Q: down-to 和 down 有什么区别？

**A**:
- `down`：回滚**一次**
- `down-to`：回滚**到**指定版本（可能多次）

### Q: 如何在代码中使用这些命令？

**A**: 直接导入 migrations 包：

```go
import "github.com/alfredchaos/demo/migrations"

// 执行迁移
err := migrations.MigrateUp(sqlDB)

// 迁移到指定版本
err := migrations.MigrateUpTo(sqlDB, 20251101000000)

// 回滚到指定版本
err := migrations.MigrateDownTo(sqlDB, 20251101000000)

// 查看状态
err := migrations.MigrateStatus(sqlDB)

// 获取当前版本
version, err := migrations.GetCurrentVersion(sqlDB)
```

## 总结

本项目已完整实现 goose 的所有主要迁移命令：

✅ **完全实现**：
- up - 迁移到最新
- up-to - 迁移到指定版本
- down - 回滚一次
- down-to - 回滚到指定版本
- status - 查看状态
- version - 查看版本
- reset - 重置数据库

⚠️ **需要外部工具**：
- create - 使用 goose CLI 创建迁移文件

所有命令都可以通过 Makefile、编译的二进制文件或 go run 三种方式执行，灵活方便！
