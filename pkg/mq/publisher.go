package mq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher 消息发布者接口
// 使用接口定义发布者行为,便于测试和替换实现
type Publisher interface {
	Publish(ctx context.Context, message []byte) error
	Close() error
}

// RabbitMQPublisher RabbitMQ 消息发布者实现
type RabbitMQPublisher struct {
	client *RabbitMQClient
}

// NewRabbitMQPublisher 创建新的 RabbitMQ 发布者
func NewRabbitMQPublisher(client *RabbitMQClient) *RabbitMQPublisher {
	return &RabbitMQPublisher{
		client: client,
	}
}

// Publish 发布消息到 RabbitMQ
// ctx: 上下文,用于控制超时和取消
// message: 要发布的消息内容
func (p *RabbitMQPublisher) Publish(ctx context.Context, message []byte) error {
	if !p.client.IsConnected() {
		return fmt.Errorf("rabbitmq connection is closed")
	}
	
	// 发布消息
	err := p.client.channel.PublishWithContext(
		ctx,
		p.client.config.Exchange,   // 交换机
		p.client.config.RoutingKey, // 路由键
		false,                      // mandatory: 如果为true,当消息无法路由到队列时会返回错误
		false,                      // immediate: 如果为true,当消息无法立即投递给消费者时会返回错误
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         message,
			DeliveryMode: amqp.Persistent, // 持久化消息
		},
	)
	
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	
	return nil
}

// PublishWithOptions 使用自定义选项发布消息
// 提供更灵活的发布方式,允许自定义消息属性
func (p *RabbitMQPublisher) PublishWithOptions(
	ctx context.Context,
	exchange string,
	routingKey string,
	message []byte,
	contentType string,
	persistent bool,
) error {
	if !p.client.IsConnected() {
		return fmt.Errorf("rabbitmq connection is closed")
	}
	
	deliveryMode := amqp.Transient
	if persistent {
		deliveryMode = amqp.Persistent
	}
	
	err := p.client.channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  contentType,
			Body:         message,
			DeliveryMode: deliveryMode,
		},
	)
	
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	
	return nil
}

// Close 关闭发布者
func (p *RabbitMQPublisher) Close() error {
	// 发布者不直接关闭客户端,由客户端管理者负责
	return nil
}
