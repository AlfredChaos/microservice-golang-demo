package grpcclient

import (
	"context"
	"time"
	
	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// LoggingInterceptor 日志拦截器
func LoggingInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		
		log.WithContext(ctx).Info("grpc client call",
			zap.String("method", method),
			zap.String("target", cc.Target()))
		
		err := invoker(ctx, method, req, reply, cc, opts...)
		
		duration := time.Since(start)
		if err != nil {
			log.WithContext(ctx).Error("grpc client call failed",
				zap.String("method", method),
				zap.Duration("duration", duration),
				zap.Error(err))
		} else {
			log.WithContext(ctx).Info("grpc client call completed",
				zap.String("method", method),
				zap.Duration("duration", duration))
		}
		
		return err
	}
}

// TracingInterceptor 追踪拦截器
// 将trace ID从context传递到gRPC metadata
func TracingInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// 从context中提取trace ID
		traceID := ""
		if val := ctx.Value("X-Request-ID"); val != nil {
			if id, ok := val.(string); ok {
				traceID = id
			}
		}
		
		// 添加到metadata
		if traceID != "" {
			md := metadata.Pairs("X-Trace-ID", traceID)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// RetryInterceptor 重试拦截器
func RetryInterceptor(cfg *RetryConfig) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var err error
		
		for i := 0; i <= cfg.Max; i++ {
			err = invoker(ctx, method, req, reply, cc, opts...)
			if err == nil {
				return nil
			}
			
			// 最后一次不需要等待
			if i < cfg.Max {
				time.Sleep(cfg.Backoff)
			}
		}
		
		return err
	}
}
