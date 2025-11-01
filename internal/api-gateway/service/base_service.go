package service

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// baseService 基础服务
// 提供公共方法供其他服务使用
type baseService struct{}

// withTraceID 将 trace ID 从 context 中提取并添加到 gRPC metadata
// 用于跨服务追踪请求
func (s *baseService) withTraceID(ctx context.Context) context.Context {
	// 尝试从 context 中获取 trace ID
	traceID := ""
	if val := ctx.Value("X-Request-ID"); val != nil {
		if id, ok := val.(string); ok {
			traceID = id
		}
	}

	// 如果有 trace ID，添加到 metadata
	if traceID != "" {
		md := metadata.Pairs("X-Trace-ID", traceID)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	return ctx
}
