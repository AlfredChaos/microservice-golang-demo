package client

import (
	"context"
	"fmt"
	"time"

	orderv1 "github.com/alfredchaos/demo/api/order/v1"
	userv1 "github.com/alfredchaos/demo/api/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// GRPCClients gRPC 客户端集合
// 封装所有后端服务的 gRPC 客户端
type GRPCClients struct {
	UserClient userv1.UserServiceClient
	BookClient orderv1.BookServiceClient

	// 保存连接以便关闭
	userConn *grpc.ClientConn
	bookConn *grpc.ClientConn
}

// NewGRPCClients 创建新的 gRPC 客户端集合
// userAddr: user-service 地址
// bookAddr: book-service 地址
func NewGRPCClients(userAddr, bookAddr string) (*GRPCClients, error) {
	// 连接 user-service
	userConn, err := grpc.Dial(
		userAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user-service: %w", err)
	}

	// 连接 book-service
	bookConn, err := grpc.Dial(
		bookAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		userConn.Close()
		return nil, fmt.Errorf("failed to connect to book-service: %w", err)
	}

	return &GRPCClients{
		UserClient: userv1.NewUserServiceClient(userConn),
		BookClient: orderv1.NewBookServiceClient(bookConn),
		userConn:   userConn,
		bookConn:   bookConn,
	}, nil
}

// Close 关闭所有 gRPC 连接
// 应该在服务关闭时调用，防止连接泄漏
func (c *GRPCClients) Close() error {
	var errUser, errBook error

	if c.userConn != nil {
		errUser = c.userConn.Close()
	}

	if c.bookConn != nil {
		errBook = c.bookConn.Close()
	}

	// 返回第一个错误
	if errUser != nil {
		return fmt.Errorf("failed to close user-service connection: %w", errUser)
	}
	if errBook != nil {
		return fmt.Errorf("failed to close book-service connection: %w", errBook)
	}

	return nil
}

// CallUserService 调用 user-service
func (c *GRPCClients) CallUserService(ctx context.Context) (string, error) {
	// 传递 trace ID 到 gRPC metadata
	ctx = c.withTraceID(ctx)

	resp, err := c.UserClient.SayHello(ctx, &userv1.HelloRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to call user service: %w", err)
	}
	return resp.Message, nil
}

// CallBookService 调用 book-service
func (c *GRPCClients) CallBookService(ctx context.Context) (string, error) {
	// 传递 trace ID 到 gRPC metadata
	ctx = c.withTraceID(ctx)

	resp, err := c.BookClient.GetBook(ctx, &orderv1.BookRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to call book service: %w", err)
	}
	return resp.Message, nil
}

// withTraceID 将 trace ID 从 context 中提取并添加到 gRPC metadata
// 用于跨服务追踪请求
func (c *GRPCClients) withTraceID(ctx context.Context) context.Context {
	// 从gRPC上下文中获取trace ID（来自 gin.Context）
	// 在 controller 中，我们使用的是 c.Request.Context()
	// 需要在 request_id 中间件中将 request ID 添加到 context

	// 尝试从 context 中获取 trace ID
	traceID := ""
	if val := ctx.Value("X-Request-ID"); val != nil {
		if id, ok := val.(string); ok {
			traceID = id
		}
	}

	// 如果有 trace ID，添加到 metadata
	if traceID != "" {
		md := metadata.Pairs("X-Trace-ID", traceID)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	return ctx
}
