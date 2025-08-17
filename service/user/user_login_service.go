package user

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils"
	"github.com/gin-gonic/gin"
)

type UserLoginService struct {
	UserName string `form:"username" json:"username" binding:"required,min=3,max=30"`
	Password string `form:"password" json:"password" binding:"required,min=3,max=40"`
}

type returnUser struct {
	Token string `json:"token"`
	serializer.User
}

// Login 检查用户名和密码是否匹配
// 并返回用户信息和JWT令牌
func (service *UserLoginService) Login(c *gin.Context) serializer.Response {
	var user model.User

	if err := model.DB.Where("user_name = ?", service.UserName).First(&user).Error; err != nil {
		return serializer.ParamsErr("账号或密码错误", nil)
	}

	if !user.CheckPassword(service.Password) {
		return serializer.ParamsErr("账号或密码错误", nil)
	}
	token, err := utils.GenToken("crow", 24, &user)

	// 管理员令牌只有1小时有效期
	if user.Status == model.StatusAdmin || user.Status == model.StatusSuperAdmin {
		token, err = utils.GenToken("crow", 1, &user)
	}

	if err != nil {
		return serializer.InternalErr("GetTokenErr", err)
	}
	return serializer.Success(returnUser{
		Token: token,
		User:  serializer.BuildUser(user),
	})
}
