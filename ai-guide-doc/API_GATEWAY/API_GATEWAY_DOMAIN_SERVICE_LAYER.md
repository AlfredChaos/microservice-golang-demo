# API Gateway - Domain & Service 层实现

> 领域接口和服务实现层的完整代码

## 1. Domain 层 - 领域接口定义

### 1.1 用户服务接口

#### 文件路径
`internal/api-gateway/domain/user_service.go`

#### 完整代码

```go
package domain

import (
	"context"
)

// IUserService 用户服务领域接口
// 定义用户相关的业务能力
type IUserService interface {
	// SayHello 问候接口
	// 返回问候消息
	SayHello(ctx context.Context) (string, error)
}
```

---

### 1.2 图书服务接口

#### 文件路径
`internal/api-gateway/domain/book_service.go`

#### 完整代码

```go
package domain

import (
	"context"
)

// IBookService 图书服务领域接口
// 定义图书相关的业务能力
type IBookService interface {
	// GetBook 获取图书信息
	// 返回图书信息消息
	GetBook(ctx context.Context) (string, error)
}
```

---

### Domain 层设计说明

1. **纯接口定义**：只定义方法签名，不包含实现
2. **业务语义**：方法名体现业务含义，而非技术细节
3. **上下文传递**：所有方法接收 `context.Context` 用于超时控制和追踪
4. **简洁明了**：接口职责单一，易于理解和实现

---

## 2. Service 层 - 服务实现

### 2.1 用户服务实现

#### 文件路径
`internal/api-gateway/service/user_service.go`

#### 完整代码

```go
package service

import (
	"context"
	"fmt"

	userv1 "github.com/alfredchaos/demo/api/user/v1"
	"github.com/alfredchaos/demo/internal/api-gateway/domain"
	"google.golang.org/grpc/metadata"
)

// userService 用户服务实现
// 封装对 user-service 的 gRPC 调用
type userService struct {
	userClient userv1.UserServiceClient
}

// NewUserService 创建用户服务实例
// 注入 gRPC 客户端依赖
func NewUserService(userClient userv1.UserServiceClient) domain.IUserService {
	return &userService{
		userClient: userClient,
	}
}

// SayHello 调用 user-service 的 SayHello 接口
func (s *userService) SayHello(ctx context.Context) (string, error) {
	// 传递 trace ID 到 gRPC metadata
	ctx = s.withTraceID(ctx)

	// 调用 user-service
	resp, err := s.userClient.SayHello(ctx, &userv1.HelloRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to call user service: %w", err)
	}

	return resp.Message, nil
}

// withTraceID 将 trace ID 从 context 中提取并添加到 gRPC metadata
// 用于跨服务追踪请求
func (s *userService) withTraceID(ctx context.Context) context.Context {
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
```

---

### 2.2 图书服务实现

#### 文件路径
`internal/api-gateway/service/book_service.go`

#### 完整代码

```go
package service

import (
	"context"
	"fmt"

	orderv1 "github.com/alfredchaos/demo/api/order/v1"
	"github.com/alfredchaos/demo/internal/api-gateway/domain"
	"google.golang.org/grpc/metadata"
)

// bookService 图书服务实现
// 封装对 book-service 的 gRPC 调用
type bookService struct {
	bookClient orderv1.BookServiceClient
}

// NewBookService 创建图书服务实例
// 注入 gRPC 客户端依赖
func NewBookService(bookClient orderv1.BookServiceClient) domain.IBookService {
	return &bookService{
		bookClient: bookClient,
	}
}

// GetBook 调用 book-service 的 GetBook 接口
func (s *bookService) GetBook(ctx context.Context) (string, error) {
	// 传递 trace ID 到 gRPC metadata
	ctx = s.withTraceID(ctx)

	// 调用 book-service
	resp, err := s.bookClient.GetBook(ctx, &orderv1.BookRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to call book service: %w", err)
	}

	return resp.Message, nil
}

// withTraceID 将 trace ID 从 context 中提取并添加到 gRPC metadata
// 用于跨服务追踪请求
func (s *bookService) withTraceID(ctx context.Context) context.Context {
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
```

---

### Service 层设计说明

1. **实现 Domain 接口**：返回类型为 `domain.IXxxService`
2. **封装 gRPC 调用**：隐藏 gRPC 调用细节
3. **统一错误处理**：使用 `fmt.Errorf` 包装错误，提供上下文信息
4. **Trace ID 传递**：通过 `withTraceID` 方法传递追踪信息
5. **依赖注入**：通过构造函数注入 gRPC 客户端

---

## 3. 可选优化

### 3.1 提取公共方法

如果多个 Service 都需要 `withTraceID` 方法，可以提取到公共基类：

#### 文件路径
`internal/api-gateway/service/base_service.go`

```go
package service

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// baseService 基础服务
// 提供公共方法供其他服务使用
type baseService struct{}

// withTraceID 将 trace ID 从 context 中提取并添加到 gRPC metadata
// 用于跨服务追踪请求
func (s *baseService) withTraceID(ctx context.Context) context.Context {
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
```

然后在各个 Service 中嵌入：

```go
type userService struct {
	baseService  // 嵌入基础服务
	userClient userv1.UserServiceClient
}

func (s *userService) SayHello(ctx context.Context) (string, error) {
	// 使用继承的方法
	ctx = s.withTraceID(ctx)
	
	resp, err := s.userClient.SayHello(ctx, &userv1.HelloRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to call user service: %w", err)
	}

	return resp.Message, nil
}
```

---

### 3.2 添加超时控制

```go
func (s *userService) SayHello(ctx context.Context) (string, error) {
	// 设置调用超时
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	ctx = s.withTraceID(ctx)
	
	resp, err := s.userClient.SayHello(ctx, &userv1.HelloRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to call user service: %w", err)
	}

	return resp.Message, nil
}
```

---

### 3.3 添加日志记录

```go
import "github.com/alfredchaos/demo/pkg/log"
import "go.uber.org/zap"

func (s *userService) SayHello(ctx context.Context) (string, error) {
	log.WithContext(ctx).Info("calling user service SayHello")
	
	ctx = s.withTraceID(ctx)
	
	resp, err := s.userClient.SayHello(ctx, &userv1.HelloRequest{})
	if err != nil {
		log.WithContext(ctx).Error("failed to call user service", zap.Error(err))
		return "", fmt.Errorf("failed to call user service: %w", err)
	}

	log.WithContext(ctx).Info("user service SayHello success", 
		zap.String("message", resp.Message))
	
	return resp.Message, nil
}
```

---

### 3.4 添加重试机制

```go
import (
	"time"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *userService) SayHello(ctx context.Context) (string, error) {
	ctx = s.withTraceID(ctx)
	
	// 最多重试3次
	maxRetries := 3
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		resp, err := s.userClient.SayHello(ctx, &userv1.HelloRequest{})
		if err == nil {
			return resp.Message, nil
		}
		
		// 只对特定错误重试
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.Unavailable || st.Code() == codes.DeadlineExceeded {
				lastErr = err
				time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
				continue
			}
		}
		
		// 其他错误直接返回
		return "", fmt.Errorf("failed to call user service: %w", err)
	}
	
	return "", fmt.Errorf("failed to call user service after %d retries: %w", maxRetries, lastErr)
}
```

---

## 4. 测试示例

### 4.1 单元测试（使用 Mock）

```go
package service

import (
	"context"
	"testing"

	userv1 "github.com/alfredchaos/demo/api/user/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserServiceClient Mock 用户服务客户端
type MockUserServiceClient struct {
	mock.Mock
}

func (m *MockUserServiceClient) SayHello(ctx context.Context, in *userv1.HelloRequest) (*userv1.HelloResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userv1.HelloResponse), args.Error(1)
}

func TestUserService_SayHello(t *testing.T) {
	// 创建 mock 客户端
	mockClient := new(MockUserServiceClient)
	
	// 设置期望行为
	mockClient.On("SayHello", mock.Anything, mock.Anything).
		Return(&userv1.HelloResponse{Message: "Hello from mock"}, nil)
	
	// 创建服务实例
	service := NewUserService(mockClient)
	
	// 调用方法
	message, err := service.SayHello(context.Background())
	
	// 断言结果
	assert.NoError(t, err)
	assert.Equal(t, "Hello from mock", message)
	
	// 验证 mock 调用
	mockClient.AssertExpectations(t)
}
```

---

## 5. 设计优势

### 5.1 解耦性

- Service 实现 Domain 接口，Controller 依赖 Domain 接口
- 可以轻松替换实现（如用 HTTP 替代 gRPC）

### 5.2 可测试性

- 可以 Mock Domain 接口进行 Controller 测试
- 可以 Mock gRPC 客户端进行 Service 测试

### 5.3 单一职责

- Domain 层：定义业务能力
- Service 层：实现业务逻辑，封装外部调用

### 5.4 易于扩展

添加新方法只需：
1. 在 Domain 接口中声明
2. 在 Service 中实现

---

**Domain & Service 层实现完成**
