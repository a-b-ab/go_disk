package tag

import (
	"context"
	"encoding/json"
	"time"

	"go-cloud-disk/rabbitMQ"
	"go-cloud-disk/rabbitMQ/task"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"

	"github.com/gin-gonic/gin"
)

type AutoGetTag struct {
	FileID string `json:"file_id"`
}

func (service *AutoGetTag) GetAutoTags(c *gin.Context) serializer.Response {
	// todo 用户鉴权

	// 发送自动标签识别任务到MQ
	if err := service.sendAutoTagToMQ(service.FileID, ""); err != nil {
		logger.Log().Error("[AutoGetTag.GetAutoTags] 发送MQ消息失败: ", err)
		return serializer.InternalErr("发送标签识别任务失败", err)
	}

	return serializer.Response{
		Code: 0,
		Msg:  "标签识别任务已提交，稍后自动完成",
	}
}

// sendAutoTagToMQ 将自动标签识别任务发送到消息队列
func (service *AutoGetTag) sendAutoTagToMQ(fileID string, userID string) error {
	// 限制1秒超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	autoTagReq := task.AutoTagRequest{
		FileID: fileID,
		UserID: userID,
	}

	body, err := json.Marshal(autoTagReq)
	if err != nil {
		logger.Log().Error("[AutoGetTag.sendAutoTagToMQ] 序列化请求失败: ", err)
		return err
	}

	err = rabbitMQ.SendMessageToMQ(ctx, rabbitMQ.RabbitMqAutoTagQueue, body)
	if err != nil {
		return err
	}

	return nil
}
