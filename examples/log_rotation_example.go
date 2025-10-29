package main

import (
	"time"

	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

func main() {
	// 示例 1: 不使用日志切割
	example1()

	// 示例 2: 基础日志切割配置
	example2()

	// 示例 3: 完整的日志切割配置
	example3()
}

// example1 不使用日志切割
func example1() {
	cfg := &log.LogConfig{
		Level:       "info",
		OutputPaths: []string{"./logs/basic.log"},
	}

	log.MustInitLogger(cfg, "basic-service")
	log.Info("这是一个不带切割的日志示例")
}

// example2 基础日志切割配置
func example2() {
	cfg := &log.LogConfig{
		Level:       "info",
		OutputPaths: []string{"./logs/app.log"},
		Rotation: &log.RotationConfig{
			MaxSize:    10,    // 10MB 切割一次（测试用）
			MaxAge:     7,     // 保留 7 天
			MaxBackups: 5,     // 最多 5 个备份
			LocalTime:  true,  // 使用本地时间
		},
	}

	log.MustInitLogger(cfg, "app-service")
	
	// 写入一些日志
	for i := 0; i < 100; i++ {
		log.Info("测试日志切割",
			zap.Int("index", i),
			zap.String("message", "这是一条测试消息，用于验证日志切割功能"),
		)
		time.Sleep(10 * time.Millisecond)
	}
}

// example3 完整的日志切割配置（生产环境推荐）
func example3() {
	cfg := &log.LogConfig{
		Level: "info",
		OutputPaths: []string{
			"stdout",                      // 输出到控制台
			"./logs/production.log",       // 输出到文件（带切割）
		},
		EnableConsoleWriter: true,  // 控制台启用彩色输出
		Rotation: &log.RotationConfig{
			MaxSize:    100,   // 100MB 切割一次
			MaxAge:     30,    // 保留 30 天
			MaxBackups: 15,    // 最多 15 个备份
			Compress:   true,  // 压缩旧日志
			LocalTime:  true,  // 使用本地时间
		},
	}

	log.MustInitLogger(cfg, "production-service")
	
	// 模拟生产环境日志
	log.Info("应用启动")
	log.Info("数据库连接成功", zap.String("host", "localhost:5432"))
	log.Info("Redis 连接成功", zap.String("host", "localhost:6379"))
	log.Info("HTTP 服务器启动", zap.Int("port", 8080))
	
	// 模拟一些业务日志
	for i := 0; i < 50; i++ {
		log.Info("处理用户请求",
			zap.Int("user_id", 1000+i),
			zap.String("action", "查询订单"),
			zap.Duration("duration", time.Millisecond*time.Duration(10+i)),
		)
		
		if i%10 == 0 {
			log.Warn("慢查询警告",
				zap.Int("user_id", 1000+i),
				zap.Duration("duration", time.Millisecond*100),
			)
		}
		
		time.Sleep(20 * time.Millisecond)
	}
	
	log.Info("应用正常运行中")
}
