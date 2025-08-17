package filefolder

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

// FileFolderGetAllFileFolderService 获取文件夹所有子文件夹服务结构体
type FileFolderGetAllFileFolderService struct{}

// GetAllFileFolder 获取用户文件夹中的所有子文件夹
func (service *FileFolderGetAllFileFolderService) GetAllFileFolder(userId string, fileFolderID string) serializer.Response {
	var filefolder []model.FileFolder
	if err := model.DB.Where("parent_folder_id = ? and owner_id = ?", fileFolderID, userId).Find(&filefolder).Error; err != nil {
		logger.Log().Error("[FileFolderGetAllFileFolderService.GetAllFileFolder] 查找文件夹失败: ", err)
		return serializer.DBErr("", err)
	}
	return serializer.Success(serializer.BuildFileFolders(filefolder))
}
