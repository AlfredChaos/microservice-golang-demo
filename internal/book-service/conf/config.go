package conf

import "github.com/alfredchaos/demo/pkg/log"

// Config book-service 配置结构
type Config struct {
	Server ServerConfig  `yaml:"server" mapstructure:"server"` // 服务器配置
	Log    log.LogConfig `yaml:"log" mapstructure:"log"`       // 日志配置
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Name string `yaml:"name" mapstructure:"name"` // 服务名称
	Host string `yaml:"host" mapstructure:"host"` // 监听地址
	Port int    `yaml:"port" mapstructure:"port"` // 监听端口
}
