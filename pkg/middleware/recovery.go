package middleware

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerRecovery gRPC 一元拦截器 - Panic恢复
// 捕获panic，记录错误日志，并返回Internal错误
func UnaryServerRecovery() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				// 获取堆栈信息
				stackBytes := debug.Stack()
				stackStr := string(stackBytes)
				
				// 将堆栈按行分割，便于日志查看
				stackLines := strings.Split(stackStr, "\n")
				
				// 过滤空行
				var filteredStack []string
				for _, line := range stackLines {
					if strings.TrimSpace(line) != "" {
						filteredStack = append(filteredStack, line)
					}
				}
				
				// 记录错误日志
				log.Error("gRPC panic recovered",
					zap.String("method", info.FullMethod),
					zap.String("panic_error", fmt.Sprintf("%v", r)),
					zap.String("service_type", "unary"),
					zap.Strings("stack_trace", filteredStack),
				)
				
				// 返回Internal错误
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		
		return handler(ctx, req)
	}
}

// StreamServerRecovery gRPC 流拦截器 - Panic恢复
// 捕获流式请求中的panic
func StreamServerRecovery() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) (err error) {
		defer func() {
			if r := recover(); r != nil {
				// 获取堆栈信息
				stackBytes := debug.Stack()
				stackStr := string(stackBytes)
				
				// 将堆栈按行分割，便于日志查看
				stackLines := strings.Split(stackStr, "\n")
				
				// 过滤空行
				var filteredStack []string
				for _, line := range stackLines {
					if strings.TrimSpace(line) != "" {
						filteredStack = append(filteredStack, line)
					}
				}
				
				// 记录错误日志
				log.Error("gRPC stream panic recovered",
					zap.String("method", info.FullMethod),
					zap.String("panic_error", fmt.Sprintf("%v", r)),
					zap.String("service_type", "stream"),
					zap.Bool("is_client_stream", info.IsClientStream),
					zap.Bool("is_server_stream", info.IsServerStream),
					zap.Strings("stack_trace", filteredStack),
				)
				
				// 返回Internal错误
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		
		return handler(srv, ss)
	}
}
