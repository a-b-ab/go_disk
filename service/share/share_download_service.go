package share

import (
	"go-cloud-disk/disk"
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

// ShareDownloadService 分享下载服务
type ShareDownloadService struct {
	ShareId string `json:"shareid" form:"shareid""` // 分享ID
}

type shareDownloadResponse struct {
	DownloadUrl string `json:"downloadurl"` // 预签名下载链接
}

// GetDownloadUrl 根据分享ID生成预签名下载链接
func (service *ShareDownloadService) GetDownloadUrl(shareId string) serializer.Response {
	// 查找分享记录
	var share model.Share
	if err := model.DB.Where("uuid = ?", shareId).First(&share).Error; err != nil {
		logger.Log().Error("[ShareDownloadService.GetDownloadUrl] 查找分享失败: ", err)
		return serializer.DBErr("分享不存在", err)
	}

	// 查找对应的文件信息
	var file model.File
	if err := model.DB.Where("uuid = ?", share.FileId).First(&file).Error; err != nil {
		logger.Log().Error("[ShareDownloadService.GetDownloadUrl] 查找文件失败: ", err)
		return serializer.DBErr("文件不存在", err)
	}

	// 生成预签名下载URL
	downloadUrl, err := disk.BaseCloudDisk.GetDownloadURL(file.FilePath, file.FileUuid)
	if err != nil {
		logger.Log().Error("[ShareDownloadService.GetDownloadUrl] 生成预签名下载URL失败: ", err)
		return serializer.DBErr("生成预签名下载URL失败", err)
	}

	return serializer.Success(shareDownloadResponse{
		DownloadUrl: downloadUrl,
	})
}
