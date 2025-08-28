package tag

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"

	"go-cloud-disk/disk"
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"

	"github.com/disintegration/imaging"

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

	resp, err := http.Get(downLoadURL)
	if err != nil || resp.StatusCode != 200 {
		logger.Log().Error("[ShareDownloadService.GetDownloadUrl] 下载图片失败: ", err)
		return serializer.DBErr("下载图片失败", err)
	}
	defer resp.Body.Close()

	// 读取图片数据
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log().Error("[ShareDownloadService.GetDownloadUrl] 读取图片数据失败: ", err)
		return serializer.DBErr("读取图片数据失败", err)
	}

	// 记录原图大小
	originalSize := len(imageData)

	// 解码图片
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		logger.Log().Error("[ShareDownloadService.GetDownloadUrl] 解码图片失败: ", err)
		return serializer.DBErr("解码图片失败", err)
	}
	// 裁剪中心区域（保留原图 80% 的宽高）
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
		logger.Log().Error("[AutoGetTag.GetAutoTags] 图片压缩失败: ", err)
		return serializer.DBErr("图片压缩失败", err)
	}

	fmt.Printf("原图大小: %d bytes, 压缩后大小: %d bytes\n", originalSize, buf.Len())

	base64Img := base64.StdEncoding.EncodeToString(buf.Bytes())

	tags, err := tencentTag.DetectLabelBase64(base64Img)
	if err != nil {
		logger.Log().Error("[ShareDownloadService.GetDownloadUrl] 调用图片标签识别失败: ", err)
		return serializer.DBErr("调用图片标签识别失败", err)
	}

	// 压缩掉了约 84.7% 的文件体积

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
