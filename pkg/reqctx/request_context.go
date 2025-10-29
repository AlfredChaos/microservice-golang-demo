package reqctx

import "context"

// contextKey 用于在 context 中存储值的键类型
type contextKey string

const (
	// TraceIDKey trace_id 在 context 中的键
	TraceIDKey contextKey = "trace_id"
	// RequestIDKey request_id 在 context 中的键
	RequestIDKey contextKey = "request_id"
	// UserIDKey user_id 在 context 中的键
	UserIDKey contextKey = "user_id"
	// RequestInfoKey 请求信息在 context 中的键
	RequestInfoKey contextKey = "request_info"
)

// RequestInfo 请求信息结构体
type RequestInfo struct {
	Method   string
	Path     string
	ClientIP string
}

// ========== Context 辅助函数 ==========
// 以下函数用于将请求上下文信息存储到 context 中

// WithTraceID 将 trace_id 存储到 context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// WithRequestID 将 request_id 存储到 context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// WithUserID 将 user_id 存储到 context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// WithRequestInfo 将请求信息存储到 context
func WithRequestInfo(ctx context.Context, method, path, clientIP string) context.Context {
	return context.WithValue(ctx, RequestInfoKey, &RequestInfo{
		Method:   method,
		Path:     path,
		ClientIP: clientIP,
	})
}

// GetTraceID 从 context 中获取 trace_id
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// GetRequestID 从 context 中获取 request_id
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// GetUserID 从 context 中获取 user_id
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// GetRequestInfo 从 context 中获取请求信息
func GetRequestInfo(ctx context.Context) *RequestInfo {
	if reqInfo, ok := ctx.Value(RequestInfoKey).(*RequestInfo); ok {
		return reqInfo
	}
	return nil
}
