package rabbitmq

import (
	"context"

	"github.com/alfredchaos/demo/internal/user-service/messaging"
	"github.com/alfredchaos/demo/pkg/mq"
)

// consumer RabbitMQ 消费者实现
// 实现 messaging.Consumer 接口
type consumer struct {
	mqConsumer *mq.RabbitMQConsumer
}

// NewConsumer 创建 RabbitMQ 消费者
func NewConsumer(client *mq.RabbitMQClient) messaging.Consumer {
	return &consumer{
		mqConsumer: mq.NewRabbitMQConsumer(client),
	}
}

// Consume 开始消费消息
// 将 messaging.MessageHandler 适配到 mq.MessageHandler
func (c *consumer) Consume(ctx context.Context, handler messaging.MessageHandler) error {
	// 适配器：将 messaging.MessageHandler 转换为 mq.MessageHandler
	mqHandler := func(ctx context.Context, message []byte) error {
		return handler(ctx, message)
	}

	// 调用底层消费者
	return c.mqConsumer.Consume(ctx, mqHandler)
}

// Close 关闭消费者
func (c *consumer) Close() error {
	return c.mqConsumer.Close()
}
