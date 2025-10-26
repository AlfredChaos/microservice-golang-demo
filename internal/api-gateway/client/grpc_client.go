package client

import (
	"context"
	"fmt"
	"time"

	userv1 "github.com/alfredchaos/demo/api/user/v1"
	orderv1 "github.com/alfredchaos/demo/api/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClients gRPC 客户端集合
// 封装所有后端服务的 gRPC 客户端
type GRPCClients struct {
	UserClient userv1.UserServiceClient
	BookClient orderv1.BookServiceClient
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
	}, nil
}

// CallUserService 调用 user-service
func (c *GRPCClients) CallUserService(ctx context.Context) (string, error) {
	resp, err := c.UserClient.SayHello(ctx, &userv1.HelloRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to call user service: %w", err)
	}
	return resp.Message, nil
}

// CallBookService 调用 book-service
func (c *GRPCClients) CallBookService(ctx context.Context) (string, error) {
	resp, err := c.BookClient.GetBook(ctx, &orderv1.BookRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to call book service: %w", err)
	}
	return resp.Message, nil
}
