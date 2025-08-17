package admin

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

type UserSearchService struct {
	Uuid     string `json:"uuid" form:"uuid"`
	NickName string `json:"nickname" form:"nickname"`
	Status   string `json:"status" form:"status"`
}

// UserSearch 根据uuid、昵称或状态搜索用户
func (service *UserSearchService) UserSearch() serializer.Response {
	var users []model.User

	// 构建搜索查询条件
	searchInfo := model.DB.Model(&model.User{})
	if service.Uuid != "" {
		searchInfo.Where("uuid = ?", service.Uuid)
	}
	if service.NickName != "" {
		searchInfo.Where("nick_name like ?", "%"+service.NickName+"%")
	}
	if service.Status != "" {
		searchInfo.Where("status = ?", service.Status)
	}

	// 在数据库中搜索用户
	if err := searchInfo.Find(&users).Error; err != nil {
		logger.Log().Error("[UserSearchService.UserSearch] 查找用户失败: ", err)
		return serializer.DBErr("", err)
	}

	return serializer.Success(serializer.BuildUsers(users))
}
