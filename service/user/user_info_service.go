package user

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

type UserInfoService struct{}

// GetUserInfo 根据用户ID获取用户信息
func (service *UserInfoService) GetUserInfo(userid string) serializer.Response {
	var user model.User

	err := model.DB.Model(&model.User{}).Where("uuid = ?", userid).First(&user).Error
	if err != nil {
		logger.Log().Error("[UserInfoService.GetUserInfo] 查找用户失败")
		return serializer.ParamsErr("未找到用户", err)
	}

	return serializer.Success(serializer.BuildUser(user))
}
