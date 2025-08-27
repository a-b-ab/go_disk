package tag

import (
	"fmt"

	"go-cloud-disk/disk"
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AutoGetTag struct {
	FileID string `json:"file_id"`
}

func (service *AutoGetTag) GetAutoTags(c *gin.Context) serializer.Response {
	// 查找文件的信息
	var file model.File
	if err := model.DB.Where("file_uuid = ?", service.FileID).First(&file).Error; err != nil {
		logger.Log().Error("[ShareDownloadService.GetDownloadUrl] 查找文件失败: ", err)
		return serializer.DBErr("文件不存在", err)
	}

	// 生成预签名下载URL
	downLoadURL, err := disk.BaseCloudDisk.GetDownloadURL(file.FilePath, file.FileUuid)
	if err != nil {
		logger.Log().Error("[ShareDownloadService.GetDownloadUrl] 生成预签名下载URL失败: ", err)
		return serializer.DBErr("生成预签名下载URL失败", err)
	}

	// 创建腾讯云标签识别实例
	tencentTag := disk.NewTencentImageTag()

	// // 编码为Base64
	// imageBase64, err := utils.ImageURLToBase64(downLoadURL)
	// if err != nil {
	// 	logger.Log().Error("[ShareDownloadService.GetDownloadUrl] 图片转Base64失败: ", err)
	// 	return serializer.DBErr("图片转Base64失败", err)
	// }

	// // 调用图片标签识别
	// tags, err := tencentTag.DetectLabelBase64(imageBase64)
	tags, err := tencentTag.DetectLabels(downLoadURL)
	if err != nil {
		logger.Log().Error("[ShareDownloadService.GetDownloadUrl] 调用图片标签识别失败: ", err)
		return serializer.DBErr("调用图片标签识别失败", err)
	}

	// 过滤敏感词，todo

	// 保存标签到数据库
	if err := service.saveTagsToDatabase(service.FileID, tags); err != nil {
		logger.Log().Error("[ShareDownloadService.GetDownloadUrl] 保存标签到数据库失败: ", err)
		return serializer.DBErr("保存标签到数据库失败", err)
	}

	return serializer.Response{}
}

// saveTagsToDatabase 保存标签到数据库
func (service *AutoGetTag) saveTagsToDatabase(fileID string, tags []disk.TagResult) error {
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
