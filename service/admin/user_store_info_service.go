package admin

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

type FileStoreGetInfoService struct{}

// FileStoreGetInfo 根据用户ID获取用户存储信息
func (service *FileStoreGetInfoService) FileStoreGetInfo(userId string) serializer.Response {
	// 从数据库获取存储信息
	var store model.FileStore
	if err := model.DB.Where("owner_id = ?", userId).First(&store).Error; err != nil {
		logger.Log().Error("[FileStoreGetInfoService.FileStoreGetInfo] 获取用户文件存储信息失败: ", err)
		return serializer.DBErr("", err)
	}
	return serializer.Success(serializer.BuildFileStore(store))
}
