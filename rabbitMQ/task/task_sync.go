package task

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"

	"go-cloud-disk/disk"
	"go-cloud-disk/model"
	"go-cloud-disk/rabbitMQ"
	"go-cloud-disk/utils"
	"go-cloud-disk/utils/logger"

	"github.com/disintegration/imaging"
	"gorm.io/gorm"
)

type SendConfirmEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type AutoTagRequest struct {
	FileID string `json:"file_id"`
	UserID string `json:"user_id"`
}

func RunSendConfirmEmail(ctx context.Context) error {
	msgs, err := rabbitMQ.ConsumerMessage(ctx, rabbitMQ.RabbitMqSendEmailQueue)
	if err != nil {
		return err
	}
	var forever chan struct{}

	go func() {
		for msg := range msgs {
			logger.Log().Info("[RunSendConfirmEmail] 收到消息: ", string(msg.Body))

			sendConirmEmailReq := SendConfirmEmailRequest{}
			err = json.Unmarshal(msg.Body, &sendConirmEmailReq)
			if err != nil {
				logger.Log().Error("[RunSendConfirmEmail] 解析消息错误: ", err)
			}

			err = utils.SendConfirmMessage(sendConirmEmailReq.Email, sendConirmEmailReq.Code)
			if err != nil {
				logger.Log().Error("[RunSendConfirmEmail] 发送确认邮件错误: ", err)
			}

			msg.Ack(false)
		}
	}()

	logger.Log().Info("发送确认邮件服务已启动")
	<-forever
	return nil
}

func RunAutoTagService(ctx context.Context) error {
	msgs, err := rabbitMQ.ConsumerMessage(ctx, rabbitMQ.RabbitMqAutoTagQueue)
	if err != nil {
		return err
	}
	forever := make(chan struct{})

	go func() {
		for msg := range msgs {
			logger.Log().Info("[RunAutoTagService] 收到消息: ", string(msg.Body))

			autoTagReq := AutoTagRequest{}
			err = json.Unmarshal(msg.Body, &autoTagReq)
			if err != nil {
				logger.Log().Error("[RunAutoTagService] 解析消息错误: ", err)
				msg.Nack(false, false) // 拒绝消息，不重新入队
				continue
			}

			err = processAutoTag(autoTagReq.FileID, autoTagReq.UserID)
			if err != nil {
				logger.Log().Error("[RunAutoTagService] 处理自动标签失败: ", err)
				msg.Nack(false, true) // 拒绝消息，重新入队
			} else {
				msg.Ack(false) // 确认消息
			}
		}
	}()

	logger.Log().Info("自动标签识别服务已启动")
	<-forever
	return nil
}

// processAutoTag 处理自动标签识别
func processAutoTag(fileID, userID string) error {
	// 查找文件的信息
	var file model.File
	if err := model.DB.Where("file_uuid = ?", fileID).First(&file).Error; err != nil {
		logger.Log().Error("[processAutoTag] 查找文件失败: ", err)
		return err
	}

	// 生成预签名下载URL
	downLoadURL, err := disk.BaseCloudDisk.GetDownloadURL(file.FilePath, file.FileUuid)
	if err != nil {
		logger.Log().Error("[processAutoTag] 生成预签名下载URL失败: ", err)
		return err
	}

	// 创建腾讯云标签识别实例
	tencentTag := disk.NewTencentImageTag()

	resp, err := http.Get(downLoadURL)
	if err != nil || resp.StatusCode != 200 {
		logger.Log().Error("[processAutoTag] 下载图片失败: ", err)
		return err
	}
	defer resp.Body.Close()

	// 读取图片数据
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log().Error("[processAutoTag] 读取图片数据失败: ", err)
		return err
	}

	// 记录原图大小
	originalSize := len(imageData)

	// 解码图片
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		logger.Log().Error("[processAutoTag] 解码图片失败: ", err)
		return err
	}

	// 裁剪中心区域（保留原图 90% 的宽高）
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	cropW := int(float64(w) * 0.9)
	cropH := int(float64(h) * 0.9)
	cropped := imaging.CropCenter(img, cropW, cropH)

	// 缩放图片（宽度最大 4096px，高度按比例）
	maxWidth := 4096
	resized := imaging.Resize(cropped, maxWidth, 0, imaging.Lanczos)

	// 压缩为JPEG并编码为Base64
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 90}); err != nil {
		logger.Log().Error("[processAutoTag] 图片压缩失败: ", err)
		return err
	}

	logger.Log().Info(fmt.Sprintf("[processAutoTag] 原图大小: %d bytes, 压缩后大小: %d bytes", originalSize, buf.Len()))

	base64Img := base64.StdEncoding.EncodeToString(buf.Bytes())

	tags, err := tencentTag.DetectLabelBase64(base64Img)
	if err != nil {
		logger.Log().Error("[processAutoTag] 调用图片标签识别失败: ", err)
		return err
	}

	// 保存标签到数据库
	if err := saveTagsToDatabase(fileID, tags); err != nil {
		logger.Log().Error("[processAutoTag] 保存标签到数据库失败: ", err)
		return err
	}

	logger.Log().Info(fmt.Sprintf("[processAutoTag] 文件标签识别完成: FileID=%s, 识别到%d个标签", fileID, len(tags)))
	return nil
}

// saveTagsToDatabase 保存标签到数据库
func saveTagsToDatabase(fileID string, tags []disk.TagResult) error {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, tag := range tags {
		// 查找或创建标签
		var dbTag model.Tag
		if err := tx.Where("name = ?", tag.Name).First(&dbTag).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 创建新标签
				dbTag = model.Tag{
					Name: tag.Name,
				}
				if err := tx.Create(&dbTag).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("创建标签失败: %v", err)
				}
			} else {
				tx.Rollback()
				return fmt.Errorf("查询标签失败: %v", err)
			}
		}

		// 检查是否已经存在文件标签关联
		var existingFileTag model.FileTag
		if err := tx.Where("file_id = ? AND tag_id = ?", fileID, dbTag.ID).First(&existingFileTag).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 创建新文件标签关联
				existingFileTag = model.FileTag{
					FileID: fileID,
					TagID:  dbTag.ID,
				}
				if err := tx.Create(&existingFileTag).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("创建文件标签关联失败: %v", err)
				}
			} else {
				tx.Rollback()
				return fmt.Errorf("查询文件标签关联失败: %v", err)
			}
		}
		// 如果已存在关联，跳过
	}

	return tx.Commit().Error
}
