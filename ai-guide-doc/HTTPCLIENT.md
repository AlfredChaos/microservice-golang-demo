# HTTP Client 包

基于 go-resty v3 封装的 HTTP 客户端包，提供统一的日志记录、错误处理、请求延迟监控、重试机制等功能。

## 特性

- ✅ 封装 go-resty v3，提供简洁的 API
- ✅ 统一的日志记录（使用项目的 pkg/log）
- ✅ 自动记录请求延迟
- ✅ 慢请求警告（可配置阈值）
- ✅ 错误处理和重试机制
- ✅ 支持多种认证方式（Token、Basic Auth、Bearer Token）
- ✅ 灵活的配置选项
- ✅ 请求级别和客户端级别配置

## 安装

```bash
go get resty.dev/v3
```

## 快速开始

### 基本用法

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/alfredchaos/demo/pkg/httpclient"
)

func main() {
    // 创建客户端
    client := httpclient.New(
        httpclient.WithBaseURL("https://api.example.com"),
        httpclient.WithTimeout(10 * time.Second),
        httpclient.WithRetryCount(3),
    )
    defer client.Close()
    
    // 发送 GET 请求
    var result map[string]interface{}
    resp, err := client.Get(context.Background(), "/api/users", &result)
    if err != nil {
        fmt.Printf("请求失败: %v\n", err)
        return
    }
    
    fmt.Printf("状态码: %d\n", resp.StatusCode())
    fmt.Printf("结果: %+v\n", result)
}
```

### POST 请求

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

type CreateUserResponse struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func createUser() {
    client := httpclient.New(
        httpclient.WithBaseURL("https://api.example.com"),
    )
    defer client.Close()
    
    user := User{
        Name:  "张三",
        Email: "zhangsan@example.com",
    }
    
    var result CreateUserResponse
    resp, err := client.Post(
        context.Background(),
        "/api/users",
        user,
        &result,
        httpclient.WithAuthToken("your-token"),
    )
    
    if err != nil {
        fmt.Printf("创建用户失败: %v\n", err)
        return
    }
    
    fmt.Printf("用户创建成功: ID=%d, Name=%s\n", result.ID, result.Name)
}
```

### PUT 请求

```go
func updateUser(userID int) {
    client := httpclient.New(
        httpclient.WithBaseURL("https://api.example.com"),
    )
    defer client.Close()
    
    updates := map[string]interface{}{
        "name": "李四",
        "age":  25,
    }
    
    var result User
    _, err := client.Put(
        context.Background(),
        fmt.Sprintf("/api/users/%d", userID),
        updates,
        &result,
    )
    
    if err != nil {
        fmt.Printf("更新用户失败: %v\n", err)
        return
    }
}
```

### DELETE 请求

```go
func deleteUser(userID int) {
    client := httpclient.New(
        httpclient.WithBaseURL("https://api.example.com"),
    )
    defer client.Close()
    
    _, err := client.Delete(
        context.Background(),
        fmt.Sprintf("/api/users/%d", userID),
        nil,
        httpclient.WithAuthToken("your-token"),
    )
    
    if err != nil {
        fmt.Printf("删除用户失败: %v\n", err)
        return
    }
    
    fmt.Println("用户删除成功")
}
```

## 配置选项

### 客户端级别配置

```go
client := httpclient.New(
    // 设置基础URL
    httpclient.WithBaseURL("https://api.example.com"),
    
    // 设置超时时间
    httpclient.WithTimeout(30 * time.Second),
    
    // 设置重试次数
    httpclient.WithRetryCount(3),
    
    // 设置重试等待时间
    httpclient.WithRetryWaitTime(1 * time.Second),
    
    // 设置最大重试等待时间
    httpclient.WithRetryMaxWaitTime(5 * time.Second),
    
    // 设置默认请求头（客户端级别）
    httpclient.WithDefaultHeaders(map[string]string{
        "User-Agent": "MyApp/1.0",
        "Accept":     "application/json",
    }),
    
    // 设置调试模式
    httpclient.WithDebug(true),
    
    // 设置慢请求阈值（超过此时间会记录警告日志）
    httpclient.WithLogSlowThreshold(3 * time.Second),
)
```

### 请求级别配置

```go
// 设置查询参数
client.Get(ctx, "/api/users", &result,
    httpclient.WithQueryParams(map[string]string{
        "page":     "1",
        "per_page": "20",
    }),
)

// 设置请求头
client.Post(ctx, "/api/users", user, &result,
    httpclient.WithHeader("X-Custom-Header", "value"),
)

// 设置认证
client.Get(ctx, "/api/protected", &result,
    // Token认证
    httpclient.WithAuthToken("your-token"),
    
    // 或 Basic认证
    // httpclient.WithBasicAuth("username", "password"),
    
    // 或 Bearer Token
    // httpclient.WithBearerToken("your-token"),
)

// 设置路径参数
client.Get(ctx, "/api/users/{id}/posts/{postId}", &result,
    httpclient.WithPathParams(map[string]string{
        "id":     "123",
        "postId": "456",
    }),
)

// 设置表单数据
client.Post(ctx, "/api/login", nil, &result,
    httpclient.WithFormData(map[string]string{
        "username": "user",
        "password": "pass",
    }),
)

// 设置Cookies
client.Get(ctx, "/api/data", &result,
    httpclient.WithCookies(map[string]string{
        "session": "session-token",
    }),
)

// 请求级别重试
client.Get(ctx, "/api/data", &result,
    httpclient.WithRetry(5), // 这个请求重试5次
)
```

## 日志记录

客户端会自动记录以下信息：

- **请求开始**: 记录请求的 method 和 url
- **请求完成**: 记录请求的 method、url、status_code 和 duration_ms
- **慢请求警告**: 当请求时间超过配置的阈值时，会记录警告日志
- **请求失败**: 记录错误信息

日志示例：

```json
{
  "timestamp": "2024-11-03 11:25:45.123",
  "level": "info",
  "message": "HTTP请求开始",
  "method": "GET",
  "url": "https://api.example.com/api/users"
}

{
  "timestamp": "2024-11-03 11:25:45.456",
  "level": "info",
  "message": "HTTP请求完成",
  "method": "GET",
  "url": "https://api.example.com/api/users",
  "status_code": 200,
  "duration_ms": 333
}

{
  "timestamp": "2024-11-03 11:25:50.789",
  "level": "warn",
  "message": "HTTP慢请求",
  "method": "POST",
  "url": "https://api.example.com/api/process",
  "status_code": 200,
  "duration_ms": 3500
}
```

## 错误处理

### 使用自定义错误类型

```go
resp, err := client.Get(ctx, "/api/users", &result)
if err != nil {
    // 判断是否为HTTPError
    if httpErr, ok := err.(*httpclient.HTTPError); ok {
        fmt.Printf("HTTP错误: 状态码=%d\n", httpErr.StatusCode)
        fmt.Printf("错误消息: %s\n", httpErr.Message)
        
        // 判断是否为客户端错误（4xx）
        if httpErr.IsClientError() {
            fmt.Println("客户端错误")
        }
        
        // 判断是否为服务端错误（5xx）
        if httpErr.IsServerError() {
            fmt.Println("服务端错误")
        }
    }
}
```

### 设置错误响应结构

```go
type ErrorResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

var result UserResponse
var errResp ErrorResponse

resp, err := client.Get(ctx, "/api/users", &result,
    httpclient.WithError(&errResp),
)

if err != nil {
    fmt.Printf("错误响应: %+v\n", errResp)
}
```

## 高级用法

### 获取底层 resty 客户端

如果需要使用 resty 的高级特性，可以获取底层客户端：

```go
client := httpclient.New()
restyClient := client.GetRestyClient()

// 使用 resty 的高级特性
restyClient.SetRetryAfter(func(c *resty.Client, r *resty.Response) (time.Duration, error) {
    return 0, nil
})
```

## 认证授权

目前认证功能通过请求选项实现，支持：

- Token 认证: `WithAuthToken("token")`
- Basic 认证: `WithBasicAuth("username", "password")`
- Bearer Token: `WithBearerToken("token")`

客户端预留了 `applyAuth` 方法用于统一认证处理，可以根据需要扩展实现：

```go
// 在 client.go 中扩展 applyAuth 方法
func (c *Client) applyAuth(req *resty.Request) error {
    // 从配置或其他地方获取token
    token := getTokenFromSomewhere()
    if token != "" {
        req.SetAuthToken(token)
    }
    return nil
}
```

## 注意事项

1. **记得关闭客户端**: 使用 `defer client.Close()` 确保资源被正确释放
2. **慢请求阈值**: 默认为3秒，可以根据业务需求调整
3. **重试机制**: 默认重试3次，使用指数退避策略
4. **日志依赖**: 依赖项目的 `pkg/log` 包，确保日志系统已初始化

## 与原有 http_client 的区别

| 特性 | 原 http_client | 新 httpclient |
|------|---------------|---------------|
| 底层实现 | net/http | go-resty v3 |
| 日志库 | logrus | zap (pkg/log) |
| API 风格 | 需要构建 Request 结构 | 直接调用方法 |
| 配置方式 | 选项模式 | 选项模式 |
| 中间件 | 自定义实现 | resty 内置 |
| 性能 | 较好 | 更好 |

## 示例项目

查看 `example_test.go` 了解更多使用示例。

## 许可证

MIT License
