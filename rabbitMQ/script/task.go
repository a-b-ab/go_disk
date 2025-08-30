package script

import (
	"context"

	"go-cloud-disk/rabbitMQ/task"
	"go-cloud-disk/utils/logger"
)

func SendConfirmEmailSync(ctx context.Context) {
	err := task.RunSendConfirmEmail(ctx)
	if err != nil {
		logger.Log().Error("[SendConfirmEmailSync] 发送确认邮件失败: ", err)
	}
}

func AutoTagSync(ctx context.Context) {
	err := task.RunAutoTagService(ctx)
	if err != nil {
		logger.Log().Error("[AutoTagSync] 自动标签识别服务失败: ", err)
	}
}

func FileCleanSync(ctx context.Context) {
	err := task.RunFileCleanService(ctx)
	if err != nil {
		logger.Log().Error("[FileCleanSync] 文件清理服务失败: ", err)
	}
}
