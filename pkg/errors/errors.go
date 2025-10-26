package errors

import (
	"fmt"
)

// ErrorCode 错误码类型
type ErrorCode int

const (
	// Success 成功
	Success ErrorCode = 0
	
	// ErrInternalServer 内部服务器错误
	ErrInternalServer ErrorCode = 10001
	
	// ErrInvalidParams 参数错误
	ErrInvalidParams ErrorCode = 10002
	
	// ErrNotFound 资源不存在
	ErrNotFound ErrorCode = 10003
	
	// ErrUnauthorized 未授权
	ErrUnauthorized ErrorCode = 10004
	
	// ErrForbidden 禁止访问
	ErrForbidden ErrorCode = 10005
	
	// ErrServiceUnavailable 服务不可用
	ErrServiceUnavailable ErrorCode = 10006
	
	// ErrTimeout 请求超时
	ErrTimeout ErrorCode = 10007
	
	// ErrDatabaseError 数据库错误
	ErrDatabaseError ErrorCode = 20001
	
	// ErrCacheError 缓存错误
	ErrCacheError ErrorCode = 20002
	
	// ErrMessageQueueError 消息队列错误
	ErrMessageQueueError ErrorCode = 20003
	
	// ErrRPCError RPC调用错误
	ErrRPCError ErrorCode = 30001
)

// AppError 应用错误结构
type AppError struct {
	Code    ErrorCode // 错误码
	Message string    // 错误消息
	Err     error     // 原始错误
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 返回原始错误,支持 errors.Unwrap
func (e *AppError) Unwrap() error {
	return e.Err
}

// New 创建新的应用错误
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap 包装已有错误
func Wrap(code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// GetErrorMessage 获取错误码对应的默认消息
func GetErrorMessage(code ErrorCode) string {
	messages := map[ErrorCode]string{
		Success:               "success",
		ErrInternalServer:     "internal server error",
		ErrInvalidParams:      "invalid parameters",
		ErrNotFound:           "resource not found",
		ErrUnauthorized:       "unauthorized",
		ErrForbidden:          "forbidden",
		ErrServiceUnavailable: "service unavailable",
		ErrTimeout:            "request timeout",
		ErrDatabaseError:      "database error",
		ErrCacheError:         "cache error",
		ErrMessageQueueError:  "message queue error",
		ErrRPCError:           "rpc call error",
	}
	
	if msg, ok := messages[code]; ok {
		return msg
	}
	return "unknown error"
}

// IsAppError 判断是否为应用错误
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError 从 error 中提取 AppError
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return nil
}
