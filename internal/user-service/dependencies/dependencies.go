package dependencies

import (
	bookv1 "github.com/alfredchaos/demo/api/book/v1"
	"github.com/alfredchaos/demo/internal/user-service/biz"
	"github.com/alfredchaos/demo/internal/user-service/cache"
	"github.com/alfredchaos/demo/internal/user-service/conf"
	"github.com/alfredchaos/demo/internal/user-service/messaging"
	"github.com/alfredchaos/demo/internal/user-service/messaging/rabbitmq"
	"github.com/alfredchaos/demo/internal/user-service/repository"
	"github.com/alfredchaos/demo/internal/user-service/repository/mongo"
	"github.com/alfredchaos/demo/internal/user-service/repository/psql"
	"github.com/alfredchaos/demo/internal/user-service/service"
	"github.com/alfredchaos/demo/pkg/db"
	"github.com/alfredchaos/demo/pkg/grpcclient"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

type AppContext struct {
	Data         *repository.Data
	UserCache    cache.UserCache
	MessageQueue messaging.MessageQueue
	UserUseCase  *biz.UserUseCase
	UserService  *service.UserService
}

type Dependencies struct {
	ClientManager *grpcclient.Manager
	Cfg           *conf.Config
}

func InjectDependencies(deps *Dependencies) (*AppContext, error) {
	// 获取 gRPC 客户端（使用 GetClient 自动创建类型化客户端）
	client, err := deps.ClientManager.GetClient("book-service")
	if err != nil {
		log.Fatal("failed to get book service client", zap.Error(err))
		return nil, err
	}
	bookClient := client.(bookv1.BookServiceClient)

	var pgClient *db.PostgresClient
	var userRepo repository.UserRepository
	if deps.Cfg.Database.Enabled {
		pgClient = psql.MustInitPostgresClient(&deps.Cfg.Database)
		userRepo = psql.NewUserPgRepository(pgClient.GetDB())
	}

	var mongoClient *db.MongoClient
	var userDocumentRepo repository.UserDocumentRepository
	if deps.Cfg.MongoDB.URI != "" {
		mongoCfg := mongo.FromMongoConfig(&deps.Cfg.MongoDB)
		mongoClient = mongo.MustInitMongoClient(mongoCfg)
		userDocumentRepo = mongo.NewUserMongoDocumentRepository(mongoClient)
	}

	data := repository.NewData(pgClient, mongoClient, userRepo, userDocumentRepo)
	userCache := cache.NewUserRedisCache(&deps.Cfg.Redis)
	
	// 初始化 RabbitMQ，user-service 仅作为消息发布者
	messageQueue := rabbitmq.MustInitRabbitMQ(&deps.Cfg.RabbitMQ)
	publisher, err := messageQueue.NewPublisher()
	if err != nil {
		log.Fatal("failed to create publisher", zap.Error(err))
		return nil, err
	}

	userUseCase := biz.NewUserUseCase(
		bookClient,
		data.UserRepo,
		data.UserDocumentRepo,
		userCache,
		publisher,
	)

	userService := service.NewUserService(userUseCase)

	return &AppContext{
		Data:         data,
		UserCache:    userCache,
		MessageQueue: messageQueue,
		UserUseCase:  userUseCase,
		UserService:  userService,
	}, nil
}
