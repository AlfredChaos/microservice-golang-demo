package biz

import (
	"context"

	"github.com/alfredchaos/demo/pkg/log"
	"go.uber.org/zap"
)

// TaskMessage 任务消息结构
type TaskMessage struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	TaskType  string `json:"task_type"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

// TaskUseCase 任务业务逻辑用例接口
type ITaskUseCase interface {
	HandleSayHelloTask(ctx context.Context, msg *TaskMessage) error
	// 未来可以添加其他任务处理方法
	// HandleNotificationTask(ctx context.Context, msg *TaskMessage) error
	// HandleReportTask(ctx context.Context, msg *TaskMessage) error
}

// TaskUseCase 任务业务逻辑用例实现
type TaskUseCase struct {
	// 可以注入其他依赖，如数据库、缓存、gRPC客户端等
	// userClient userv1.UserServiceClient
	// db         *sql.DB
	// cache      cache.Cache
}

// NewTaskUseCase 创建新的任务业务逻辑用例
func NewTaskUseCase() *TaskUseCase {
	return &TaskUseCase{}
}

// HandleSayHelloTask 处理 SayHello 任务
func (uc *TaskUseCase) HandleSayHelloTask(ctx context.Context, msg *TaskMessage) error {
	log.WithContext(ctx).Info("processing sayhello task",
		zap.String("user_id", msg.UserID),
		zap.String("username", msg.Username),
		zap.String("message", msg.Message))

	// 这里可以添加实际的业务逻辑
	// 例如：
	// 1. 调用其他微服务
	// if uc.userClient != nil {
	//     resp, err := uc.userClient.GetUser(ctx, &userv1.GetUserRequest{Id: msg.UserID})
	//     if err != nil {
	//         return fmt.Errorf("failed to get user: %w", err)
	//     }
	// }
	//
	// 2. 写入数据库
	// if uc.db != nil {
	//     _, err := uc.db.ExecContext(ctx, "INSERT INTO task_logs ...")
	//     if err != nil {
	//         return fmt.Errorf("failed to log task: %w", err)
	//     }
	// }
	//
	// 3. 更新缓存
	// if uc.cache != nil {
	//     uc.cache.Set(ctx, key, value)
	// }
	//
	// 4. 发送通知
	// - 发送邮件通知
	// - 发送短信通知
	// - 推送消息
	//
	// 5. 生成报表或文件
	// 等等...

	log.WithContext(ctx).Info("sayhello task processed successfully",
		zap.String("user_id", msg.UserID))

	return nil
}
