package config

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

// LoadConfig 加载指定服务的配置文件
// serviceName: 服务名称,如 "api-gateway", "user-service" 等
// cfg: 配置结构体指针,用于接收解析后的配置
func LoadConfig(serviceName string, cfg interface{}) error {
	v := viper.New()
	
	// 设置配置文件名和路径
	v.SetConfigName(serviceName)
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("../configs")
	v.AddConfigPath("../../configs")
	
	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	// 解析配置到结构体
	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	return nil
}

// LoadConfigFromPath 从指定路径加载配置文件
// configPath: 配置文件的完整路径
// cfg: 配置结构体指针
func LoadConfigFromPath(configPath string, cfg interface{}) error {
	v := viper.New()
	
	// 获取文件名和目录
	dir := filepath.Dir(configPath)
	filename := filepath.Base(configPath)
	ext := filepath.Ext(filename)
	name := filename[:len(filename)-len(ext)]
	
	v.SetConfigName(name)
	v.SetConfigType(ext[1:]) // 去掉点号
	v.AddConfigPath(dir)
	
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file from %s: %w", configPath, err)
	}
	
	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	return nil
}

// MustLoadConfig 加载配置,失败则panic
// 适用于服务启动阶段,配置加载失败应该直接终止程序
func MustLoadConfig(serviceName string, cfg interface{}) {
	if err := LoadConfig(serviceName, cfg); err != nil {
		panic(fmt.Sprintf("failed to load config for %s: %v", serviceName, err))
	}
}
