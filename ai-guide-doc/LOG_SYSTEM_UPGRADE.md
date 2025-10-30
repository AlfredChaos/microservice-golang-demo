# 日志系统升级文档

## 升级概述

本次升级基于 zap 日志库，实现了功能丰富、灵活可配置的日志系统。

## 主要变更

### 1. 配置文件变更

**旧配置格式：**
```yaml
log:
  level: info
  format: console
  output_path: stdout
```

**新配置格式：**
```yaml
log:
  level: debug  # 默认级别改为 debug
  format: console
  output_paths:  # 支持多输出路径
    - stdout
    - /var/log/app.log  # 可选，同时输出到文件
  enable_console_writer: true  # 控制终端彩色输出
```

### 2. 核心功能

#### 2.1 多输出支持
- ✅ 同时输出到 stdout 和文件
- ✅ stdout 可使用 ConsoleWriter（彩色、人眼友好）
- ✅ 文件始终使用 JSON 格式（便于日志分析）

#### 2.2 时间戳格式
- ✅ 使用 RFC3339Nano 格式：`2025-10-29T15:10:00.123456789Z`
- ✅ 支持纳秒级精度
- ✅ 字段名改为 `timestamp`

#### 2.3 日志字段

所有日志自动包含以下字段：
- `timestamp`: 时间戳（RFC3339Nano）
- `level`: 日志级别
- `service`: 服务名称
- `message`: 日志消息
- `caller`: 代码位置

可选字段（通过辅助函数添加）：
- `trace_id`: 追踪ID
- `user_id`: 用户ID
- `request`: 请求上下文（包含 method, path, client_ip）
- `duration_ms`: 操作耗时（毫秒）
- `error`: 错误信息
- `extra_data.*`: 业务自定义数据

#### 2.4 辅助函数

**字段构造器（用于单次日志调用）：**
```go
log.TraceID(traceID string) zap.Field
log.UserID(userID string) zap.Field
log.Request(method, path, clientIP string) zap.Field
log.DurationMs(durationMs int64) zap.Field
log.ExtraData(key string, value interface{}) zap.Field
```

**上下文 Logger（用于多次日志调用）：**
```go
log.WithTraceID(traceID string) *zap.Logger
log.WithUserID(userID string) *zap.Logger
log.WithRequest(method, path, clientIP string) *zap.Logger
log.WithDuration(durationMs int64) *zap.Logger
log.WithError(err error) *zap.Logger
log.WithExtraData(key string, value interface{}) *zap.Logger
```

### 3. 配置文件更新

已更新以下配置文件：
- ✅ `configs/api-gateway.yaml`
- ✅ `configs/user-service.yaml`
- ✅ `configs/book-service.yaml`
- ✅ `configs/nice-service.yaml`

所有服务默认配置：
- 日志级别：`debug`
- 输出路径：`stdout`
- ConsoleWriter：已启用
- 文件输出：已注释（可按需启用）

### 4. 兼容性

**保持兼容的 API：**
```go
log.Info(msg string, fields ...zap.Field)
log.Debug(msg string, fields ...zap.Field)
log.Warn(msg string, fields ...zap.Field)
log.Error(msg string, fields ...zap.Field)
log.Fatal(msg string, fields ...zap.Field)
log.Sync() error
```

现有代码无需修改即可正常工作。

### 5. 使用示例

#### 基础用法（保持不变）
```go
log.Info("用户登录", zap.String("username", "alice"))
```

#### 使用新辅助函数
```go
log.Info("API请求",
    log.TraceID("trace-123"),
    log.UserID("user-456"),
    log.Request("GET", "/api/users", "192.168.1.100"),
    log.DurationMs(150),
)
```

#### 使用上下文 Logger
```go
ctxLogger := log.WithTraceID("trace-123").
    With(log.UserID("user-456"))

ctxLogger.Info("开始处理")
ctxLogger.Info("处理完成", log.DurationMs(250))
```

## 输出示例

### Console Writer 模式（开发环境）
```
2025-10-29T15:10:00.123456789+08:00    INFO    pkg/log/log.go:127    API请求    {"service": "api-gateway", "trace_id": "trace-123", "user_id": "user-456", "request": {"method": "GET", "path": "/api/users", "client_ip": "192.168.1.100"}, "duration_ms": 150}
```

### JSON 模式（生产环境）
```json
{
  "timestamp": "2025-10-29T15:10:00.123456789Z",
  "level": "info",
  "service": "api-gateway",
  "message": "API请求",
  "caller": "pkg/log/log.go:127",
  "trace_id": "trace-123",
  "user_id": "user-456",
  "request": {
    "method": "GET",
    "path": "/api/users",
    "client_ip": "192.168.1.100"
  },
  "duration_ms": 150
}
```

## 环境配置建议

### 开发环境
```yaml
log:
  level: debug
  output_paths: [stdout]
  enable_console_writer: true
```

### 测试环境
```yaml
log:
  level: debug
  output_paths: [stdout, /var/log/app.log]
  enable_console_writer: true
```

### 生产环境
```yaml
log:
  level: info
  output_paths: [stdout, /var/log/app.log]
  enable_console_writer: false  # 输出 JSON 便于日志分析
```

## 迁移指南

### 无需修改的场景
如果现有代码使用以下方式，无需任何修改：
```go
log.Info("message", zap.String("key", "value"))
log.Error("error message", zap.Error(err))
```

### 建议优化的场景

**场景1：HTTP 请求日志**

旧代码：
```go
log.Info("HTTP request",
    zap.String("method", method),
    zap.String("path", path),
    zap.String("client_ip", clientIP),
    zap.Int64("duration_ms", duration),
)
```

优化后：
```go
log.Info("HTTP request",
    log.Request(method, path, clientIP),
    log.DurationMs(duration),
)
```

**场景2：带追踪ID的多条日志**

旧代码：
```go
log.Info("step1", zap.String("trace_id", traceID))
log.Info("step2", zap.String("trace_id", traceID))
log.Info("step3", zap.String("trace_id", traceID))
```

优化后：
```go
ctxLogger := log.WithTraceID(traceID)
ctxLogger.Info("step1")
ctxLogger.Info("step2")
ctxLogger.Info("step3")
```

## 性能说明

- zap 是零分配的结构化日志库
- 字段构造器是零分配的
- 多输出使用 `zapcore.NewTee` 高效合并
- ConsoleWriter 仅影响终端输出性能，文件输出不受影响

## 文档

详细使用文档请参考：`pkg/log/README.md`

示例代码请参考：`pkg/log/example_test.go`
