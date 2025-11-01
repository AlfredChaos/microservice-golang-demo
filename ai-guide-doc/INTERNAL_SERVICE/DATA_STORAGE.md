# 数据存储层设计

## 概述

内部服务支持多种数据存储方案，所有数据存储都通过依赖注入方式接入，遵循仓储模式（Repository Pattern），实现业务逻辑与数据存储的解耦。

## 支持的数据存储

### 1. PostgreSQL（关系型数据库）
- **用途**：存储结构化业务数据
- **ORM**：GORM
- **迁移工具**：Goose
- **特点**：ACID事务、强一致性、支持复杂查询

### 2. MongoDB（文档数据库）
- **用途**：存储非结构化文档数据
- **驱动**：官方mongo-driver
- **特点**：灵活的模式、水平扩展

### 3. Redis（缓存）
- **用途**：会话管理、热数据缓存、分布式锁
- **驱动**：go-redis
- **特点**：高性能、支持多种数据结构

---

## 设计原则

### 1. 仓储模式（Repository Pattern）

通过仓储接口抽象数据访问，业务层不关心数据存储的具体实现。

**优势**：
- 业务逻辑与数据存储解耦
- 易于切换存储方案（PostgreSQL ↔ MongoDB）
- 易于编写单元测试（使用Mock）
- 统一的错误处理

**示例**：

```go
// 定义仓库接口（业务层视角）
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id string) (*domain.User, error)
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, id string) error
}

// PostgreSQL实现
type userPgRepository struct {
    db *gorm.DB
}

// MongoDB实现
type userMongoRepository struct {
    collection *mongo.Collection
}
```

### 2. 接口隔离

根据不同的使用场景定义最小化接口。

```go
// 只读接口
type UserReader interface {
    GetByID(ctx context.Context, id string) (*domain.User, error)
    List(ctx context.Context, offset, limit int) ([]*domain.User, error)
}

// 只写接口
type UserWriter interface {
    Create(ctx context.Context, user *domain.User) error
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, id string) error
}

// 完整接口（组合）
type UserRepository interface {
    UserReader
    UserWriter
}
```

---

## PostgreSQL集成

### 1. 配置

```yaml
# configs/user-service.yaml
database:
  enabled: true                        # 是否启用PostgreSQL
  driver: postgres
  host: localhost
  port: 5432
  username: postgres
  password: password
  database: user_service
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600
  log_level: info                      # gorm日志级别
```

### 2. 初始化

```go
// pkg/db/postgres.go
package db

import (
    "fmt"
    "time"
    
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// DatabaseConfig PostgreSQL配置
type DatabaseConfig struct {
    Enabled         bool   `yaml:"enabled" mapstructure:"enabled"`
    Driver          string `yaml:"driver" mapstructure:"driver"`
    Host            string `yaml:"host" mapstructure:"host"`
    Port            int    `yaml:"port" mapstructure:"port"`
    Username        string `yaml:"username" mapstructure:"username"`
    Password        string `yaml:"password" mapstructure:"password"`
    Database        string `yaml:"database" mapstructure:"database"`
    MaxOpenConns    int    `yaml:"max_open_conns" mapstructure:"max_open_conns"`
    MaxIdleConns    int    `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
    ConnMaxLifetime int    `yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`
    LogLevel        string `yaml:"log_level" mapstructure:"log_level"`
}

// NewPostgresDB 创建PostgreSQL连接
func NewPostgresDB(cfg *DatabaseConfig) (*gorm.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
        cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database,
    )
    
    // 配置日志
    var logLevel logger.LogLevel
    switch cfg.LogLevel {
    case "silent":
        logLevel = logger.Silent
    case "error":
        logLevel = logger.Error
    case "warn":
        logLevel = logger.Warn
    case "info":
        logLevel = logger.Info
    default:
        logLevel = logger.Info
    }
    
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logLevel),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }
    
    // 配置连接池
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get database instance: %w", err)
    }
    
    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
    sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
    
    return db, nil
}

// MustNewPostgresDB 创建PostgreSQL连接，失败则panic
func MustNewPostgresDB(cfg *DatabaseConfig) *gorm.DB {
    db, err := NewPostgresDB(cfg)
    if err != nil {
        panic(fmt.Sprintf("failed to create postgres db: %v", err))
    }
    return db
}
```

### 3. 数据迁移（Goose）

**安装Goose**：

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

**创建迁移文件**：

```bash
# 在项目根目录执行
mkdir -p internal/user-service/migrations
cd internal/user-service/migrations

# 创建迁移文件
goose create create_users_table sql
```

**迁移文件示例**：

```sql
-- internal/user-service/migrations/001_create_users_table.sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
```

**执行迁移**：

```bash
# 设置数据库连接
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="host=localhost port=5432 user=postgres password=password dbname=user_service sslmode=disable"

# 执行迁移
goose -dir internal/user-service/migrations up

# 回滚迁移
goose -dir internal/user-service/migrations down

# 查看迁移状态
goose -dir internal/user-service/migrations status
```

**在代码中执行迁移**：

```go
// internal/user-service/data/migrate.go
package data

import (
    "database/sql"
    "embed"
    
    "github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// RunMigrations 执行数据库迁移
func RunMigrations(db *sql.DB) error {
    goose.SetBaseFS(embedMigrations)
    
    if err := goose.SetDialect("postgres"); err != nil {
        return err
    }
    
    if err := goose.Up(db, "migrations"); err != nil {
        return err
    }
    
    return nil
}
```

### 4. 仓库实现

```go
// internal/user-service/data/user_pg_repo.go
package data

import (
    "context"
    "errors"
    
    "github.com/alfredchaos/demo/internal/user-service/domain"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

// UserPO 用户持久化对象（与数据库表映射）
type UserPO struct {
    ID        string    `gorm:"column:id;primaryKey"`
    Username  string    `gorm:"column:username;uniqueIndex;not null"`
    Email     string    `gorm:"column:email;uniqueIndex;not null"`
    CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名
func (UserPO) TableName() string {
    return "users"
}

// ToDomain 转换为领域对象
func (po *UserPO) ToDomain() *domain.User {
    return &domain.User{
        ID:        po.ID,
        Username:  po.Username,
        Email:     po.Email,
        CreatedAt: po.CreatedAt,
        UpdatedAt: po.UpdatedAt,
    }
}

// FromDomain 从领域对象转换
func (po *UserPO) FromDomain(user *domain.User) {
    po.ID = user.ID
    po.Username = user.Username
    po.Email = user.Email
    po.CreatedAt = user.CreatedAt
    po.UpdatedAt = user.UpdatedAt
}

// userPgRepository PostgreSQL仓库实现
type userPgRepository struct {
    db *gorm.DB
}

// NewUserPgRepository 创建PostgreSQL仓库
func NewUserPgRepository(db *gorm.DB) UserRepository {
    return &userPgRepository{db: db}
}

func (r *userPgRepository) Create(ctx context.Context, user *domain.User) error {
    // 生成UUID
    if user.ID == "" {
        user.ID = uuid.New().String()
    }
    
    po := &UserPO{}
    po.FromDomain(user)
    
    return r.db.WithContext(ctx).Create(po).Error
}

func (r *userPgRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
    var po UserPO
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&po).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, domain.ErrUserNotFound
        }
        return nil, err
    }
    return po.ToDomain(), nil
}

func (r *userPgRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
    var po UserPO
    err := r.db.WithContext(ctx).Where("username = ?", username).First(&po).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, domain.ErrUserNotFound
        }
        return nil, err
    }
    return po.ToDomain(), nil
}

func (r *userPgRepository) Update(ctx context.Context, user *domain.User) error {
    po := &UserPO{}
    po.FromDomain(user)
    
    return r.db.WithContext(ctx).Save(po).Error
}

func (r *userPgRepository) Delete(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Delete(&UserPO{}, "id = ?", id).Error
}

func (r *userPgRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, error) {
    var pos []UserPO
    err := r.db.WithContext(ctx).
        Offset(offset).
        Limit(limit).
        Order("created_at DESC").
        Find(&pos).Error
    if err != nil {
        return nil, err
    }
    
    users := make([]*domain.User, len(pos))
    for i, po := range pos {
        users[i] = po.ToDomain()
    }
    
    return users, nil
}
```

---

## MongoDB集成

### 1. 配置

```yaml
# configs/user-service.yaml
mongodb:
  enabled: false                       # 是否启用MongoDB
  uri: mongodb://admin:password@localhost:27017
  database: user_service
  max_pool_size: 100
  min_pool_size: 10
  connect_timeout: 10
```

### 2. 仓库实现

```go
// internal/user-service/data/user_mongo_repo.go
package data

import (
    "context"
    "time"
    
    "github.com/alfredchaos/demo/internal/user-service/domain"
    "github.com/alfredchaos/demo/pkg/db"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// UserMongoPO MongoDB持久化对象
type UserMongoPO struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    Username  string             `bson:"username"`
    Email     string             `bson:"email"`
    CreatedAt time.Time          `bson:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at"`
}

// ToDomain 转换为领域对象
func (po *UserMongoPO) ToDomain() *domain.User {
    return &domain.User{
        ID:        po.ID.Hex(),
        Username:  po.Username,
        Email:     po.Email,
        CreatedAt: po.CreatedAt,
        UpdatedAt: po.UpdatedAt,
    }
}

// userMongoRepository MongoDB仓库实现
type userMongoRepository struct {
    collection *mongo.Collection
}

// NewUserMongoRepository 创建MongoDB仓库
func NewUserMongoRepository(client *db.MongoClient) UserRepository {
    collection := client.Database().Collection("users")
    
    // 创建索引
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    indexes := []mongo.IndexModel{
        {
            Keys:    bson.D{{Key: "username", Value: 1}},
            Options: options.Index().SetUnique(true),
        },
        {
            Keys:    bson.D{{Key: "email", Value: 1}},
            Options: options.Index().SetUnique(true),
        },
    }
    
    _, _ = collection.Indexes().CreateMany(ctx, indexes)
    
    return &userMongoRepository{
        collection: collection,
    }
}

func (r *userMongoRepository) Create(ctx context.Context, user *domain.User) error {
    po := &UserMongoPO{
        Username:  user.Username,
        Email:     user.Email,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    result, err := r.collection.InsertOne(ctx, po)
    if err != nil {
        return err
    }
    
    user.ID = result.InsertedID.(primitive.ObjectID).Hex()
    return nil
}

func (r *userMongoRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, domain.ErrUserNotFound
    }
    
    var po UserMongoPO
    err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&po)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, domain.ErrUserNotFound
        }
        return nil, err
    }
    
    return po.ToDomain(), nil
}

func (r *userMongoRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
    var po UserMongoPO
    err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&po)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, domain.ErrUserNotFound
        }
        return nil, err
    }
    
    return po.ToDomain(), nil
}
```

---

## Redis集成

### 1. 配置

```yaml
# configs/user-service.yaml
redis:
  enabled: true                        # 是否启用Redis
  addr: localhost:6379
  password: ""
  db: 0
  pool_size: 10
  min_idle_conns: 5
  dial_timeout: 5
  read_timeout: 3
  write_timeout: 3
```

### 2. 缓存使用示例

```go
// internal/user-service/data/user_cache.go
package data

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/alfredchaos/demo/internal/user-service/domain"
    "github.com/alfredchaos/demo/pkg/cache"
)

const (
    userCachePrefix = "user:"
    userCacheTTL    = 5 * time.Minute
)

// UserCache 用户缓存
type UserCache struct {
    redis *cache.RedisClient
}

// NewUserCache 创建用户缓存
func NewUserCache(redis *cache.RedisClient) *UserCache {
    return &UserCache{redis: redis}
}

// Get 从缓存获取用户
func (c *UserCache) Get(ctx context.Context, id string) (*domain.User, error) {
    key := fmt.Sprintf("%s%s", userCachePrefix, id)
    
    data, err := c.redis.Get(ctx, key).Bytes()
    if err != nil {
        return nil, err
    }
    
    var user domain.User
    if err := json.Unmarshal(data, &user); err != nil {
        return nil, err
    }
    
    return &user, nil
}

// Set 设置用户缓存
func (c *UserCache) Set(ctx context.Context, user *domain.User) error {
    key := fmt.Sprintf("%s%s", userCachePrefix, user.ID)
    
    data, err := json.Marshal(user)
    if err != nil {
        return err
    }
    
    return c.redis.Set(ctx, key, data, userCacheTTL).Err()
}

// Delete 删除用户缓存
func (c *UserCache) Delete(ctx context.Context, id string) error {
    key := fmt.Sprintf("%s%s", userCachePrefix, id)
    return c.redis.Del(ctx, key).Err()
}
```

**带缓存的仓库实现**：

```go
// internal/user-service/data/user_cached_repo.go
package data

import (
    "context"
    "github.com/alfredchaos/demo/internal/user-service/domain"
)

// userCachedRepository 带缓存的仓库实现
type userCachedRepository struct {
    repo  UserRepository  // 底层仓库（PostgreSQL或MongoDB）
    cache *UserCache      // 缓存层
}

// NewUserCachedRepository 创建带缓存的仓库
func NewUserCachedRepository(repo UserRepository, cache *UserCache) UserRepository {
    return &userCachedRepository{
        repo:  repo,
        cache: cache,
    }
}

func (r *userCachedRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
    // 1. 尝试从缓存获取
    user, err := r.cache.Get(ctx, id)
    if err == nil {
        return user, nil
    }
    
    // 2. 缓存未命中，从数据库获取
    user, err = r.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // 3. 写入缓存
    _ = r.cache.Set(ctx, user)
    
    return user, nil
}

func (r *userCachedRepository) Create(ctx context.Context, user *domain.User) error {
    // 1. 写入数据库
    if err := r.repo.Create(ctx, user); err != nil {
        return err
    }
    
    // 2. 写入缓存
    _ = r.cache.Set(ctx, user)
    
    return nil
}

func (r *userCachedRepository) Update(ctx context.Context, user *domain.User) error {
    // 1. 更新数据库
    if err := r.repo.Update(ctx, user); err != nil {
        return err
    }
    
    // 2. 更新缓存
    _ = r.cache.Set(ctx, user)
    
    return nil
}

func (r *userCachedRepository) Delete(ctx context.Context, id string) error {
    // 1. 删除数据库
    if err := r.repo.Delete(ctx, id); err != nil {
        return err
    }
    
    // 2. 删除缓存
    _ = r.cache.Delete(ctx, id)
    
    return nil
}
```

---

## 数据层容器

统一管理所有数据源和仓库。

```go
// internal/user-service/data/data.go
package data

import (
    "context"
    "fmt"
    
    "gorm.io/gorm"
    "github.com/alfredchaos/demo/pkg/cache"
    "github.com/alfredchaos/demo/pkg/db"
    "github.com/alfredchaos/demo/pkg/mq"
)

// Data 数据访问层容器
type Data struct {
    // 基础设施客户端
    pgDB        *gorm.DB
    mongoClient *db.MongoClient
    redisClient *cache.RedisClient
    mqClient    *mq.RabbitMQClient
    
    // 仓库实例（导出，供业务层使用）
    UserRepo UserRepository
}

// NewData 创建数据访问层
func NewData(
    pgDB *gorm.DB,
    mongoClient *db.MongoClient,
    redisClient *cache.RedisClient,
    mqClient *mq.RabbitMQClient,
) (*Data, error) {
    d := &Data{
        pgDB:        pgDB,
        mongoClient: mongoClient,
        redisClient: redisClient,
        mqClient:    mqClient,
    }
    
    // 初始化仓库
    var baseRepo UserRepository
    
    // 根据可用的数据源选择仓库实现
    if pgDB != nil {
        baseRepo = NewUserPgRepository(pgDB)
    } else if mongoClient != nil {
        baseRepo = NewUserMongoRepository(mongoClient)
    } else {
        return nil, fmt.Errorf("no database configured")
    }
    
    // 如果有Redis，使用带缓存的仓库
    if redisClient != nil {
        cache := NewUserCache(redisClient)
        d.UserRepo = NewUserCachedRepository(baseRepo, cache)
    } else {
        d.UserRepo = baseRepo
    }
    
    return d, nil
}

// Close 关闭所有连接
func (d *Data) Close(ctx context.Context) error {
    if d.pgDB != nil {
        sqlDB, err := d.pgDB.DB()
        if err == nil {
            sqlDB.Close()
        }
    }
    
    if d.mongoClient != nil {
        d.mongoClient.Close(ctx)
    }
    
    if d.redisClient != nil {
        d.redisClient.Close()
    }
    
    if d.mqClient != nil {
        d.mqClient.Close()
    }
    
    return nil
}
```

---

## 最佳实践

### 1. PO与DO分离

- **DO (Domain Object)**：领域对象，业务层使用
- **PO (Persistent Object)**：持久化对象，数据层使用
- **好处**：数据库表结构变化不影响业务逻辑

### 2. 使用事务

```go
// WithTransaction 在事务中执行函数
func (d *Data) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
    return d.pgDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // 创建新的Data实例，使用事务连接
        txData := &Data{
            pgDB:     tx,
            UserRepo: NewUserPgRepository(tx),
        }
        
        // 将txData注入到context
        newCtx := context.WithValue(ctx, "txData", txData)
        
        return fn(newCtx)
    })
}
```

### 3. 错误处理

```go
// 统一错误转换
func (r *userPgRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
    var po UserPO
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&po).Error
    if err != nil {
        // 转换数据库错误为领域错误
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, domain.ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return po.ToDomain(), nil
}
```

### 4. 使用Context

所有数据库操作都应该接受context，支持超时和取消。

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

user, err := repo.GetByID(ctx, id)
```

---

## 总结

通过仓储模式和依赖注入，我们实现了：

1. **灵活切换**：可以轻松在PostgreSQL和MongoDB之间切换
2. **分层缓存**：透明地添加Redis缓存层
3. **易于测试**：可以使用Mock仓库进行单元测试
4. **统一接口**：业务层不关心数据存储的具体实现

---

## 参考

- [ARCHITECTURE.md](./ARCHITECTURE.md) - 分层架构设计
- [DI_AND_WIRE.md](./DI_AND_WIRE.md) - 依赖注入实现
