package file

import (
	"go-cloud-disk/disk"
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

// FileGetDownloadURLService 获取文件下载URL服务结构体
type FileGetDownloadURLService struct{}

// fileGetDownloadURLResponse 获取文件下载URL响应结构体
type fileGetDownloadURLResponse struct {
	Url string `json:"dowload_url"` // 下载URL
}

// GetDownloadURL 获取文件下载URL
func (service *FileGetDownloadURLService) GetDownloadURL(userId string, fileid string) serializer.Response {
	var file model.File
	if err := model.DB.Where("uuid = ?", fileid).Find(&file).Error; err != nil {
		logger.Log().Error("[fileGetDownloadURLResponse.GetDownloadURL] 查找用户文件失败: ", err)
		return serializer.DBErr("", err)
	}

	if userId != file.Owner {
		return serializer.NotAuthErr("")
	}

	fileName := file.FileUuid + "." + file.FilePostfix
	// 使用预签名URL，避免桶为私有时 ObjectURL 无法直接访问；
	// 同时通过预签名的 response-content-disposition=inline 支持浏览器直接预览图片等资源。
	url, err := disk.BaseCloudDisk.GetDownloadPresignedURL(file.Owner, "", fileName)
	if err != nil {
		logger.Log().Error("[FileGetDownloadURLService.GetDownloadURL] 获取下载URL失败: ", err)
		return serializer.InternalErr("", err)
	}
	return serializer.Success(fileGetDownloadURLResponse{
		Url: url,
	})
}
