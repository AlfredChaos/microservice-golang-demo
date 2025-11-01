package client

import (
	orderv1 "github.com/alfredchaos/demo/api/order/v1"
	userv1 "github.com/alfredchaos/demo/api/user/v1"
)

// ClientFactory gRPC 客户端工厂
// 提供创建各服务 gRPC 客户端的方法
type ClientFactory struct {
	connManager *ConnectionManager
}

// NewClientFactory 创建客户端工厂
func NewClientFactory(connManager *ConnectionManager) *ClientFactory {
	return &ClientFactory{
		connManager: connManager,
	}
}

// CreateUserClient 创建用户服务客户端
func (f *ClientFactory) CreateUserClient() (userv1.UserServiceClient, error) {
	conn, err := f.connManager.GetConnection("user-service")
	if err != nil {
		return nil, err
	}
	return userv1.NewUserServiceClient(conn), nil
}

// CreateBookClient 创建图书服务客户端
func (f *ClientFactory) CreateBookClient() (orderv1.BookServiceClient, error) {
	conn, err := f.connManager.GetConnection("book-service")
	if err != nil {
		return nil, err
	}
	return orderv1.NewBookServiceClient(conn), nil
}

// 扩展示例：添加更多服务客户端创建方法
//
// CreateOrderClient 创建订单服务客户端
// func (f *ClientFactory) CreateOrderClient() (orderv1.OrderServiceClient, error) {
//     conn, err := f.connManager.GetConnection("order-service")
//     if err != nil {
//         return nil, err
//     }
//     return orderv1.NewOrderServiceClient(conn), nil
// }
//
// CreatePaymentClient 创建支付服务客户端
// func (f *ClientFactory) CreatePaymentClient() (paymentv1.PaymentServiceClient, error) {
//     conn, err := f.connManager.GetConnection("payment-service")
//     if err != nil {
//         return nil, err
//     }
//     return paymentv1.NewPaymentServiceClient(conn), nil
// }
