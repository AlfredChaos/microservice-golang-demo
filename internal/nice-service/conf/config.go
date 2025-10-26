package conf

import (
	"github.com/alfredchaos/demo/pkg/log"
	"github.com/alfredchaos/demo/pkg/mq"
)

// Config nice-service 配置结构
type Config struct {
	Server   ServerConfig     `yaml:"server" mapstructure:"server"`       // 服务器配置
	Log      log.LogConfig    `yaml:"log" mapstructure:"log"`             // 日志配置
	RabbitMQ mq.RabbitMQConfig `yaml:"rabbitmq" mapstructure:"rabbitmq"`  // RabbitMQ 配置
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Name string `yaml:"name" mapstructure:"name"` // 服务名称
}
