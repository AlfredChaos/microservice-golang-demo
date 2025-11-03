package httpclient

import (
	"fmt"
	"net/http"
)

// HTTPError HTTP请求错误
type HTTPError struct {
	StatusCode int
	Method     string
	URL        string
	Body       []byte
	Message    string
	Err        error
}

// Error 实现error接口
func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("HTTP %s %s failed: %s, status_code=%d, message=%s", 
			e.Method, e.URL, e.Err.Error(), e.StatusCode, e.Message)
	}
	return fmt.Sprintf("HTTP %s %s failed: status_code=%d, message=%s", 
		e.Method, e.URL, e.StatusCode, e.Message)
}

// Unwrap 返回原始错误
func (e *HTTPError) Unwrap() error {
	return e.Err
}

// IsClientError 判断是否为客户端错误（4xx）
func (e *HTTPError) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

// IsServerError 判断是否为服务端错误（5xx）
func (e *HTTPError) IsServerError() bool {
	return e.StatusCode >= 500 && e.StatusCode < 600
}

// IsTimeout 判断是否为超时错误
func IsTimeout(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		if httpErr.Err != nil {
			return IsTimeout(httpErr.Err)
		}
	}
	return false
}

// NewHTTPError 创建HTTP错误
func NewHTTPError(statusCode int, method, url string, body []byte, err error) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Method:     method,
		URL:        url,
		Body:       body,
		Err:        err,
	}
}

// NewHTTPErrorWithMessage 创建带消息的HTTP错误
func NewHTTPErrorWithMessage(statusCode int, method, url, message string, err error) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Method:     method,
		URL:        url,
		Message:    message,
		Err:        err,
	}
}

// IsSuccessStatus 判断状态码是否为成功状态
func IsSuccessStatus(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}
