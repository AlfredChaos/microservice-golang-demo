# 日志切割功能使用指南

## 功能概述

日志切割功能支持自动管理日志文件，防止单个日志文件过大，并自动清理过期日志。

## 核心特性

### 1. **按天切割**
- 每天自动创建新的日志文件
- 文件名格式：`{basename}_{YYYYMMDD}.log`
- 示例：`app_20251029.log`, `app_20251030.log`

### 2. **按大小切割**
- 当日志文件达到指定大小时自动切割
- 可配置最大文件大小（MB）

### 3. **自动清理**
- 自动删除超过保存天数的日志文件
- 可配置保留的旧日志文件数量

### 4. **压缩支持**
- 可选择是否压缩旧日志文件（gzip）
- 节省磁盘空间

## 配置说明

### YAML 配置格式

```yaml
log:
  level: info
  format: json
  output_paths:
    - ./logs/app.log
  enable_console_writer: false
  rotation:
    max_size: 100      # 每个日志文件最大 100MB
    max_age: 30        # 保留 30 天的日志
    max_backups: 10    # 最多保留 10 个旧日志文件
    compress: true     # 压缩旧日志文件
    local_time: true   # 使用本地时间（而非 UTC）
```

### 配置项说明

#### `rotation` (可选)
日志切割配置，如果不配置则不启用日志切割功能。

- **`max_size`** (int): 每个日志文件的最大尺寸（MB）
  - 默认值：100
  - 当文件达到此大小时，会创建新文件
  
- **`max_age`** (int): 日志文件的最大保存天数
  - 默认值：30
  - 超过此天数的日志文件会被自动删除
  
- **`max_backups`** (int): 保留的旧日志文件的最大数量
  - 默认值：10
  - 超过此数量的最旧文件会被删除
  - 设置为 0 表示保留所有旧文件
  
- **`compress`** (bool): 是否压缩旧日志文件
  - 默认值：false
  - 启用后，切割的日志文件会被压缩为 `.gz` 格式
  
- **`local_time`** (bool): 是否使用本地时间
  - 默认值：false（使用 UTC 时间）
  - 建议设置为 true，使用本地时间更直观

## 使用示例

### 1. 基础配置（不使用日志切割）

```yaml
log:
  level: info
  output_paths:
    - ./logs/app.log
```

这种配置会将所有日志写入单个文件，不会自动切割。

### 2. 启用日志切割

```yaml
log:
  level: info
  output_paths:
    - ./logs/app.log
  rotation:
    max_size: 50       # 每个文件最大 50MB
    max_age: 7         # 保留 7 天
    max_backups: 5     # 最多保留 5 个旧文件
    compress: true     # 压缩旧文件
    local_time: true   # 使用本地时间
```

### 3. 多输出（控制台 + 文件切割）

```yaml
log:
  level: debug
  output_paths:
    - stdout           # 输出到控制台
    - ./logs/app.log   # 同时写入文件（带切割）
  enable_console_writer: true
  rotation:
    max_size: 100
    max_age: 30
    max_backups: 10
    compress: true
    local_time: true
```

### 4. 不同服务的配置示例

#### API Gateway
```yaml
log:
  level: info
  output_paths:
    - stdout
    - ./logs/api-gateway.log
  enable_console_writer: true
  rotation:
    max_size: 100      # API 日志较多，设置较大的文件
    max_age: 30
    max_backups: 15    # 保留更多备份
    compress: true
    local_time: true
```

#### User Service
```yaml
log:
  level: debug
  output_paths:
    - ./logs/user-service.log
  rotation:
    max_size: 50
    max_age: 15
    max_backups: 5
    compress: true
    local_time: true
```

## 日志文件命名规则

### 当天日志文件
```
{basename}_{YYYYMMDD}.log
```
示例：
- `app_20251029.log` - 2025年10月29日的日志
- `api-gateway_20251030.log` - API Gateway 2025年10月30日的日志

### 切割后的文件（按大小）
```
{basename}_{YYYYMMDD}.log.1
{basename}_{YYYYMMDD}.log.2
...
```

### 压缩后的文件
```
{basename}_{YYYYMMDD}.log.1.gz
{basename}_{YYYYMMDD}.log.2.gz
...
```

## 实际文件示例

假设配置：
```yaml
output_paths: ["./logs/app.log"]
rotation:
  max_size: 10
  max_age: 7
  max_backups: 3
  compress: true
  local_time: true
```

运行一段时间后，`logs` 目录可能包含：
```
logs/
├── app_20251029.log           # 当天的日志
├── app_20251029.log.1.gz      # 当天第一次切割（已压缩）
├── app_20251028.log.1.gz      # 昨天的日志
├── app_20251027.log.1.gz      # 前天的日志
└── app_20251026.log.1.gz      # 3天前的日志
```

## 自动清理规则

日志文件会根据以下规则自动清理：

1. **按时间**: 删除超过 `max_age` 天的所有文件
2. **按数量**: 如果备份文件超过 `max_backups`，删除最旧的文件
3. **优先级**: 时间规则优先于数量规则

## 代码使用示例

### Go 代码初始化

```go
package main

import (
    "github.com/alfredchaos/demo/pkg/log"
)

func main() {
    cfg := &log.LogConfig{
        Level: "info",
        OutputPaths: []string{"./logs/app.log"},
        Rotation: &log.RotationConfig{
            MaxSize:    100,
            MaxAge:     30,
            MaxBackups: 10,
            Compress:   true,
            LocalTime:  true,
        },
    }
    
    log.MustInitLogger(cfg, "my-service")
    
    // 使用日志
    log.Info("application started")
}
```

### 动态配置示例

```go
// 从配置文件加载
import (
    "github.com/spf13/viper"
    "github.com/alfredchaos/demo/pkg/log"
)

func initLogger() {
    var cfg log.LogConfig
    if err := viper.UnmarshalKey("log", &cfg); err != nil {
        panic(err)
    }
    
    log.MustInitLogger(&cfg, "my-service")
}
```

## 性能建议

### 1. 合理设置文件大小
- **高流量服务**: 设置较大的 `max_size` (100-500MB)，减少切割频率
- **低流量服务**: 设置较小的 `max_size` (10-50MB)，便于查看

### 2. 保留天数建议
- **生产环境**: 30-90 天
- **测试环境**: 7-15 天
- **开发环境**: 3-7 天

### 3. 压缩建议
- **生产环境**: 建议开启压缩，节省磁盘空间
- **开发环境**: 可以不开启，方便实时查看

### 4. 备份数量
- 根据磁盘空间和日志重要性决定
- 一般设置 5-15 个备份即可

## 磁盘空间估算

假设配置：
```yaml
max_size: 100      # MB
max_age: 30        # 天
max_backups: 10    # 个
compress: true     # 压缩率约 10:1
```

**最大磁盘占用**（未压缩）：
```
当天日志: 100 MB
备份: 10 × 100 MB = 1000 MB
总计: 约 1.1 GB
```

**实际磁盘占用**（启用压缩）：
```
当天日志: 100 MB (未压缩)
备份: 10 × 10 MB = 100 MB (压缩后)
总计: 约 200 MB
```

## 故障排查

### 问题1: 日志文件没有按天切割

**可能原因**:
- 没有配置 `rotation` 参数
- 应用没有跨天运行

**解决方案**:
检查配置文件，确保 `rotation` 配置存在。

### 问题2: 日志文件一直增长，不切割

**可能原因**:
- `max_size` 设置过大
- 日志写入量较小

**解决方案**:
调整 `max_size` 参数，或等待文件达到设定大小。

### 问题3: 旧日志没有自动删除

**可能原因**:
- `max_age` 和 `max_backups` 设置过大
- 日志切割不频繁

**解决方案**:
检查配置，确保清理规则合理。只有在新日志切割时才会检查清理。

## 最佳实践

1. ✅ **生产环境务必启用日志切割**
2. ✅ **使用本地时间** (`local_time: true`)
3. ✅ **启用压缩** 节省磁盘空间
4. ✅ **设置合理的保留期** 平衡磁盘空间和日志可用性
5. ✅ **监控磁盘空间** 避免日志占满磁盘
6. ✅ **定期检查日志配置** 根据实际情况调整参数

## 总结

日志切割功能提供了完善的日志文件管理能力：
- 🔄 自动按天和按大小切割
- 🗑️ 自动清理过期日志
- 📦 可选的压缩功能
- ⚙️ 灵活的配置选项

合理配置日志切割可以有效管理磁盘空间，同时保留足够的日志用于问题排查。
