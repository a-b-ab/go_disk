package share

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

// ShareDeleteService 删除分享服务结构体
type ShareDeleteService struct{}

// DeleteShare 删除用户分享
func (service *ShareDeleteService) DeleteShare(shareId string, userId string) serializer.Response {
	// 从数据库获取分享信息
	var share model.Share
	if err := model.DB.Where("uuid = ? and owner = ?", shareId, userId).First(&share).Error; err != nil {
		logger.Log().Error("[ShareDeleteService.DeleteShare] 获取分享信息失败: ", err)
		return serializer.DBErr("", err)
	}

	// 延迟双重删除，确保信息安全
	share.DeleteShareInfoInRedis()
	if err := model.DB.Delete(&share).Error; err != nil {
		logger.Log().Error("[ShareDeleteService.DeleteShare] 删除分享失败: ", err)
		return serializer.DBErr("", err)
	}
	share.DeleteShareInfoInRedis()

	return serializer.Success(nil)
}
