package admin

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

type ShareDeleteService struct{}

func (service *ShareDeleteService) ShareDelete(shareId string) serializer.Response {
	// 从数据库获取分享信息
	if err := model.DB.Where("uuid = ?", shareId).Delete(&model.Share{}).Error; err != nil {
		logger.Log().Error("[ShareDeleteService.ShareDelete] 获取分享信息失败: ", err)
		return serializer.DBErr("", err)
	}

	// 删除存储在Redis中的分享信息
	share := model.Share{
		Uuid: shareId,
	}
	share.DeleteShareInfoInRedis()

	return serializer.Success(nil)
}
