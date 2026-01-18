package tag

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"go-cloud-disk/model"
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
	userID := c.MustGet("UserId").(string)

	// 校验文件归属 + 仅图片支持自动标签
	var file model.File
	if err := model.DB.Where("uuid = ? AND owner = ?", service.FileID, userID).First(&file).Error; err != nil {
		return serializer.ParamsErr("文件不存在", err)
	}
	postfix := strings.ToLower(file.FilePostfix)
	switch postfix {
	case "png", "jpg", "jpeg", "gif", "webp", "bmp":
	default:
		return serializer.ParamsErr("仅图片支持自动标签识别", nil)
	}

	// 发送自动标签识别任务到MQ
	// 注意：MQ 侧目前按 file_uuid（md5）查找，因此这里传 file.FileUuid
	if err := service.sendAutoTagToMQ(file.FileUuid, userID); err != nil {
		logger.Log().Error("[AutoGetTag.GetAutoTags] 发送MQ消息失败: ", err)
		return serializer.InternalErr("发送标签识别任务失败", err)
	}

	// 返回标准成功响应（前端 http.ts 以 code===200 判断成功）
	return serializer.Success(map[string]any{
		"msg": "标签识别任务已提交，稍后自动完成",
	})
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
