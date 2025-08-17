package user

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

type UserUpdateService struct {
	NickName string `form:"nickname" json:"nickname" binding:"required,min=2,max=30"`
}

// UpdateUserInfo 更新用户昵称
func (service *UserUpdateService) UpdateUserInfo(userId string) serializer.Response {
	// 检查是否能找到对应用户ID的用户
	var user model.User
	if err := model.DB.Where("uuid = ?", userId).Find(&user).Error; err != nil {
		logger.Log().Error("[UserUpdateService.UpdateUserInfo] 查找用户失败: ", err)
		return serializer.DBErr("", err)
	}
	// 更新用户信息到数据库
	user.NickName = service.NickName
	if err := model.DB.Save(&user).Error; err != nil {
		logger.Log().Error("[UserUpdateService.UpdateUserInfo] 保存用户信息失败: ", err)
		return serializer.DBErr("", err)
	}
	return serializer.Success(serializer.BuildUser(user))
}
