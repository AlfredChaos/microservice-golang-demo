package dependencies

import (
	userv1 "github.com/alfredchaos/demo/api/user/v1"
	"github.com/alfredchaos/demo/internal/api-gateway/controller"
	"github.com/alfredchaos/demo/internal/api-gateway/service"
	"github.com/alfredchaos/demo/pkg/grpcclient"
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

// AppContext 应用上下文
// 持有所有控制器实例
type AppContext struct {
	UserController controller.IUserController
}

// Dependencies 依赖项
type Dependencies struct {
	ClientManager *grpcclient.Manager
}

// InjectDependencies 依赖注入函数
func InjectDependencies(deps *Dependencies) *AppContext {
	// 获取 gRPC 客户端（使用 GetClient 自动创建类型化客户端）
	userClientRaw, err := deps.ClientManager.GetClient("user-service")
	if err != nil {
		log.Fatal("failed to get user service client", zap.Error(err))
	}
	userClient := userClientRaw.(userv1.UserServiceClient)

	// 创建 Service 层（实现 Domain 接口）
	userService := service.NewUserService(userClient)

	// 创建 Controller 层（依赖 Domain 接口）
	userController := controller.NewUserController(userService)

	return &AppContext{
		UserController: userController,
	}
}
