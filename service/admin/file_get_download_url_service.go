package admin

import (
	"go-cloud-disk/disk"
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

// FileGetDownloadURLService 管理员获取文件下载/预览URL
type FileGetDownloadURLService struct{}

type fileGetDownloadURLResponse struct {
	Url string `json:"dowload_url"`
}

// GetDownloadURL 管理员获取任意文件的预签名URL（用于审核/预览）
func (service *FileGetDownloadURLService) GetDownloadURL(operStatus string, fileId string) serializer.Response {
	var file model.File
	if err := model.DB.Where("uuid = ?", fileId).First(&file).Error; err != nil {
		return serializer.DBErr("文件不存在", err)
	}

	// 普通管理员不能预览管理员/超级管理员的文件（与删除逻辑一致）
	if operStatus == model.StatusAdmin {
		var owner model.User
		if err := model.DB.Where("uuid = ?", file.Owner).First(&owner).Error; err != nil {
			return serializer.DBErr("文件拥有者不存在", err)
		}
		if owner.Status == model.StatusAdmin || owner.Status == model.StatusSuperAdmin {
			return serializer.NotAuthErr("")
		}
	}

	fileName := file.FileUuid + "." + file.FilePostfix
	// FilePath 为空时（历史数据/预签名入库漏写），回退到 Owner
	prefix := file.FilePath
	if prefix == "" {
		prefix = file.Owner
	}

	_, err := disk.BaseCloudDisk.IsObjectExist(prefix, "", fileName)
	if err != nil {
		logger.Log().Error("[AdminFileGetDownloadURL] 检查对象存在失败: ", err)
		return serializer.InternalErr("检查对象失败", err)
	}

	url, err := disk.BaseCloudDisk.GetDownloadPresignedURL(prefix, "", fileName)
	if err != nil {
		logger.Log().Error("[AdminFileGetDownloadURL] 获取预签名URL失败: ", err)
		return serializer.InternalErr("获取下载链接失败", err)
	}
	return serializer.Success(fileGetDownloadURLResponse{Url: url})
}
