package consumer

import (
	"context"

	"github.com/alfredchaos/demo/pkg/log"
	"github.com/alfredchaos/demo/pkg/mq"
	"go.uber.org/zap"
)

// MessageConsumer 消息消费者
type MessageConsumer struct {
	consumer mq.Consumer
}

// NewMessageConsumer 创建新的消息消费者
func NewMessageConsumer(consumer mq.Consumer) *MessageConsumer {
	return &MessageConsumer{
		consumer: consumer,
	}
}

// Start 启动消费者
func (mc *MessageConsumer) Start(ctx context.Context) error {
	log.Info("starting message consumer")
	
	// 定义消息处理函数
	handler := func(ctx context.Context, message []byte) error {
		log.Info("received message", zap.String("message", string(message)))
		
		// 打印 "Nice"
		log.Info("Nice")
		
		return nil
	}
	
	// 开始消费消息
	return mc.consumer.Consume(ctx, handler)
}

// Stop 停止消费者
func (mc *MessageConsumer) Stop() error {
	log.Info("stopping message consumer")
	return mc.consumer.Close()
}
