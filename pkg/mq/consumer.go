package mq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// MessageHandler 消息处理函数类型
// 使用函数类型定义消息处理器,提供灵活的处理方式
type MessageHandler func(ctx context.Context, message []byte) error

// Consumer 消息消费者接口
type Consumer interface {
	Consume(ctx context.Context, handler MessageHandler) error
	Close() error
}

// RabbitMQConsumer RabbitMQ 消息消费者实现
type RabbitMQConsumer struct {
	client *RabbitMQClient
}

// NewRabbitMQConsumer 创建新的 RabbitMQ 消费者
func NewRabbitMQConsumer(client *RabbitMQClient) *RabbitMQConsumer {
	return &RabbitMQConsumer{
		client: client,
	}
}

// Consume 开始消费消息
// ctx: 上下文,用于控制消费者的生命周期
// handler: 消息处理函数
func (c *RabbitMQConsumer) Consume(ctx context.Context, handler MessageHandler) error {
	if !c.client.IsConnected() {
		return fmt.Errorf("rabbitmq connection is closed")
	}
	
	// 开始消费消息
	msgs, err := c.client.channel.Consume(
		c.client.config.Queue, // 队列名称
		"",                    // 消费者标签
		false,                 // 自动确认: false表示手动确认
		false,                 // 独占
		false,                 // no-local
		false,                 // no-wait
		nil,                   // 额外参数
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}
	
	// 处理消息
	go func() {
		for {
			select {
			case <-ctx.Done():
				// 上下文取消,停止消费
				return
			case msg, ok := <-msgs:
				if !ok {
					// 通道关闭
					return
				}
				
				// 调用处理函数
				if err := handler(ctx, msg.Body); err != nil {
					// 处理失败,拒绝消息并重新入队
					msg.Nack(false, true)
				} else {
					// 处理成功,确认消息
					msg.Ack(false)
				}
			}
		}
	}()
	
	return nil
}

// ConsumeWithOptions 使用自定义选项消费消息
// 提供更细粒度的控制,如自动确认、预取数量等
func (c *RabbitMQConsumer) ConsumeWithOptions(
	ctx context.Context,
	handler MessageHandler,
	autoAck bool,
	prefetchCount int,
) error {
	if !c.client.IsConnected() {
		return fmt.Errorf("rabbitmq connection is closed")
	}
	
	// 设置 QoS (预取数量)
	if prefetchCount > 0 {
		err := c.client.channel.Qos(
			prefetchCount, // 预取数量
			0,             // 预取大小
			false,         // global
		)
		if err != nil {
			return fmt.Errorf("failed to set qos: %w", err)
		}
	}
	
	// 开始消费消息
	msgs, err := c.client.channel.Consume(
		c.client.config.Queue,
		"",
		autoAck,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}
	
	// 处理消息
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				
				if err := handler(ctx, msg.Body); err != nil {
					if !autoAck {
						msg.Nack(false, true)
					}
				} else {
					if !autoAck {
						msg.Ack(false)
					}
				}
			}
		}
	}()
	
	return nil
}

// Close 关闭消费者
func (c *RabbitMQConsumer) Close() error {
	// 消费者不直接关闭客户端,由客户端管理者负责
	return nil
}
