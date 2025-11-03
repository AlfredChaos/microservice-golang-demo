package httpclient

import (
	"net/http"

	"resty.dev/v3"
)

// RequestOption 请求配置选项
type RequestOption func(*resty.Request)

// WithQueryParams 设置查询参数
func WithQueryParams(params map[string]string) RequestOption {
	return func(req *resty.Request) {
		req.SetQueryParams(params)
	}
}

// WithQueryParam 设置单个查询参数
func WithQueryParam(key, value string) RequestOption {
	return func(req *resty.Request) {
		req.SetQueryParam(key, value)
	}
}

// WithHeader 设置请求头
func WithHeader(key, value string) RequestOption {
	return func(req *resty.Request) {
		req.SetHeader(key, value)
	}
}

// WithHeaders 设置多个请求头
func WithHeaders(headers map[string]string) RequestOption {
	return func(req *resty.Request) {
		req.SetHeaders(headers)
	}
}

// WithAuthToken 设置认证Token
func WithAuthToken(token string) RequestOption {
	return func(req *resty.Request) {
		req.SetAuthToken(token)
	}
}

// WithBasicAuth 设置Basic认证
func WithBasicAuth(username, password string) RequestOption {
	return func(req *resty.Request) {
		req.SetBasicAuth(username, password)
	}
}

// WithBearerToken 设置Bearer Token
func WithBearerToken(token string) RequestOption {
	return func(req *resty.Request) {
		req.SetHeader("Authorization", "Bearer "+token)
	}
}

// WithContentType 设置Content-Type
func WithContentType(contentType string) RequestOption {
	return func(req *resty.Request) {
		req.SetHeader("Content-Type", contentType)
	}
}

// WithError 设置错误响应的结构
func WithError(err interface{}) RequestOption {
	return func(req *resty.Request) {
		req.SetError(err)
	}
}

// WithPathParams 设置路径参数
func WithPathParams(params map[string]string) RequestOption {
	return func(req *resty.Request) {
		req.SetPathParams(params)
	}
}

// WithFormData 设置表单数据
func WithFormData(data map[string]string) RequestOption {
	return func(req *resty.Request) {
		req.SetFormData(data)
	}
}

// WithCookies 设置Cookies
func WithCookies(cookies map[string]string) RequestOption {
	return func(req *resty.Request) {
		for name, value := range cookies {
			req.SetCookie(&http.Cookie{
				Name:  name,
				Value: value,
			})
		}
	}
}

// WithRetry 设置重试次数（请求级别）
func WithRetry(count int) RequestOption {
	return func(req *resty.Request) {
		req.SetRetryCount(count)
	}
}

// WithContext 从context中自动提取trace_id等信息并添加到请求头
// func WithContextHeaders(ctx context.Context) RequestOption {
// 	return func(req *resty.Request) {
// 		// 可以从context中提取trace_id、request_id等信息
// 		// 示例：
// 		// if traceID := reqctx.GetTraceID(ctx); traceID != "" {
// 		// 	req.SetHeader("X-Trace-ID", traceID)
// 		// }
// 	}
// }
