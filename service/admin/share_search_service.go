package admin

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

type ShareSearchService struct {
	Uuid  string `json:"uuid" form:"uuid"`
	Title string `json:"title" form:"title"`
	Owner string `json:"owner" form:"owner"`
}

// ShareSearch 根据uuid、标题或所有者搜索分享
func (service *ShareSearchService) ShareSearch() serializer.Response {
	var shares []model.Share

	// 构建搜索条件
	searchInfo := model.DB.Model(&model.Share{})
	if service.Uuid != "" {
		searchInfo.Where("uuid = ?", service.Uuid)
	}
	if service.Title != "" {
		searchInfo.Where("title like ?", "%"+service.Title+"%")
	}
	if service.Owner != "" {
		searchInfo.Where("owner = ?", service.Owner)
	}

	// 从数据库搜索分享
	if err := searchInfo.Find(&shares).Error; err != nil {
		logger.Log().Error("[ShareSearchService.ShareSearch] 查找分享失败: ", err)
		return serializer.DBErr("", err)
	}

	return serializer.Success(serializer.BuildShares(shares))
}
