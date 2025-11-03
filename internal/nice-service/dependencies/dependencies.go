package dependencies

import (
	"github.com/alfredchaos/demo/internal/nice-service/biz"
	"github.com/alfredchaos/demo/internal/nice-service/conf"
	"github.com/alfredchaos/demo/internal/nice-service/messaging"
	"github.com/alfredchaos/demo/internal/nice-service/messaging/rabbitmq"
	"github.com/alfredchaos/demo/internal/nice-service/service"
	"github.com/alfredchaos/demo/pkg/grpcclient"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

// AppContext nice-service 应用上下文
type AppContext struct {
	MessageQueue  messaging.MessageQueue // 消息队列
	Consumer      messaging.Consumer     // 消息消费者
	HandleService *service.HandleService // 消息处理服务（Service层）
	TaskUseCase   *biz.TaskUseCase       // 任务业务逻辑（Biz层）

	// 未来可能需要的字段（暂时注释）
	// GRPCClients  map[string]interface{}  // gRPC客户端
	// Database     *db.PostgresClient      // 数据库连接
	// Cache        cache.Cache             // 缓存
	// NiceService  *service.NiceService    // gRPC服务实现
}

// Dependencies 依赖注入所需的外部依赖
type Dependencies struct {
	ClientManager *grpcclient.Manager // gRPC客户端管理器
	Cfg           *conf.Config        // 配置
}

// InjectDependencies 注入依赖并初始化应用上下文
func InjectDependencies(deps *Dependencies) (*AppContext, error) {
	// 初始化 RabbitMQ 消息队列（nice-service作为消费者）
	messageQueue := rabbitmq.MustInitRabbitMQ(&deps.Cfg.RabbitMQ)
	log.Info("rabbitmq message queue initialized successfully")

	// 创建消费者
	consumer, err := messageQueue.NewConsumer()
	if err != nil {
		log.Error("failed to create consumer", zap.Error(err))
		return nil, err
	}
	log.Info("consumer created successfully")

	// ============================================================
	// 依赖注入 - 按照分层架构组装
	// ============================================================

	// 1. Biz层 - 业务逻辑
	taskUseCase := biz.NewTaskUseCase()
	log.Info("task usecase created successfully")

	// 2. Service层 - 服务层（依赖Biz层）
	handleService := service.NewHandleService(taskUseCase)
	log.Info("handle service created successfully")

	// 未来如果需要 gRPC 客户端调用其他服务
	// client, err := deps.ClientManager.GetClient("user-service")
	// if err != nil {
	//     log.Error("failed to get user service client", zap.Error(err))
	//     return nil, err
	// }
	// userClient := client.(userv1.UserServiceClient)
	// 然后注入到 TaskUseCase: taskUseCase := biz.NewTaskUseCase(userClient)

	// 未来如果需要数据库
	// var pgClient *db.PostgresClient
	// if deps.Cfg.Database.Enabled {
	//     pgClient = psql.MustInitPostgresClient(&deps.Cfg.Database)
	// }
	// 然后注入到 TaskUseCase

	// 未来如果需要缓存
	// var cache cache.Cache
	// if deps.Cfg.Redis.Addr != "" {
	//     cache = cache.NewRedisCache(&deps.Cfg.Redis)
	// }
	// 然后注入到 TaskUseCase

	return &AppContext{
		MessageQueue:  messageQueue,
		Consumer:      consumer,
		HandleService: handleService,
		TaskUseCase:   taskUseCase,
	}, nil
}
