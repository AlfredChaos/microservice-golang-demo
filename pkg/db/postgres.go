package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresConfig PostgreSQL 配置
type PostgresConfig struct {
	Enabled            bool   `yaml:"enabled" mapstructure:"enabled"`
	Driver             string `yaml:"driver" mapstructure:"driver"`
	Host               string `yaml:"host" mapstructure:"host"`                                 // 主机地址
	Port               int    `yaml:"port" mapstructure:"port"`                                 // 端口
	UserName           string `yaml:"username" mapstructure:"username"`                         // 用户名
	Password           string `yaml:"password" mapstructure:"password"`                         // 密码
	Database           string `yaml:"database" mapstructure:"database"`                         // 数据库名称
	SSLMode            string `yaml:"ssl_mode" mapstructure:"ssl_mode"`                         // SSL 模式 (disable, require, verify-ca, verify-full)
	MaxOpenConns       int    `yaml:"max_open_conns" mapstructure:"max_open_conns"`             // 最大打开连接数
	MaxIdleConns       int    `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`             // 最大空闲连接数
	ConnMaxLifetime    int    `yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`       // 连接最大生命周期(秒)
	ConnMaxIdleTime    int    `yaml:"conn_max_idle_time" mapstructure:"conn_max_idle_time"`     // 连接最大空闲时间(秒)
	LogLevel           string `yaml:"log_level" mapstructure:"log_level"`                       // 日志级别 (silent, error, warn, info)
	SlowQueryThreshold int    `yaml:"slow_query_threshold" mapstructure:"slow_query_threshold"` // 慢查询阈值(毫秒)，默认200ms
	EnableDetailedLog  bool   `yaml:"enable_detailed_log" mapstructure:"enable_detailed_log"`   // 是否启用详细日志（记录SQL和参数）
}

// PostgresClient PostgreSQL 客户端封装
type PostgresClient struct {
	db     *gorm.DB
	config *PostgresConfig
}

// NewPostgresClient 创建新的 PostgreSQL 客户端
// 使用工厂模式创建客户端实例,便于测试和依赖注入
func NewPostgresClient(cfg *PostgresConfig) (*PostgresClient, error) {
	// 构建 DSN (Data Source Name)
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.UserName,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
	)

	// 配置 GORM 自定义 Logger（集成现有的 log 包）
	gormConfig := &gorm.Config{
		Logger: NewGormLogger(cfg),
		// 禁用外键约束检查 (可根据需求调整)
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgresql: %w", err)
	}

	// 获取底层的 *sql.DB 用于配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 配置连接池参数
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	}
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)
	}

	// 验证连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgresql: %w", err)
	}

	return &PostgresClient{
		db:     db,
		config: cfg,
	}, nil
}

// GetDB 获取 GORM DB 实例
func (pc *PostgresClient) GetDB() *gorm.DB {
	return pc.db
}

// Close 关闭 PostgreSQL 连接
func (pc *PostgresClient) Close() error {
	if pc.db != nil {
		sqlDB, err := pc.db.DB()
		if err != nil {
			return fmt.Errorf("failed to get sql.DB: %w", err)
		}
		return sqlDB.Close()
	}
	return nil
}

// Ping 检查 PostgreSQL 连接是否正常
func (pc *PostgresClient) Ping() error {
	sqlDB, err := pc.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	return sqlDB.Ping()
}

// Transaction 在事务中执行操作
func (pc *PostgresClient) Transaction(fn func(tx *gorm.DB) error) error {
	return pc.db.Transaction(fn)
}

// AutoMigrate 自动迁移表结构
func (pc *PostgresClient) AutoMigrate(models ...interface{}) error {
	return pc.db.AutoMigrate(models...)
}

// Stats 获取连接池统计信息
func (pc *PostgresClient) Stats() (map[string]interface{}, error) {
	sqlDB, err := pc.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration,
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}

// MustNewPostgresClient 创建 PostgreSQL 客户端,失败则 panic
// 适用于服务启动阶段,数据库连接失败应该直接终止程序
func MustNewPostgresClient(cfg *PostgresConfig) *PostgresClient {
	client, err := NewPostgresClient(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create postgresql client: %v", err))
	}
	return client
}

// parseLogLevel 解析日志级别字符串
func parseLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Warn // 默认使用 warn 级别
	}
}

// ============================================================
// 自定义 GORM Logger（集成现有的 log 包）
// ============================================================

// GormLogger 自定义 GORM 日志记录器，集成项目的 log 包
type GormLogger struct {
	logLevel          logger.LogLevel
	slowThreshold     time.Duration
	enableDetailedLog bool
	ignoreNotFoundErr bool
}

// NewGormLogger 创建新的 GORM Logger
func NewGormLogger(cfg *PostgresConfig) logger.Interface {
	slowThreshold := 200 * time.Millisecond // 默认 200ms
	if cfg.SlowQueryThreshold > 0 {
		slowThreshold = time.Duration(cfg.SlowQueryThreshold) * time.Millisecond
	}

	return &GormLogger{
		logLevel:          parseLogLevel(cfg.LogLevel),
		slowThreshold:     slowThreshold,
		enableDetailedLog: cfg.EnableDetailedLog,
		ignoreNotFoundErr: true, // 默认忽略未找到记录错误
	}
}

// LogMode 设置日志级别
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

// Info 记录 Info 级别日志
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Info {
		log.Info(fmt.Sprintf(msg, data...))
	}
}

// Warn 记录 Warn 级别日志
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Warn {
		log.Warn(fmt.Sprintf(msg, data...))
	}
}

// Error 记录 Error 级别日志
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Error {
		log.Error(fmt.Sprintf(msg, data...))
	}
}

// Trace 记录 SQL 执行详情
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.logLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// 使用 log.WithContext 自动提取上下文信息（trace_id、request_id、user_id 等）
	contextLogger := log.WithContext(ctx).WithOptions(zap.AddCallerSkip(3))

	// 基础字段
	fields := []zap.Field{
		zap.Float64("duration_ms", float64(elapsed.Nanoseconds())/1e6),
		zap.Int64("rows_affected", rows),
	}

	// 根据配置决定是否记录 SQL
	if l.enableDetailedLog {
		fields = append(fields, zap.String("sql", sql))
	}

	switch {
	case err != nil && l.logLevel >= logger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.ignoreNotFoundErr):
		// 错误日志
		fields = append(fields, zap.Error(err))
		contextLogger.Error("postgres query error", fields...)

	case elapsed > l.slowThreshold && l.slowThreshold != 0 && l.logLevel >= logger.Warn:
		// 慢查询警告
		fields = append(fields,
			zap.Bool("is_slow_query", true),
			zap.Float64("threshold_ms", float64(l.slowThreshold.Nanoseconds())/1e6),
		)
		contextLogger.Warn("postgres slow query detected", fields...)

	case l.logLevel >= logger.Info:
		// 普通查询日志
		contextLogger.Info("postgres query", fields...)
	}
}
