package dependencies

import (
	"github.com/alfredchaos/demo/internal/book-service/biz"
	"github.com/alfredchaos/demo/internal/book-service/cache"
	"github.com/alfredchaos/demo/internal/book-service/conf"
	"github.com/alfredchaos/demo/internal/book-service/messaging"
	"github.com/alfredchaos/demo/internal/book-service/messaging/rabbitmq"
	"github.com/alfredchaos/demo/internal/book-service/repository"
	"github.com/alfredchaos/demo/internal/book-service/repository/mongo"
	"github.com/alfredchaos/demo/internal/book-service/repository/psql"
	"github.com/alfredchaos/demo/internal/book-service/service"
	"github.com/alfredchaos/demo/pkg/db"
	"github.com/alfredchaos/demo/pkg/grpcclient"
)

type AppContext struct {
	Data         *repository.Data
	BookCache    cache.BookCache
	MessageQueue messaging.MessageQueue
	BookUseCase  *biz.BookUseCase
	BookService  *service.BookService
}

type Dependencies struct {
	ClientManager *grpcclient.Manager
	Cfg           *conf.Config
}

func InjectDependencies(deps *Dependencies) (*AppContext, error) {
	// 获取 gRPC 客户端（使用 GetClient 自动创建类型化客户端）
	// client, err := deps.ClientManager.GetClient("book-service")
	// if err != nil {
	// 	log.Fatal("failed to get book service client", zap.Error(err))
	// 	return nil, err
	// }
	// bookClient := client.(bookv1.BookServiceClient)

	var pgClient *db.PostgresClient
	var bookRepo repository.BookRepository
	if deps.Cfg.Database.Enabled {
		pgClient = psql.MustInitPostgresClient(&deps.Cfg.Database)
		bookRepo = psql.NewBookPgRepository(pgClient.GetDB())
	}

	var mongoClient *db.MongoClient
	var bookDocumentRepo repository.BookDocumentRepository
	if deps.Cfg.MongoDB.URI != "" {
		mongoCfg := mongo.FromMongoConfig(&deps.Cfg.MongoDB)
		mongoClient = mongo.MustInitMongoClient(mongoCfg)
		bookDocumentRepo = mongo.NewBookMongoDocumentRepository(mongoClient)
	}

	data := repository.NewData(pgClient, mongoClient, bookRepo, bookDocumentRepo)
	// bookCache := cache.NewBookRedisCache(&deps.Cfg.Redis)

	// 初始化 RabbitMQ，book-service 仅作为消息发布者
	messageQueue := rabbitmq.MustInitRabbitMQ(&deps.Cfg.RabbitMQ)
	// publisher, err := messageQueue.NewPublisher()
	// if err != nil {
	// 	log.Fatal("failed to create publisher", zap.Error(err))
	// 	return nil, err
	// }

	bookUseCase := biz.NewBookUseCase()
	bookService := service.NewBookService(bookUseCase)

	return &AppContext{
		Data:         data,
		BookCache:    nil,
		MessageQueue: messageQueue,
		BookUseCase:  bookUseCase,
		BookService:  bookService,
	}, nil
}
