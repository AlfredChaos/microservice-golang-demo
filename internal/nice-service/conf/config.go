package conf

import (
	"fmt"

	"github.com/alfredchaos/demo/pkg/cache"
	"github.com/alfredchaos/demo/pkg/db"
	"github.com/alfredchaos/demo/pkg/grpcclient"
	"github.com/alfredchaos/demo/pkg/log"
	"github.com/alfredchaos/demo/pkg/mq"
)

// 配置类型别名
type (
	DatabaseConfig = db.PostgresConfig
	CacheConfig    = cache.RedisConfig
	MQConfig       = mq.RabbitMQConfig
)

// Config nice-service 配置结构
type Config struct {
	Server      ServerConfig      `yaml:"server" mapstructure:"server"`             // 服务器配置（未来可能需要）
	Log         log.LogConfig     `yaml:"log" mapstructure:"log"`                   // 日志配置
	RabbitMQ    MQConfig          `yaml:"rabbitmq" mapstructure:"rabbitmq"`         // 消息队列配置（主要）
	GRPCClients grpcclient.Config `yaml:"grpc_clients" mapstructure:"grpc_clients"` // gRPC客户端配置（未来可能需要）
	
	// 未来可能需要的配置（暂时注释）
	// Database    DatabaseConfig    `yaml:"database" mapstructure:"database"`
	// MongoDB     db.MongoConfig    `yaml:"mongodb" mapstructure:"mongodb"`
	// Redis       CacheConfig       `yaml:"redis" mapstructure:"redis"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Name string `yaml:"name" mapstructure:"name"` // 服务名称
	Host string `yaml:"host" mapstructure:"host"` // 监听地址
	Port int    `yaml:"port" mapstructure:"port"` // 监听端口
}

// GetAddr 获取完整的服务地址
func (c *ServerConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
