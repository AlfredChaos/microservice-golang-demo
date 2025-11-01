## Gin依赖注入框架示例 —— 以User为例
### 分层架构
#### 定义IUserService领域模型接口
```
type IUserService interface {
    GetUser(id string) (*User, error)
}
```
#### 定义IUserUseCase应用层接口并实现
实现IUserUseCase接口，也实现了IUserService接口
```
type IUserUseCase interface {
    GetUser(id string) (*User, error)
}

type userUseCase struct {
    userService IUserService
}

func NewUserUseCase(userService IUserService) IUserUseCase {
    return &userUseCase{userService: userService}
}

func (u *userUseCase) GetUser(id string) (*User, error) {
    return u.userService.GetUser(id)
}
```
#### 定义IUserController控制层接口并实现
```
type IUserController interface {
    GetUser(c *gin.Context)
}

type userController struct {
    userService IUserService
}

// 入参使用domain接口
func NewUserController(userService IUserService) IUserController {
    return &userController{userService: userService}
}

func (c *userController) GetUser(ctx *gin.Context) {
    id := ctx.Param("id")
    user, err := c.userUseCase.GetUser(id)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, user)
}
```

### 依赖注入函数
#### 定义一个全局AppContext
```
type AppContext struct {
    userUseCase IUserUseCase
    userController IUserController
}
```
#### 定义依赖注入函数
```
func InjectDependencies() *AppContext {
    userUseCase := NewUserUseCase(userService)
    // useCase实现了domain接口，可以直接注入
    userController := NewUserController(userUseCase)
    return &AppContext{
        userUseCase: userUseCase,
        userController: userController,
    }
}
```
#### 在main函数中使用依赖注入
```
func main() {
    appContext := InjectDependencies()
}
```

### 启动路由
#### 路由分组
```
func main() {
    appContext := InjectDependencies()
    router := gin.Default()
    AppRouter(router, appContext)
    router.Run(":8080")
}

func AppRouter(router *gin.Engine, appContext *AppContext) {
    v1 := router.Group("/v1")
    v1.GET("/user/:id", appContext.userController)
}

func UserRouter(router *gin.RouterGroup, controller IUserController) {
    routerUser := router.Group("/user")
    {
        routerUser.GET("/:id", controller.GetUser)
    }
}
```

