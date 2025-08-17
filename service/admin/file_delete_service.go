package admin

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

type FileDeleteService struct{}

// FileDelete 删除所有具有相同MD5码的文件
func (service *FileDeleteService) FileDelete(operStatus string, fileId string) serializer.Response {
	// 从数据库获取要删除的文件
	var err error
	var deleteFile model.File
	if err = model.DB.Where("uuid = ?", fileId).Find(&deleteFile).Error; err != nil {
		logger.Log().Error("[FileDeleteService.FileDelete] 查找要删除的文件信息失败: ", err)
		return serializer.DBErr("", err)
	}

	// 获取文件所有者
	var fileOwner model.User
	if err = model.DB.Where("uuid = ?", deleteFile.Owner).Find(&fileOwner).Error; err != nil {
		logger.Log().Error("[FileDeleteService.FileDelete] 查找文件所有者失败: ", err)
		return serializer.DBErr("", err)
	}

	// 不能删除管理员的文件
	if operStatus == model.StatusAdmin {
		if fileOwner.Status == model.StatusAdmin || fileOwner.Status == model.StatusSuperAdmin {
			return serializer.NotAuthErr("")
		}
	}

	// 删除所有相同文件
	var files []model.File
	if err = model.DB.Where("file_uuid = ?", deleteFile.FileUuid).Find(&files).Error; err != nil {
		logger.Log().Error("[FileDeleteService.FileDelete] 查找要删除的文件列表失败: ", err)
		return serializer.DBErr("", err)
	}

	if err = model.DB.Delete(&files).Error; err != nil {
		logger.Log().Error("[FileDeleteService.FileDelete] 删除文件失败: ", err)
		return serializer.DBErr("", err)
	}

	return serializer.Success(nil)
}
