package admin

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

type UserChangeAuthService struct {
	UserId    string `json:"userid" form:"userid" required:"binding"`
	NewStatus string `json:"status" form:"status" required:"binding"`
}

// UserChangeAuth 更改用户权限，需要输入使用此功能的用户状态
func (service *UserChangeAuthService) UserChangeAuth(userStatus string) serializer.Response {
	// 从数据库获取用户信息
	var user model.User
	if err := model.DB.Where("uuid = ?", service.UserId).Find(&user).Error; err != nil {
		logger.Log().Error("[UserChangeAuthService.UserChangeAuth] 查找用户信息失败: ", err)
		return serializer.DBErr("", err)
	}

	if user.Uuid == "" {
		return serializer.ParamsErr("", nil)
	}

	// 检查用户是否为管理员
	if userStatus != model.StatusAdmin && userStatus != model.StatusSuperAdmin {
		return serializer.NotAuthErr("")
	}

	// 普通管理员不能更改管理员权限
	if userStatus == model.StatusAdmin {
		if user.Status == model.StatusAdmin || user.Status == model.StatusSuperAdmin {
			return serializer.NotAuthErr("")
		}

		if service.NewStatus == model.StatusAdmin || service.NewStatus == model.StatusSuperAdmin {
			return serializer.NotAuthErr("")
		}
	}

	// 保存用户权限
	user.Status = service.NewStatus
	if err := model.DB.Save(&user).Error; err != nil {
		logger.Log().Error("[UserChangeAuthService.UserChangeAuth] 保存用户信息失败: ", err)
		return serializer.DBErr("", err)
	}
	return serializer.Success(serializer.BuildUser(user))
}
