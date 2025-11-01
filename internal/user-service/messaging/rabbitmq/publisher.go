package rabbitmq

import (
	"context"

	"github.com/alfredchaos/demo/internal/user-service/messaging"
	"github.com/alfredchaos/demo/pkg/mq"
)

// publisher RabbitMQ 发布者实现
type publisher struct {
	mqPublisher *mq.RabbitMQPublisher
	exchange    string
}

// NewPublisher 创建 RabbitMQ 发布者
func NewPublisher(client *mq.RabbitMQClient) messaging.Publisher {
	return &publisher{
		mqPublisher: mq.NewRabbitMQPublisher(client),
		exchange:    "", // 将在 init 中设置
	}
}

// newPublisherWithExchange 创建带交换机配置的发布者
func newPublisherWithExchange(client *mq.RabbitMQClient, exchange string) messaging.Publisher {
	return &publisher{
		mqPublisher: mq.NewRabbitMQPublisher(client),
		exchange:    exchange,
	}
}

// Publish 发布消息
func (p *publisher) Publish(ctx context.Context, message []byte) error {
	return p.mqPublisher.Publish(ctx, message)
}

// PublishWithRouting 使用指定的路由键发布消息
func (p *publisher) PublishWithRouting(ctx context.Context, routingKey string, message []byte) error {
	return p.mqPublisher.PublishWithOptions(
		ctx,
		p.exchange,
		routingKey,
		message,
		"application/json",
		true, // 持久化
	)
}

// Close 关闭发布者
func (p *publisher) Close() error {
	return p.mqPublisher.Close()
}
