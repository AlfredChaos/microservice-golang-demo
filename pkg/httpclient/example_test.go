package httpclient_test

import (
	"context"
	"fmt"
	"time"

	"github.com/alfredchaos/demo/pkg/httpclient"
)

// User 用户结构
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Example_basicGet 基本GET请求示例
func Example_basicGet() {
	// 创建客户端
	client := httpclient.New(
		httpclient.WithBaseURL("https://jsonplaceholder.typicode.com"),
		httpclient.WithTimeout(10*time.Second),
	)
	defer client.Close()

	// 发送GET请求
	var users []User
	resp, err := client.Get(context.Background(), "/users", &users)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode())
	fmt.Printf("用户数量: %d\n", len(users))
}

// Example_getWithQueryParams GET请求带查询参数
func Example_getWithQueryParams() {
	client := httpclient.New(
		httpclient.WithBaseURL("https://jsonplaceholder.typicode.com"),
	)
	defer client.Close()

	var posts []map[string]interface{}
	_, err := client.Get(
		context.Background(),
		"/posts",
		&posts,
		httpclient.WithQueryParams(map[string]string{
			"userId": "1",
		}),
	)

	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	fmt.Printf("找到 %d 篇文章\n", len(posts))
}

// Example_postRequest POST请求示例
func Example_postRequest() {
	client := httpclient.New(
		httpclient.WithBaseURL("https://jsonplaceholder.typicode.com"),
	)
	defer client.Close()

	newUser := CreateUserRequest{
		Name:  "张三",
		Email: "zhangsan@example.com",
	}

	var result User
	resp, err := client.Post(
		context.Background(),
		"/users",
		newUser,
		&result,
	)

	if err != nil {
		fmt.Printf("创建用户失败: %v\n", err)
		return
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode())
	fmt.Printf("创建的用户: %+v\n", result)
}

// Example_putRequest PUT请求示例
func Example_putRequest() {
	client := httpclient.New(
		httpclient.WithBaseURL("https://jsonplaceholder.typicode.com"),
	)
	defer client.Close()

	updates := map[string]interface{}{
		"name":  "李四",
		"email": "lisi@example.com",
	}

	var result User
	_, err := client.Put(
		context.Background(),
		"/users/1",
		updates,
		&result,
	)

	if err != nil {
		fmt.Printf("更新用户失败: %v\n", err)
		return
	}

	fmt.Printf("更新后的用户: %+v\n", result)
}

// Example_deleteRequest DELETE请求示例
func Example_deleteRequest() {
	client := httpclient.New(
		httpclient.WithBaseURL("https://jsonplaceholder.typicode.com"),
	)
	defer client.Close()

	resp, err := client.Delete(
		context.Background(),
		"/users/1",
		nil,
	)

	if err != nil {
		fmt.Printf("删除用户失败: %v\n", err)
		return
	}

	fmt.Printf("删除成功，状态码: %d\n", resp.StatusCode())
}

// Example_withAuth 带认证的请求
func Example_withAuth() {
	client := httpclient.New(
		httpclient.WithBaseURL("https://api.example.com"),
		httpclient.WithDefaultHeaders(map[string]string{
			"User-Agent": "MyApp/1.0",
		}),
	)
	defer client.Close()

	var result map[string]interface{}
	_, err := client.Get(
		context.Background(),
		"/api/protected/resource",
		&result,
		// 使用Token认证
		httpclient.WithAuthToken("your-api-token"),
		// 或使用Bearer Token
		// httpclient.WithBearerToken("your-bearer-token"),
		// 或使用Basic认证
		// httpclient.WithBasicAuth("username", "password"),
	)

	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	fmt.Printf("受保护的资源: %+v\n", result)
}

// Example_withRetry 带重试的请求
func Example_withRetry() {
	client := httpclient.New(
		httpclient.WithBaseURL("https://api.example.com"),
		httpclient.WithRetryCount(3),
		httpclient.WithRetryWaitTime(1*time.Second),
		httpclient.WithRetryMaxWaitTime(5*time.Second),
	)
	defer client.Close()

	var result map[string]interface{}
	_, err := client.Get(
		context.Background(),
		"/api/unstable/endpoint",
		&result,
	)

	if err != nil {
		fmt.Printf("请求失败（已重试3次）: %v\n", err)
		return
	}

	fmt.Printf("请求成功: %+v\n", result)
}

// Example_pathParams 路径参数示例
func Example_pathParams() {
	client := httpclient.New(
		httpclient.WithBaseURL("https://jsonplaceholder.typicode.com"),
	)
	defer client.Close()

	var post map[string]interface{}
	_, err := client.Get(
		context.Background(),
		"/users/{userId}/posts/{postId}",
		&post,
		httpclient.WithPathParams(map[string]string{
			"userId": "1",
			"postId": "1",
		}),
	)

	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	fmt.Printf("文章: %+v\n", post)
}

// Example_formData 表单数据示例
func Example_formData() {
	client := httpclient.New(
		httpclient.WithBaseURL("https://httpbin.org"),
	)
	defer client.Close()

	var result map[string]interface{}
	_, err := client.Post(
		context.Background(),
		"/post",
		nil,
		&result,
		httpclient.WithFormData(map[string]string{
			"username": "user",
			"password": "pass",
		}),
	)

	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	fmt.Printf("响应: %+v\n", result)
}

// Example_errorHandling 错误处理示例
func Example_errorHandling() {
	client := httpclient.New(
		httpclient.WithBaseURL("https://jsonplaceholder.typicode.com"),
	)
	defer client.Close()

	var result User
	resp, err := client.Get(
		context.Background(),
		"/users/999999", // 不存在的用户
		&result,
	)

	if err != nil {
		// 判断是否为HTTPError
		if httpErr, ok := err.(*httpclient.HTTPError); ok {
			fmt.Printf("HTTP错误: %d\n", httpErr.StatusCode)
			fmt.Printf("错误消息: %s\n", httpErr.Message)

			if httpErr.IsClientError() {
				fmt.Println("这是客户端错误（4xx）")
			}

			if httpErr.IsServerError() {
				fmt.Println("这是服务端错误（5xx）")
			}
		} else {
			fmt.Printf("其他错误: %v\n", err)
		}
		return
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode())
}

// Example_context 使用Context示例
func Example_context() {
	client := httpclient.New(
		httpclient.WithBaseURL("https://jsonplaceholder.typicode.com"),
	)
	defer client.Close()

	// 创建带超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var users []User
	_, err := client.Get(ctx, "/users", &users)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("请求超时")
		} else {
			fmt.Printf("请求失败: %v\n", err)
		}
		return
	}

	fmt.Printf("获取到 %d 个用户\n", len(users))
}
