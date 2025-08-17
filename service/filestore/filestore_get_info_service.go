package filestore

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

// FileStoreGetInfoService 获取文件存储信息服务结构体
type FileStoreGetInfoService struct{}

// FileStoreGetInfo 获取用户文件存储信息
func (service *FileStoreGetInfoService) FileStoreGetInfo(userId string, storeId string) serializer.Response {
	// 检查存储空间所有者
	var store model.FileStore
	if err := model.DB.Where("uuid = ? and owner_id = ?", storeId, userId).Find(&store).Error; err != nil {
		logger.Log().Error("[FileStoreGetInfoService.FileStoreGetInfo] 查找用户存储空间失败: ", err)
		return serializer.DBErr("", err)
	}
	return serializer.Success(serializer.BuildFileStore(store))
}
