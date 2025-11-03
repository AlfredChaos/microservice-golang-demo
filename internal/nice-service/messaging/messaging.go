package messaging

import "context"

// MessageHandler 消息处理函数类型
type MessageHandler func(ctx context.Context, message []byte) error

// Publisher 消息发布者接口
type Publisher interface {
	Publish(ctx context.Context, message []byte) error
	PublishWithRouting(ctx context.Context, routingKey string, message []byte) error
	Close() error
}

// Consumer 消息消费者接口
type Consumer interface {
	Consume(ctx context.Context, handler MessageHandler) error
	Close() error
}

type MessageQueue interface {
	NewPublisher() (Publisher, error)
	NewConsumer() (Consumer, error)
	Close() error
	IsHealthy() bool
}
