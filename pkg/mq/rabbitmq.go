package mq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQConfig RabbitMQ 配置
type RabbitMQConfig struct {
	URL          string `yaml:"url" mapstructure:"url"`                     // RabbitMQ 连接 URL
	Exchange     string `yaml:"exchange" mapstructure:"exchange"`           // 交换机名称
	ExchangeType string `yaml:"exchange_type" mapstructure:"exchange_type"` // 交换机类型: direct, topic, fanout
	Queue        string `yaml:"queue" mapstructure:"queue"`                 // 队列名称
	RoutingKey   string `yaml:"routing_key" mapstructure:"routing_key"`     // 路由键
	Durable      bool   `yaml:"durable" mapstructure:"durable"`             // 是否持久化
	AutoDelete   bool   `yaml:"auto_delete" mapstructure:"auto_delete"`     // 是否自动删除
}

// RabbitMQClient RabbitMQ 客户端封装
type RabbitMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  *RabbitMQConfig
}

// NewRabbitMQClient 创建新的 RabbitMQ 客户端
// 使用工厂模式创建客户端实例
func NewRabbitMQClient(cfg *RabbitMQConfig) (*RabbitMQClient, error) {
	// 连接到 RabbitMQ
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}
	
	// 创建通道
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}
	
	// 声明交换机
	if cfg.Exchange != "" {
		err = channel.ExchangeDeclare(
			cfg.Exchange,     // 交换机名称
			cfg.ExchangeType, // 交换机类型
			cfg.Durable,      // 是否持久化
			cfg.AutoDelete,   // 是否自动删除
			false,            // 是否为内部交换机
			false,            // 是否等待服务器确认
			nil,              // 额外参数
		)
		if err != nil {
			channel.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to declare exchange: %w", err)
		}
	}
	
	// 声明队列
	if cfg.Queue != "" {
		_, err = channel.QueueDeclare(
			cfg.Queue,      // 队列名称
			cfg.Durable,    // 是否持久化
			cfg.AutoDelete, // 是否自动删除
			false,          // 是否独占
			false,          // 是否等待服务器确认
			nil,            // 额外参数
		)
		if err != nil {
			channel.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to declare queue: %w", err)
		}
		
		// 绑定队列到交换机
		if cfg.Exchange != "" {
			err = channel.QueueBind(
				cfg.Queue,      // 队列名称
				cfg.RoutingKey, // 路由键
				cfg.Exchange,   // 交换机名称
				false,          // 是否等待服务器确认
				nil,            // 额外参数
			)
			if err != nil {
				channel.Close()
				conn.Close()
				return nil, fmt.Errorf("failed to bind queue: %w", err)
			}
		}
	}
	
	return &RabbitMQClient{
		conn:    conn,
		channel: channel,
		config:  cfg,
	}, nil
}

// GetChannel 获取 RabbitMQ 通道
func (r *RabbitMQClient) GetChannel() *amqp.Channel {
	return r.channel
}

// GetConnection 获取 RabbitMQ 连接
func (r *RabbitMQClient) GetConnection() *amqp.Connection {
	return r.conn
}

// Close 关闭 RabbitMQ 连接
func (r *RabbitMQClient) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			return err
		}
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// IsConnected 检查连接是否正常
func (r *RabbitMQClient) IsConnected() bool {
	return r.conn != nil && !r.conn.IsClosed()
}

// MustNewRabbitMQClient 创建 RabbitMQ 客户端,失败则 panic
// 适用于服务启动阶段
func MustNewRabbitMQClient(cfg *RabbitMQConfig) *RabbitMQClient {
	client, err := NewRabbitMQClient(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create rabbitmq client: %v", err))
	}
	return client
}
