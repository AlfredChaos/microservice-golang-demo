package grpcclient

import "time"

// Config gRPC客户端配置
type Config struct {
	Services []ServiceConfig `yaml:"services" mapstructure:"services"`
}

// ServiceConfig 单个服务配置
type ServiceConfig struct {
	Name    string        `yaml:"name" mapstructure:"name"`       // 服务名称
	Address string        `yaml:"address" mapstructure:"address"` // 服务地址
	Timeout time.Duration `yaml:"timeout" mapstructure:"timeout"` // 连接超时
	
	// 可选配置
	Retry   *RetryConfig  `yaml:"retry" mapstructure:"retry"`     // 重试配置
	TLS     *TLSConfig    `yaml:"tls" mapstructure:"tls"`         // TLS配置
}

// RetryConfig 重试配置
type RetryConfig struct {
	Max         int           `yaml:"max" mapstructure:"max"`                   // 最大重试次数
	Timeout     time.Duration `yaml:"timeout" mapstructure:"timeout"`           // 重试超时
	Backoff     time.Duration `yaml:"backoff" mapstructure:"backoff"`           // 退避时间
}

// TLSConfig TLS配置
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`   // 是否启用TLS
	CertFile string `yaml:"cert_file" mapstructure:"cert_file"` // 证书文件
	KeyFile  string `yaml:"key_file" mapstructure:"key_file"`   // 密钥文件
}
