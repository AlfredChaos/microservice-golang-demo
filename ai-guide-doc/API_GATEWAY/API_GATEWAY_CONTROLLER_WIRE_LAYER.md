# API Gateway - Controller & Wire 层实现

> 控制器和依赖注入层的完整代码

## 1. Controller 层 - 控制器实现

### 1.1 用户控制器

#### 文件路径
`internal/api-gateway/controller/user_controller.go`

#### 完整代码

```go
package controller

import (
	"net/http"

	"github.com/alfredchaos/demo/internal/api-gateway/domain"
	"github.com/alfredchaos/demo/internal/api-gateway/dto"
	"github.com/alfredchaos/demo/pkg/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// IUserController 用户控制器接口
type IUserController interface {
	SayHello(c *gin.Context)
}

// userController 用户控制器实现
type userController struct {
	userService domain.IUserService
}

// NewUserController 创建用户控制器
// 依赖领域服务接口
func NewUserController(userService domain.IUserService) IUserController {
	return &userController{
		userService: userService,
	}
}

// SayHello 处理问候请求
// @Summary 问候接口
// @Description 调用 user-service 并返回问候语
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=dto.HelloResponse} "成功响应"
// @Failure 500 {object} dto.Response "服务器错误"
// @Router /api/v1/user/hello [get]
func (ctrl *userController) SayHello(c *gin.Context) {
	ctx := c.Request.Context()
	
	// 使用 WithContext 自动附加请求上下文信息
	log.WithContext(ctx).Info("received user hello request")

	// 调用用户服务
	message, err := ctrl.userService.SayHello(ctx)
	if err != nil {
		log.WithContext(ctx).Error("failed to call user service", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(10001, "failed to call user service"))
		return
	}

	log.WithContext(ctx).Info("user hello request completed", zap.String("message", message))

	// 返回响应
	c.JSON(http.StatusOK, dto.NewSuccessResponse(dto.HelloResponse{
		Message: message,
	}))
}
```

---

### 1.2 图书控制器

#### 文件路径
`internal/api-gateway/controller/book_controller.go`

#### 完整代码

```go
package controller

import (
	"net/http"

	"github.com/alfredchaos/demo/internal/api-gateway/domain"
	"github.com/alfredchaos/demo/internal/api-gateway/dto"
	"github.com/alfredchaos/demo/pkg/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// IBookController 图书控制器接口
type IBookController interface {
	GetBook(c *gin.Context)
}

// bookController 图书控制器实现
type bookController struct {
	bookService domain.IBookService
}

// NewBookController 创建图书控制器
// 依赖领域服务接口
func NewBookController(bookService domain.IBookService) IBookController {
	return &bookController{
		bookService: bookService,
	}
}

// GetBook 处理获取图书请求
// @Summary 获取图书
// @Description 调用 book-service 获取图书信息
// @Tags Book
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=string} "成功响应"
// @Failure 500 {object} dto.Response "服务器错误"
// @Router /api/v1/book [get]
func (ctrl *bookController) GetBook(c *gin.Context) {
	ctx := c.Request.Context()

	log.WithContext(ctx).Info("received get book request")

	// 调用图书服务
	message, err := ctrl.bookService.GetBook(ctx)
	if err != nil {
		log.WithContext(ctx).Error("failed to call book service", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(10002, "failed to call book service"))
		return
	}

	log.WithContext(ctx).Info("get book request completed", zap.String("message", message))

	// 返回响应
	c.JSON(http.StatusOK, dto.NewSuccessResponse(message))
}
```

---

### Controller 层设计说明

1. **定义接口**：每个控制器都定义接口 `IXxxController`
2. **依赖注入**：通过构造函数注入 Domain 接口
3. **统一响应**：使用 `dto.Response` 统一响应格式
4. **错误处理**：记录错误日志并返回友好错误信息
5. **Swagger 注解**：添加 API 文档注解

---

## 2. Wire 层 - 依赖注入

### 文件路径
`internal/api-gateway/wire/wire.go`

### 完整代码

```go
package wire

import (
	orderv1 "github.com/alfredchaos/demo/api/order/v1"
	userv1 "github.com/alfredchaos/demo/api/user/v1"
	"github.com/alfredchaos/demo/internal/api-gateway/client"
	"github.com/alfredchaos/demo/internal/api-gateway/controller"
	"github.com/alfredchaos/demo/internal/api-gateway/service"
	"github.com/alfredchaos/demo/pkg/log"
	"github.com/alfredchaos/demo/pkg/mq"
	"go.uber.org/zap"
)

// AppContext 应用上下文
// 持有所有控制器实例
type AppContext struct {
	UserController controller.IUserController
	BookController controller.IBookController
	// 可以继续添加更多控制器
	// OrderController controller.IOrderController
}

// Dependencies 依赖项
// 封装外部依赖（gRPC 连接管理器、MQ 等）
type Dependencies struct {
	ConnManager *client.ConnectionManager
	MQPublisher mq.Publisher
	// 可以添加更多外部依赖
	// RedisClient *redis.Client
	// MongoClient *mongo.Client
}

// InjectDependencies 依赖注入函数
// 按照依赖关系组装各层组件
func InjectDependencies(deps *Dependencies) *AppContext {
	// 创建客户端工厂
	clientFactory := client.NewClientFactory(deps.ConnManager)

	// 创建各服务的 gRPC 客户端
	userClient, err := clientFactory.CreateUserClient()
	if err != nil {
		log.Fatal("failed to create user client", zap.Error(err))
	}

	bookClient, err := clientFactory.CreateBookClient()
	if err != nil {
		log.Fatal("failed to create book client", zap.Error(err))
	}

	// 创建 Service 层（实现 Domain 接口）
	userService := service.NewUserService(userClient)
	bookService := service.NewBookService(bookClient)

	// 创建 Controller 层（依赖 Domain 接口）
	userController := controller.NewUserController(userService)
	bookController := controller.NewBookController(bookService)

	return &AppContext{
		UserController: userController,
		BookController: bookController,
	}
}

// 类型别名（可选，用于简化代码）
type (
	UserClient userv1.UserServiceClient
	BookClient orderv1.BookServiceClient
)
```

---

### Wire 层设计说明

1. **AppContext**：持有所有控制器实例，供路由使用
2. **Dependencies**：封装外部依赖，便于测试和替换
3. **InjectDependencies**：按依赖关系组装组件
4. **统一管理**：所有依赖关系在此集中管理

---

## 3. 依赖注入流程图

```
Dependencies (外部依赖)
    │
    ├─ ConnectionManager ────┐
    └─ MQPublisher            │
                              │
                              ▼
                    ClientFactory
                              │
                ┌─────────────┼─────────────┐
                ▼             ▼             ▼
           UserClient    BookClient    OrderClient
                │             │             │
                ▼             ▼             ▼
           UserService   BookService   OrderService
           (实现Domain)   (实现Domain)   (实现Domain)
                │             │             │
                ▼             ▼             ▼
         UserController  BookController OrderController
                │             │             │
                └─────────────┼─────────────┘
                              ▼
                         AppContext
                              │
                              ▼
                          Router
```

---

## 4. 扩展示例：添加订单服务

### 4.1 Domain 接口

```go
// domain/order_service.go
package domain

import "context"

type IOrderService interface {
    CreateOrder(ctx context.Context, userID, productID string) (*Order, error)
    GetOrder(ctx context.Context, orderID string) (*Order, error)
}

type Order struct {
    ID        string
    UserID    string
    ProductID string
    Status    string
}
```

### 4.2 Service 实现

```go
// service/order_service.go
package service

import (
    "context"
    orderv1 "github.com/alfredchaos/demo/api/order/v1"
    "github.com/alfredchaos/demo/internal/api-gateway/domain"
)

type orderService struct {
    baseService
    orderClient orderv1.OrderServiceClient
}

func NewOrderService(client orderv1.OrderServiceClient) domain.IOrderService {
    return &orderService{orderClient: client}
}

func (s *orderService) CreateOrder(ctx context.Context, userID, productID string) (*domain.Order, error) {
    ctx = s.withTraceID(ctx)
    
    resp, err := s.orderClient.CreateOrder(ctx, &orderv1.CreateOrderRequest{
        UserId:    userID,
        ProductId: productID,
    })
    if err != nil {
        return nil, err
    }
    
    return &domain.Order{
        ID:        resp.OrderId,
        UserID:    resp.UserId,
        ProductID: resp.ProductId,
        Status:    resp.Status,
    }, nil
}
```

### 4.3 Controller 实现

```go
// controller/order_controller.go
package controller

import (
    "net/http"
    "github.com/alfredchaos/demo/internal/api-gateway/domain"
    "github.com/gin-gonic/gin"
)

type IOrderController interface {
    CreateOrder(c *gin.Context)
}

type orderController struct {
    orderService domain.IOrderService
}

func NewOrderController(service domain.IOrderService) IOrderController {
    return &orderController{orderService: service}
}

func (ctrl *orderController) CreateOrder(c *gin.Context) {
    var req struct {
        UserID    string `json:"user_id" binding:"required"`
        ProductID string `json:"product_id" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    order, err := ctrl.orderService.CreateOrder(c.Request.Context(), req.UserID, req.ProductID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, order)
}
```

### 4.4 更新 Wire

```go
// wire/wire.go
type AppContext struct {
    UserController  controller.IUserController
    BookController  controller.IBookController
    OrderController controller.IOrderController  // 新增
}

func InjectDependencies(deps *Dependencies) *AppContext {
    clientFactory := client.NewClientFactory(deps.ConnManager)
    
    userClient, _ := clientFactory.CreateUserClient()
    bookClient, _ := clientFactory.CreateBookClient()
    orderClient, _ := clientFactory.CreateOrderClient()  // 新增
    
    userService := service.NewUserService(userClient)
    bookService := service.NewBookService(bookClient)
    orderService := service.NewOrderService(orderClient)  // 新增
    
    userController := controller.NewUserController(userService)
    bookController := controller.NewBookController(bookService)
    orderController := controller.NewOrderController(orderService)  // 新增
    
    return &AppContext{
        UserController:  userController,
        BookController:  bookController,
        OrderController: orderController,  // 新增
    }
}
```

---

## 5. 测试示例

### 5.1 Controller 单元测试

```go
package controller

import (
    "context"
    "errors"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/alfredchaos/demo/internal/api-gateway/domain"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockUserService Mock 用户服务
type MockUserService struct {
    mock.Mock
}

func (m *MockUserService) SayHello(ctx context.Context) (string, error) {
    args := m.Called(ctx)
    return args.String(0), args.Error(1)
}

func TestUserController_SayHello_Success(t *testing.T) {
    // 创建 mock 服务
    mockService := new(MockUserService)
    mockService.On("SayHello", mock.Anything).Return("Hello World", nil)
    
    // 创建控制器
    controller := NewUserController(mockService)
    
    // 设置 Gin 测试环境
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("GET", "/api/v1/user/hello", nil)
    
    // 调用控制器方法
    controller.SayHello(c)
    
    // 断言结果
    assert.Equal(t, http.StatusOK, w.Code)
    mockService.AssertExpectations(t)
}

func TestUserController_SayHello_Error(t *testing.T) {
    // 创建 mock 服务（返回错误）
    mockService := new(MockUserService)
    mockService.On("SayHello", mock.Anything).Return("", errors.New("service error"))
    
    // 创建控制器
    controller := NewUserController(mockService)
    
    // 设置 Gin 测试环境
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("GET", "/api/v1/user/hello", nil)
    
    // 调用控制器方法
    controller.SayHello(c)
    
    // 断言结果
    assert.Equal(t, http.StatusInternalServerError, w.Code)
    mockService.AssertExpectations(t)
}
```

---

## 6. 设计优势

### 6.1 解耦性

- Controller 依赖 Domain 接口，不依赖具体实现
- 易于 Mock，测试友好

### 6.2 可维护性

- 依赖关系集中在 Wire 层管理
- 修改依赖只需修改一处

### 6.3 可扩展性

- 添加新控制器只需在 `AppContext` 中添加
- 添加新依赖只需在 `Dependencies` 中添加

### 6.4 类型安全

- 使用 Go 的类型系统保证编译期检查
- 避免运行时依赖注入错误

---

**Controller & Wire 层实现完成**
