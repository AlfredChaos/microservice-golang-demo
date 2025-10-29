package middleware

import (
	"context"

	"github.com/alfredchaos/demo/pkg/reqctx"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	// TraceIDKey 追踪ID的元数据键名
	TraceIDKey = "X-Trace-ID"
)

// UnaryServerTracing gRPC 一元拦截器 - 追踪
// 从metadata中提取或生成trace-id，并传递到上下文中
func UnaryServerTracing() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		var traceID string

		// 从metadata中提取trace-id
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if traceIDs := md.Get(TraceIDKey); len(traceIDs) > 0 {
				traceID = traceIDs[0]
			}
		}

		// 如果没有trace-id，生成新的UUID
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// 将trace-id存储到上下文中
		ctx = reqctx.WithTraceID(ctx, traceID)

		// 调用实际的处理函数
		return handler(ctx, req)
	}
}

// StreamServerTracing gRPC 流拦截器 - 追踪
// 从metadata中提取或生成trace-id
func StreamServerTracing() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		var traceID string

		// 从metadata中提取trace-id
		ctx := ss.Context()
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if traceIDs := md.Get(TraceIDKey); len(traceIDs) > 0 {
				traceID = traceIDs[0]
			}
		}

		// 如果没有trace-id，生成新的UUID
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// 将trace-id存储到上下文中
		ctx = reqctx.WithTraceID(ctx, traceID)

		// 调用实际的处理函数
		return handler(srv, ss)
	}
}

// GetTraceID 从上下文中获取追踪ID
func GetTraceID(ctx context.Context) string {
	return reqctx.GetTraceID(ctx)
}
