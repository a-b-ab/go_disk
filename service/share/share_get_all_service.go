package share

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

// ShareGetAllService 获取所有分享服务结构体
type ShareGetAllService struct{}

// GetAllShare 获取用户的所有分享
func (service *ShareGetAllService) GetAllShare(userId string) serializer.Response {
	// 从数据库获取分享列表
	var shares []model.Share
	if err := model.DB.Where("owner = ?", userId).Find(&shares).Error; err != nil {
		logger.Log().Error("[ShareGetAllService.GetAllShare] 获取分享信息失败: ", err)
		return serializer.DBErr("", err)
	}

	return serializer.Success(serializer.BuildShares(shares))
}
