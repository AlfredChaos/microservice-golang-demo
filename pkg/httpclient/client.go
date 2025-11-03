package httpclient

import (
	"context"
	"fmt"
	"time"

	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
	"resty.dev/v3"
)

// Client HTTP客户端封装
type Client struct {
	client *resty.Client
	config *Config
}

// New 创建HTTP客户端
func New(options ...Option) *Client {
	// 创建默认配置
	cfg := DefaultConfig()
	
	// 应用配置选项
	for _, opt := range options {
		opt(cfg)
	}
	
	// 创建 resty 客户端
	restyClient := resty.New()
	
	// 设置基础URL
	if cfg.BaseURL != "" {
		restyClient.SetBaseURL(cfg.BaseURL)
	}
	
	// 设置超时
	restyClient.SetTimeout(cfg.Timeout)
	
	// 设置重试
	if cfg.RetryCount > 0 {
		restyClient.
			SetRetryCount(cfg.RetryCount).
			SetRetryWaitTime(cfg.RetryWaitTime).
			SetRetryMaxWaitTime(cfg.RetryMaxWaitTime)
	}
	
	// 设置默认请求头
	if len(cfg.Headers) > 0 {
		restyClient.SetHeaders(cfg.Headers)
	}
	
	// 设置调试模式
	if cfg.Debug {
		restyClient.SetDebug(true)
	}
	
	c := &Client{
		client: restyClient,
		config: cfg,
	}
	
	// 添加请求中间件
	c.setupMiddlewares()
	
	return c
}

// setupMiddlewares 设置中间件
func (c *Client) setupMiddlewares() {
	// 请求前的日志和延迟记录中间件
	c.client.AddRequestMiddleware(func(client *resty.Client, req *resty.Request) error {
		// 记录请求开始时间
		req.SetContext(context.WithValue(req.Context(), requestStartTimeKey, time.Now()))
		
		// 记录请求日志
		if log.Logger != nil {
			log.Info("HTTP请求开始",
				zap.String("method", req.Method),
				zap.String("url", req.URL),
			)
		}
		
		return nil
	})
	
	// 响应后的日志和延迟记录中间件
	c.client.AddResponseMiddleware(func(client *resty.Client, resp *resty.Response) error {
		// 计算请求延迟
		startTime, ok := resp.Request.Context().Value(requestStartTimeKey).(time.Time)
		if !ok {
			startTime = time.Now()
		}
		duration := time.Since(startTime)
		
		// 记录响应日志
		if log.Logger != nil {
			fields := []zap.Field{
				zap.String("method", resp.Request.Method),
				zap.String("url", resp.Request.URL),
				zap.Int("status_code", resp.StatusCode()),
				zap.Int64("duration_ms", duration.Milliseconds()),
			}
			
			// 如果请求时间超过阈值，记录警告
			if duration > c.config.LogSlowThreshold {
				log.Warn("HTTP慢请求", fields...)
			} else {
				log.Info("HTTP请求完成", fields...)
			}
			
			// 错误处理
			if resp.Err != nil {
				log.Error("HTTP请求失败",
					zap.String("method", resp.Request.Method),
					zap.String("url", resp.Request.URL),
					zap.Error(resp.Err),
				)
			}
		}
		
		return nil
	})
}

// Get 发送GET请求
func (c *Client) Get(ctx context.Context, url string, result interface{}, options ...RequestOption) (*resty.Response, error) {
	return c.doRequest(ctx, resty.MethodGet, url, nil, result, options...)
}

// Post 发送POST请求
func (c *Client) Post(ctx context.Context, url string, body interface{}, result interface{}, options ...RequestOption) (*resty.Response, error) {
	return c.doRequest(ctx, resty.MethodPost, url, body, result, options...)
}

// Put 发送PUT请求
func (c *Client) Put(ctx context.Context, url string, body interface{}, result interface{}, options ...RequestOption) (*resty.Response, error) {
	return c.doRequest(ctx, resty.MethodPut, url, body, result, options...)
}

// Delete 发送DELETE请求
func (c *Client) Delete(ctx context.Context, url string, result interface{}, options ...RequestOption) (*resty.Response, error) {
	return c.doRequest(ctx, resty.MethodDelete, url, nil, result, options...)
}

// Patch 发送PATCH请求
func (c *Client) Patch(ctx context.Context, url string, body interface{}, result interface{}, options ...RequestOption) (*resty.Response, error) {
	return c.doRequest(ctx, resty.MethodPatch, url, body, result, options...)
}

// doRequest 执行HTTP请求
func (c *Client) doRequest(ctx context.Context, method, url string, body, result interface{}, options ...RequestOption) (*resty.Response, error) {
	// 创建请求
	req := c.client.R()
	
	// 设置上下文
	if ctx != nil {
		req.SetContext(ctx)
	}
	
	// 设置请求体
	if body != nil {
		req.SetBody(body)
	}
	
	// 设置响应结果
	if result != nil {
		req.SetResult(result)
	}
	
	// 应用请求选项
	for _, opt := range options {
		opt(req)
	}
	
	// 执行认证（预留接口，暂不实现）
	if err := c.applyAuth(req); err != nil {
		return nil, err
	}
	
	// 执行请求
	var resp *resty.Response
	var err error
	
	switch method {
	case resty.MethodGet:
		resp, err = req.Get(url)
	case resty.MethodPost:
		resp, err = req.Post(url)
	case resty.MethodPut:
		resp, err = req.Put(url)
	case resty.MethodDelete:
		resp, err = req.Delete(url)
	case resty.MethodPatch:
		resp, err = req.Patch(url)
	default:
		return nil, fmt.Errorf("不支持的HTTP方法: %s", method)
	}
	
	if err != nil {
		return nil, err
	}
	
	// 检查响应状态
	if !IsSuccessStatus(resp.StatusCode()) {
		return resp, NewHTTPErrorWithMessage(
			resp.StatusCode(),
			method,
			url,
			resp.String(),
			nil,
		)
	}
	
	return resp, nil
}

// applyAuth 应用认证（预留接口，暂不实现）
func (c *Client) applyAuth(req *resty.Request) error {
	// TODO: 实现认证逻辑
	// 可以从配置中读取认证信息，或者通过其他方式获取token
	// 示例：
	// req.SetAuthToken("your-token")
	// req.SetBasicAuth("username", "password")
	return nil
}

// Close 关闭客户端
func (c *Client) Close() error {
	if c.client != nil {
		c.client.Close()
	}
	return nil
}

// GetRestyClient 获取底层resty客户端（用于高级用法）
func (c *Client) GetRestyClient() *resty.Client {
	return c.client
}

// requestStartTimeKey 请求开始时间的context key
type contextKey string

const requestStartTimeKey contextKey = "request_start_time"
