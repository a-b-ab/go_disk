package admin

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

type UserFilestoreUpdateService struct {
	UserId        string `json:"userid" form:"userid" required:"binding"`
	NewStoreVolum int64  `json:"volum" form:"volum" required:"binding"`
}

func (service *UserFilestoreUpdateService) UserFilestoreUpdate() serializer.Response {
	// 从数据库搜索文件存储信息
	var userFilestore model.FileStore
	if err := model.DB.Where("owner_id = ?", service.UserId).First(&userFilestore).Error; err != nil {
		logger.Log().Error("[UserFilestoreUpdateService.UserFilestoreUpdate] 查找文件存储信息失败: ", err)
		return serializer.DBErr("", err)
	}

	// 最大容量限制为1GB
	userFilestore.MaxSize = min(service.NewStoreVolum, int64(1024*1024*1024))
	userFilestore.MaxSize = max(0, userFilestore.MaxSize)

	if err := model.DB.Save(&userFilestore).Error; err != nil {
		logger.Log().Error("[UserFilestoreUpdateService.UserFilestoreUpdate] 更新文件存储信息失败: ", err)
		return serializer.DBErr("", err)
	}

	return serializer.Success(nil)
}
