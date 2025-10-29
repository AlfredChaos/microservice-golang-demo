package log

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/alfredchaos/demo/pkg/reqctx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// Logger 全局日志实例
	Logger *zap.Logger
)

// customTimeEncoder 自定义时间编码器
// 格式：2025-10-28 07:46:45.296
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// LogConfig 日志配置
type LogConfig struct {
	Level               string       `yaml:"level" mapstructure:"level"`                                 // 日志级别: debug, info, warn, error
	Format              string       `yaml:"format" mapstructure:"format"`                               // 日志格式: json, console
	OutputPaths         []string     `yaml:"output_paths" mapstructure:"output_paths"`                   // 输出路径列表，支持 stdout 或文件路径
	EnableConsoleWriter bool         `yaml:"enable_console_writer" mapstructure:"enable_console_writer"` // 是否启用 ConsoleWriter（仅对stdout生效）
	Rotation            *RotationConfig `yaml:"rotation" mapstructure:"rotation"`                         // 日志切割配置（可选）
}

// RotationConfig 日志切割配置
type RotationConfig struct {
	MaxSize    int  `yaml:"max_size" mapstructure:"max_size"`       // 每个日志文件的最大尺寸（MB），默认100MB
	MaxAge     int  `yaml:"max_age" mapstructure:"max_age"`         // 日志文件的最大保存天数，默认30天
	MaxBackups int  `yaml:"max_backups" mapstructure:"max_backups"` // 保留的旧日志文件的最大数量，默认10个
	Compress   bool `yaml:"compress" mapstructure:"compress"`       // 是否压缩旧日志文件，默认false
	LocalTime  bool `yaml:"local_time" mapstructure:"local_time"`   // 是否使用本地时间，默认使用UTC时间
}

// WrapWriterLogs 日志切割写入器
type WrapWriterLogs struct {
	*lumberjack.Logger
	currentDay string
}

// NewWrapWriterLogs 创建一个支持按天切割日志文件的 WrapWriterLogs 实例。
// filename: 日志文件名（包含路径，后面会拼上 _{day}.log）
// maxSize: 每个日志文件的最大尺寸（以MB为单位）
// maxAge: 日志文件的最大保存天数
// maxBackups: 保留的旧日志文件的最大数量
func NewWrapWriterLogs(filename string, maxSize, maxAge, maxBackups int, compress, localTime bool) *WrapWriterLogs {
	// 设置默认值
	if maxSize <= 0 {
		maxSize = 100
	}
	if maxAge <= 0 {
		maxAge = 30
	}
	if maxBackups <= 0 {
		maxBackups = 10
	}

	// 生成带日期的文件名
	currentDay := getCurrentDay(localTime)
	filenameWithDay := fmt.Sprintf("%s_%s.log", filename, currentDay)

	return &WrapWriterLogs{
		Logger: &lumberjack.Logger{
			Filename:   filenameWithDay,
			MaxSize:    maxSize,
			MaxAge:     maxAge,
			MaxBackups: maxBackups,
			LocalTime:  localTime,
			Compress:   compress,
		},
		currentDay: currentDay,
	}
}

// Write 实现 io.Writer 接口，支持按天自动切割
func (w *WrapWriterLogs) Write(p []byte) (n int, err error) {
	// 检查是否需要按天切割
	newDay := getCurrentDay(w.Logger.LocalTime)
	if newDay != w.currentDay {
		// 日期变化，关闭旧文件，创建新文件
		_ = w.Logger.Close()
		
		// 更新文件名
		baseFilename := w.Logger.Filename[:len(w.Logger.Filename)-len(w.currentDay)-5] // 去掉 _{day}.log
		w.Logger.Filename = fmt.Sprintf("%s_%s.log", baseFilename, newDay)
		w.currentDay = newDay
	}

	return w.Logger.Write(p)
}

// getCurrentDay 获取当前日期字符串（格式：20060102）
func getCurrentDay(localTime bool) string {
	now := time.Now()
	if !localTime {
		now = now.UTC()
	}
	return now.Format("20060102")
}

// InitLogger 初始化日志系统
// cfg: 日志配置
// serviceName: 服务名称,会添加到日志的 service 字段
func InitLogger(cfg *LogConfig, serviceName string) error {
	// 设置默认值
	if cfg.Level == "" {
		cfg.Level = "debug"
	}
	if len(cfg.OutputPaths) == 0 {
		cfg.OutputPaths = []string{"stdout"}
	}

	// 解析日志级别
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return err
	}

	// 配置编码器 - 使用自定义时间格式
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     customTimeEncoder,             // 使用自定义时间格式：2025-10-28 07:46:45.296
		EncodeDuration: zapcore.MillisDurationEncoder, // 毫秒级别的持续时间
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 构建多个 Core（支持多输出）
	var cores []zapcore.Core

	for _, path := range cfg.OutputPaths {
		var writeSyncer zapcore.WriteSyncer
		var encoder zapcore.Encoder

		if path == "stdout" || path == "" {
			// 输出到标准输出
			if cfg.EnableConsoleWriter {
				// 使用 ConsoleEncoder 格式化输出（彩色、人眼友好）
				consoleEncoderConfig := encoderConfig
				consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 彩色级别
				consoleEncoderConfig.EncodeTime = customTimeEncoder                  // 使用自定义时间格式
				encoder = zapcore.NewConsoleEncoder(consoleEncoderConfig)
			} else {
				// 输出 JSON 格式
				encoder = zapcore.NewJSONEncoder(encoderConfig)
			}
			writeSyncer = zapcore.AddSync(os.Stdout)
		} else {
			// 输出到文件，始终使用 JSON 格式
			encoder = zapcore.NewJSONEncoder(encoderConfig)
			
			// 如果配置了日志切割，使用 WrapWriterLogs
			if cfg.Rotation != nil {
				// 去掉原路径的 .log 后缀（如果有）
				basePath := path
				if len(path) > 4 && path[len(path)-4:] == ".log" {
					basePath = path[:len(path)-4]
				}
				
				wrapWriter := NewWrapWriterLogs(
					basePath,
					cfg.Rotation.MaxSize,
					cfg.Rotation.MaxAge,
					cfg.Rotation.MaxBackups,
					cfg.Rotation.Compress,
					cfg.Rotation.LocalTime,
				)
				writeSyncer = zapcore.AddSync(wrapWriter)
			} else {
				// 不使用日志切割，直接写入文件
				file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return err
				}
				writeSyncer = zapcore.AddSync(file)
			}
		}

		// 创建 Core
		core := zapcore.NewCore(encoder, writeSyncer, level)
		cores = append(cores, core)
	}

	// 合并多个 Core
	var core zapcore.Core
	if len(cores) == 1 {
		core = cores[0]
	} else {
		core = zapcore.NewTee(cores...)
	}

	// 创建 Logger (不设置 CallerSkip，让各个函数自行调整)
	Logger = zap.New(core, zap.AddCaller())

	// 添加服务名称字段
	Logger = Logger.With(zap.String("service", serviceName))

	return nil
}

// MustInitLogger 初始化日志,失败则panic
func MustInitLogger(cfg *LogConfig, serviceName string) {
	if err := InitLogger(cfg, serviceName); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
}

// Info 记录 Info 级别日志
func Info(msg string, fields ...zap.Field) {
	Logger.WithOptions(zap.AddCallerSkip(1)).Info(msg, fields...)
}

// Debug 记录 Debug 级别日志
func Debug(msg string, fields ...zap.Field) {
	Logger.WithOptions(zap.AddCallerSkip(1)).Debug(msg, fields...)
}

// Warn 记录 Warn 级别日志
func Warn(msg string, fields ...zap.Field) {
	Logger.WithOptions(zap.AddCallerSkip(1)).Warn(msg, fields...)
}

// Error 记录 Error 级别日志
func Error(msg string, fields ...zap.Field) {
	Logger.WithOptions(zap.AddCallerSkip(1)).Error(msg, fields...)
}

// Fatal 记录 Fatal 级别日志并退出程序
func Fatal(msg string, fields ...zap.Field) {
	Logger.WithOptions(zap.AddCallerSkip(1)).Fatal(msg, fields...)
}

// Sync 刷新日志缓冲区
func Sync() error {
	if Logger != nil {
		return Logger.Sync()
	}
	return nil
}

// WithTraceID 返回带有 trace_id 的 logger
func WithTraceID(traceID string) *zap.Logger {
	return Logger.With(zap.String("trace_id", traceID))
}

// WithUserID 返回带有 user_id 的 logger
func WithUserID(userID string) *zap.Logger {
	return Logger.With(zap.String("user_id", userID))
}

// WithContext 从 context 中提取所有日志相关字段，返回带有上下文信息的 logger
// 会自动提取以下信息（如果存在）：
// - trace_id: 追踪ID
// - request_id: 请求ID
// - user_id: 用户ID
// - request: 请求信息（method, path, client_ip）
// 如果某个字段在 context 中不存在，则忽略该字段
func WithContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return Logger
	}

	fields := make([]zap.Field, 0, 4)

	// 提取 trace_id
	if traceID := reqctx.GetTraceID(ctx); traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	// 提取 request_id
	if requestID := reqctx.GetRequestID(ctx); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	// 提取 user_id
	if userID := reqctx.GetUserID(ctx); userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	// 提取请求信息
	if reqInfo := reqctx.GetRequestInfo(ctx); reqInfo != nil {
		fields = append(fields, zap.Object("request", &requestContext{
			Method:   reqInfo.Method,
			Path:     reqInfo.Path,
			ClientIP: reqInfo.ClientIP,
		}))
	}

	return Logger.With(fields...)
}

// WithRequest 返回带有请求上下文的 logger
func WithRequest(method, path, clientIP string) *zap.Logger {
	return Logger.With(
		zap.String("request.method", method),
		zap.String("request.path", path),
		zap.String("request.client_ip", clientIP),
	)
}

// WithDuration 返回带有操作耗时的 logger (毫秒)
func WithDuration(durationMs int64) *zap.Logger {
	return Logger.With(zap.Int64("duration_ms", durationMs))
}

// WithError 返回带有错误信息的 logger
func WithError(err error) *zap.Logger {
	return Logger.With(zap.Error(err))
}

// WithExtraData 返回带有业务自定义数据的 logger
func WithExtraData(key string, value interface{}) *zap.Logger {
	return Logger.With(zap.Any("extra_data."+key, value))
}

// TraceID 字段构造器
func TraceID(traceID string) zap.Field {
	return zap.String("trace_id", traceID)
}

// UserID 字段构造器
func UserID(userID string) zap.Field {
	return zap.String("user_id", userID)
}

// Request 字段构造器
func Request(method, path, clientIP string) zap.Field {
	return zap.Object("request", &requestContext{
		Method:   method,
		Path:     path,
		ClientIP: clientIP,
	})
}

// DurationMs 字段构造器 (毫秒)
func DurationMs(durationMs int64) zap.Field {
	return zap.Int64("duration_ms", durationMs)
}

// ExtraData 字段构造器
func ExtraData(key string, value interface{}) zap.Field {
	return zap.Any("extra_data."+key, value)
}

// requestContext 请求上下文
type requestContext struct {
	Method   string
	Path     string
	ClientIP string
}

// MarshalLogObject 实现 zapcore.ObjectMarshaler 接口
func (r *requestContext) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("method", r.Method)
	enc.AddString("path", r.Path)
	enc.AddString("client_ip", r.ClientIP)
	return nil
}
