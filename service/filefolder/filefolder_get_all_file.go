package filefolder

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

// FileFolderGetAllFileService 获取文件夹所有文件服务结构体
type FileFolderGetAllFileService struct{}

// GetAllFile 获取用户文件夹中的所有文件
func (service *FileFolderGetAllFileService) GetAllFile(userId string, fileFolderID string) serializer.Response {
	// 检查用户是否匹配
	var fileFolder model.FileFolder
	if err := model.DB.Where("uuid = ? and owner_id = ?", fileFolderID, userId).Find(&fileFolder).Error; err != nil {
		logger.Log().Error("[FileFolderGetAllFileService.GetAllFile] 获取文件夹失败: ", err)
		return serializer.DBErr("", err)
	}

	var files []model.File
	if err := model.DB.Where("parent_folder_id = ?", fileFolderID).Find(&files).Error; err != nil {
		logger.Log().Error("[FileFolderGetAllFileService.GetAllFile] 获取文件列表失败: ", err)
		return serializer.DBErr("", err)
	}
	return serializer.Success(serializer.BuildFiles(files))
}
