package mq

// RoutingKeys 定义所有服务使用的 RabbitMQ Routing Key
// 使用 Topic Exchange 模式，支持通配符匹配
// 命名规范：{服务}.{业务}.{操作}
const (
	// ============================================================
	// Task Service Routing Keys (耗时任务服务)
	// ============================================================
	
	// RoutingKeyTaskSayHelloCreate 创建SayHello任务
	RoutingKeyTaskSayHelloCreate = "task.sayhello.create"
	
	// RoutingKeyTaskSayHelloCompleted 任务完成通知
	RoutingKeyTaskSayHelloCompleted = "task.sayhello.completed"
	
	// RoutingKeyTaskSayHelloFailed 任务失败通知
	RoutingKeyTaskSayHelloFailed = "task.sayhello.failed"
	
	// RoutingKeyTaskPattern 监听所有task消息的通配符模式
	RoutingKeyTaskPattern = "task.#"
	
	// ============================================================
	// Subscription Service Routing Keys (订阅服务)
	// ============================================================
	
	// RoutingKeySubscriptionDeductTime 扣除时长
	RoutingKeySubscriptionDeductTime = "subscription.deduct.time"
	
	// RoutingKeySubscriptionDeductCredit 扣除积分
	RoutingKeySubscriptionDeductCredit = "subscription.deduct.credit"
	
	// RoutingKeySubscriptionPattern 监听所有subscription消息的通配符模式
	RoutingKeySubscriptionPattern = "subscription.#"
	
	// ============================================================
	// User Service Routing Keys (用户服务)
	// ============================================================
	
	// RoutingKeyUserCreated 用户创建事件
	RoutingKeyUserCreated = "user.created"
	
	// RoutingKeyUserUpdated 用户更新事件
	RoutingKeyUserUpdated = "user.updated"
	
	// RoutingKeyUserDeleted 用户删除事件
	RoutingKeyUserDeleted = "user.deleted"
	
	// RoutingKeyUserNotifyPattern 监听所有用户通知的通配符模式
	RoutingKeyUserNotifyPattern = "user.notify.#"
	
	// ============================================================
	// Nice Service Routing Keys (Nice服务)
	// ============================================================
	
	// RoutingKeyNiceProcess 处理Nice请求
	RoutingKeyNiceProcess = "nice.process"
	
	// RoutingKeyNicePattern 监听所有nice消息的通配符模式
	RoutingKeyNicePattern = "nice.#"
)

// ExchangeNames 定义所有交换机名称
const (
	// ExchangeMicroserviceEvents 微服务事件统一交换机
	ExchangeMicroserviceEvents = "microservice_events"
)
