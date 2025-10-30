### 当前实现的限制

**HTTP (Gin)**：
```go
router.Use(
    middleware.Recovery(),   // 全局应用
    middleware.RequestID(),  // 全局应用
    middleware.Logger(),     // 全局应用
)
```
**问题**：所有中间件都是全局的，无法针对特定接口定制。

**gRPC**：
```go
grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        middleware.UnaryServerRecovery(),  // 全局应用
        middleware.UnaryServerTracing(),   // 全局应用
    ),
)
```
**问题**：gRPC 拦截器本质上是全局的，更难实现细粒度控制。

---

## 💡 解决方案

### 方案 1: HTTP 中间件分组（Gin 路由组）

Gin 天然支持路由组和单路由级别的中间件：

```go
// 1. 全局中间件 A（所有接口）
router := gin.New()
router.Use(
    middleware.Recovery(),      // 中间件 A - 全局
    middleware.RequestID(),     // 中间件 A - 全局
)

// 2. 公开 API 组 - 应用中间件 B
publicAPI := router.Group("/api/v1")
publicAPI.Use(
    middleware.RateLimiter(),   // 中间件 B - 仅公开接口
    middleware.CORS(),          // 中间件 B - 仅公开接口
)
{
    publicAPI.POST("/login", loginHandler)
    publicAPI.POST("/register", registerHandler)
}

// 3. 认证 API 组 - 应用中间件 C
authAPI := router.Group("/api/v1")
authAPI.Use(
    middleware.JWTAuth(),       // 中间件 C - 仅认证接口
    middleware.Permission(),    // 中间件 C - 仅认证接口
)
{
    authAPI.GET("/profile", profileHandler)
    authAPI.POST("/order", createOrderHandler)
}

// 4. 管理员 API 组 - 应用中间件 D
adminAPI := router.Group("/api/v1/admin")
adminAPI.Use(
    middleware.JWTAuth(),       // 中间件 C
    middleware.AdminOnly(),     // 中间件 D - 仅管理员接口
    middleware.AuditLog(),      // 中间件 D - 仅管理员接口
)
{
    adminAPI.GET("/users", listUsersHandler)
    adminAPI.DELETE("/user/:id", deleteUserHandler)
}

// 5. 单个路由特定中间件
router.POST("/upload", 
    middleware.FileSize(10*1024*1024),  // 仅此接口
    uploadHandler,
)
```

---

### 方案 2: gRPC 中间件条件应用

gRPC 拦截器是全局的，需要在**拦截器内部**实现条件判断：

#### 2.1 基于方法名匹配

```go
// pkg/middleware/conditional.go
package middleware

import (
    "context"
    "strings"
    "google.golang.org/grpc"
)

// ConditionalUnaryInterceptor 条件中间件包装器
type ConditionalUnaryInterceptor struct {
    // 匹配规则
    includes []string  // 包含这些路径前缀的方法会应用
    excludes []string  // 排除这些路径前缀的方法
    
    // 实际的拦截器
    interceptor grpc.UnaryServerInterceptor
}

func NewConditionalInterceptor(
    interceptor grpc.UnaryServerInterceptor,
    includes []string,
    excludes []string,
) grpc.UnaryServerInterceptor {
    return func(
        ctx context.Context,
        req interface{},
        info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler,
    ) (interface{}, error) {
        // 检查是否应该应用此拦截器
        shouldApply := false
        
        // 检查 includes
        if len(includes) == 0 {
            shouldApply = true  // 空则应用到所有
        } else {
            for _, prefix := range includes {
                if strings.HasPrefix(info.FullMethod, prefix) {
                    shouldApply = true
                    break
                }
            }
        }
        
        // 检查 excludes
        for _, prefix := range excludes {
            if strings.HasPrefix(info.FullMethod, prefix) {
                shouldApply = false
                break
            }
        }
        
        // 应用或跳过拦截器
        if shouldApply {
            return interceptor(ctx, req, info, handler)
        }
        
        // 直接调用处理函数
        return handler(ctx, req)
    }
}

// 使用示例
func NewGRPCServer() *grpc.Server {
    return grpc.NewServer(
        grpc.ChainUnaryInterceptor(
            // 中间件 A - 全局应用
            middleware.UnaryServerRecovery(),
            middleware.UnaryServerTracing(),
            
            // 中间件 B - 仅应用到 UserService
            NewConditionalInterceptor(
                middleware.UnaryServerRateLimiter(),
                []string{"/user.v1.UserService/"},  // includes
                []string{},                          // excludes
            ),
            
            // 中间件 C - 仅应用到需要认证的方法
            NewConditionalInterceptor(
                middleware.UnaryServerAuth(),
                []string{
                    "/user.v1.UserService/GetProfile",
                    "/order.v1.OrderService/",
                },
                []string{
                    "/user.v1.UserService/Login",
                    "/user.v1.UserService/Register",
                },
            ),
            
            // 日志记录 - 全局应用
            middleware.UnaryServerLogging(),
        ),
    )
}
```

#### 2.2 基于元数据（Metadata）

```go
// 基于 metadata 标记的条件中间件
func MetadataConditionalInterceptor(
    interceptor grpc.UnaryServerInterceptor,
    metadataKey string,
    expectedValue string,
) grpc.UnaryServerInterceptor {
    return func(
        ctx context.Context,
        req interface{},
        info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler,
    ) (interface{}, error) {
        // 从 metadata 中读取标记
        md, ok := metadata.FromIncomingContext(ctx)
        if ok {
            values := md.Get(metadataKey)
            if len(values) > 0 && values[0] == expectedValue {
                // 应用拦截器
                return interceptor(ctx, req, info, handler)
            }
        }
        
        // 跳过拦截器
        return handler(ctx, req)
    }
}
```

---

### 方案 3: 配置驱动的中间件

```go
// middleware_config.yaml
middleware:
  global:
    - recovery
    - tracing
    - logging
  
  groups:
    public_api:
      paths:
        - /user.v1.UserService/Login
        - /user.v1.UserService/Register
      middleware:
        - rate_limiter
        - cors
    
    auth_api:
      paths:
        - /user.v1.UserService/GetProfile
        - /order.v1.OrderService/*
      middleware:
        - jwt_auth
        - permission
    
    admin_api:
      paths:
        - /admin.v1.AdminService/*
      middleware:
        - jwt_auth
        - admin_check
        - audit_log
```

```go
// 配置驱动的拦截器加载
type MiddlewareConfig struct {
    Global []string `yaml:"global"`
    Groups map[string]GroupConfig `yaml:"groups"`
}

type GroupConfig struct {
    Paths      []string `yaml:"paths"`
    Middleware []string `yaml:"middleware"`
}

func LoadMiddleware(config MiddlewareConfig) []grpc.UnaryServerInterceptor {
    var interceptors []grpc.UnaryServerInterceptor
    
    // 加载全局中间件
    for _, name := range config.Global {
        interceptors = append(interceptors, getInterceptor(name))
    }
    
    // 加载分组中间件
    for _, group := range config.Groups {
        for _, middlewareName := range group.Middleware {
            wrapped := NewConditionalInterceptor(
                getInterceptor(middlewareName),
                group.Paths,
                []string{},
            )
            interceptors = append(interceptors, wrapped)
        }
    }
    
    return interceptors
}
```

---

## 📊 方案对比

| 方案 | HTTP支持 | gRPC支持 | 灵活性 | 复杂度 | 推荐度 |
|------|---------|---------|-------|--------|--------|
| **路由组** | ✅ 原生 | ❌ 不支持 | ⭐⭐⭐⭐⭐ | ⭐ | ⭐⭐⭐⭐⭐ |
| **条件拦截器** | ✅ 可用 | ✅ 需要封装 | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ |
| **配置驱动** | ✅ 可用 | ✅ 可用 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Metadata标记** | ❌ 不适用 | ✅ 可用 | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ |

---

## 🎯 推荐实践

### HTTP 中间件
```go
// internal/api-gateway/router/router.go
func SetupRouter() *gin.Engine {
    router := gin.New()
    
    // 全局中间件（中间件 A）
    router.Use(
        middleware.Recovery(),
        middleware.RequestID(),
        middleware.Logger(),
    )
    
    // 公开 API（中间件 B）
    public := router.Group("/api/v1")
    public.Use(middleware.RateLimiter())
    {
        public.POST("/login", loginHandler)
    }
    
    // 认证 API（中间件 C）
    auth := router.Group("/api/v1")
    auth.Use(middleware.JWTAuth())
    {
        auth.GET("/profile", profileHandler)
    }
    
    return router
}
```

### gRPC 中间件
```go
// pkg/middleware/conditional.go - 创建此文件
// 实现条件拦截器包装器

// internal/user-service/server/grpc.go
func NewGRPCServer() *grpc.Server {
    return grpc.NewServer(
        grpc.ChainUnaryInterceptor(
            // 全局（中间件 A）
            middleware.UnaryServerRecovery(),
            middleware.UnaryServerTracing(),
            
            // 条件应用（中间件 B, C）
            middleware.NewConditionalInterceptor(
                middleware.UnaryServerAuth(),
                []string{"/user.v1.UserService/GetProfile"},
                []string{"/user.v1.UserService/Login"},
            ),
            
            // 全局日志
            middleware.UnaryServerLogging(),
        ),
    )
}
```

---

## 🤔 你需要决定

1. **HTTP 中间件**：直接使用 Gin 的路由组即可，非常简单
2. **gRPC 中间件**：需要实现条件拦截器包装器