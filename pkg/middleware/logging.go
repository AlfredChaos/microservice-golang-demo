package middleware

import (
	"context"
	"time"

	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// UnaryServerLogging gRPC 一元拦截器 - 日志记录
// 记录每个gRPC请求的详细信息
func UnaryServerLogging() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 记录开始时间
		startTime := time.Now()

		// 调用实际的处理函数
		resp, err := handler(ctx, req)

		// 计算耗时
		latency := time.Since(startTime)

		// 提取 trace ID
		traceID := GetTraceID(ctx)

		// 记录日志
		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.String("X-Trace-ID", traceID),
			zap.Duration("latency", latency),
		}

		if err != nil {
			fields = append(fields, zap.Error(err))
			log.Error("gRPC request error", fields...)
		} else {
			log.Info("gRPC request", fields...)
		}

		return resp, err
	}
}

// StreamServerLogging gRPC 流拦截器 - 日志记录
// 记录流式gRPC请求的信息
func StreamServerLogging() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// 记录开始时间
		startTime := time.Now()

		// 调用实际的处理函数
		err := handler(srv, ss)

		// 计算耗时
		latency := time.Since(startTime)

		// 提取 trace ID
		ctx := ss.Context()
		traceID := GetTraceID(ctx)

		// 记录日志
		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.String("trace_id", traceID),
			zap.Duration("latency", latency),
			zap.Bool("is_client_stream", info.IsClientStream),
			zap.Bool("is_server_stream", info.IsServerStream),
		}

		if err != nil {
			fields = append(fields, zap.Error(err))
			log.Error("gRPC stream error", fields...)
		} else {
			log.Info("gRPC stream", fields...)
		}

		return err
	}
}
