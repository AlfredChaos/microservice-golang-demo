package rabbitmq

import (
	"fmt"

	"github.com/alfredchaos/demo/internal/nice-service/messaging"
	"github.com/alfredchaos/demo/pkg/mq"
)

// MessageQueue RabbitMQ 消息队列实现
// 实现 messaging.MessageQueue 接口，提供 RabbitMQ 的具体实现
type MessageQueue struct {
	client *mq.RabbitMQClient
	config *mq.RabbitMQConfig
}

// InitRabbitMQ 初始化 RabbitMQ 消息队列
// 直接使用 mq.RabbitMQConfig，避免配置层冗余
func InitRabbitMQ(cfg *mq.RabbitMQConfig) (*MessageQueue, error) {
	// 检查是否启用
	if !cfg.Enabled {
		return nil, fmt.Errorf("rabbitmq is not enabled")
	}

	// 验证必填配置
	if cfg.URL == "" {
		return nil, fmt.Errorf("rabbitmq url is required")
	}

	// 设置默认值
	if cfg.ExchangeType == "" {
		cfg.ExchangeType = "topic"
	}
	if cfg.RoutingKey == "" {
		cfg.RoutingKey = "#" // 默认接收所有消息
	}

	// 创建 RabbitMQ 客户端（直接使用配置，无需转换）
	client, err := mq.NewRabbitMQClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create rabbitmq client: %w", err)
	}

	return &MessageQueue{
		client: client,
		config: cfg,
	}, nil
}

// NewPublisher 创建发布者
func (mq *MessageQueue) NewPublisher() (messaging.Publisher, error) {
	return newPublisherWithExchange(mq.client, mq.config.Exchange), nil
}

// NewConsumer 创建消费者
func (mq *MessageQueue) NewConsumer() (messaging.Consumer, error) {
	return NewConsumer(mq.client), nil
}

// Close 关闭消息队列连接
func (mq *MessageQueue) Close() error {
	if mq.client != nil {
		return mq.client.Close()
	}
	return nil
}

// IsHealthy 检查连接是否健康
func (mq *MessageQueue) IsHealthy() bool {
	if mq.client == nil {
		return false
	}
	return mq.client.IsConnected()
}

// MustInitRabbitMQ 初始化 RabbitMQ，失败则 panic
func MustInitRabbitMQ(cfg *mq.RabbitMQConfig) *MessageQueue {
	mq, err := InitRabbitMQ(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to init rabbitmq: %v", err))
	}
	return mq
}
