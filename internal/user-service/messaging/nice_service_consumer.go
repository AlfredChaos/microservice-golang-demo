package messaging

import (
	"context"
	"encoding/json"

	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

// NiceMessage nice-service发送的消息结构
type NiceMessage struct {
	LogID   string `json:"log_id"`
	Message string `json:"message"`
}

type NiceMessageHandler interface {
	HandleNiceMessage(ctx context.Context, logID, message string) error
}

// NiceMessageConsumer nice-service消息消费者
type NiceMessageConsumer struct {
	consumer Consumer
	handler  NiceMessageHandler
}

// consumer: 消息队列消费者
// handler: 业务处理器
func NewNiceMessageConsumer(
	consumer Consumer,
	handler NiceMessageHandler,
) *NiceMessageConsumer {
	return &NiceMessageConsumer{
		consumer: consumer,
		handler:  handler,
	}
}

// Start 开始消费消息
func (c *NiceMessageConsumer) Start(ctx context.Context) error {
	log.Info("starting nice message consumer")

	err := c.consumer.Consume(ctx, c.handleMessage)
	if err != nil {
		return err
	}

	log.Info("nice message consumer started")
	return nil
}

func (c *NiceMessageConsumer) handleMessage(ctx context.Context, message []byte) error {
	log.Info("received message from nice-service",
		zap.String("body", string(message)))

	// 解析消息
	var niceMsg NiceMessage
	if err := json.Unmarshal(message, &niceMsg); err != nil {
		log.Error("failed to unmarshal nice message", zap.Error(err))
		return err
	}

	// 调用业务处理器
	if err := c.handler.HandleNiceMessage(ctx, niceMsg.LogID, niceMsg.Message); err != nil {
		log.Error("failed to handle nice message",
			zap.Error(err),
			zap.String("log_id", niceMsg.LogID))
		return err
	}

	log.Info("successfully processed nice message",
		zap.String("log_id", niceMsg.LogID),
		zap.String("message", niceMsg.Message))

	return nil
}

// Close 关闭消费者
func (c *NiceMessageConsumer) Close() error {
	return c.consumer.Close()
}
