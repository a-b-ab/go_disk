package api

import (
	"go-cloud-disk/serializer"
	"go-cloud-disk/service/user"
	"github.com/gin-gonic/gin"
)

// UserLogin 用户登录接口
func UserLogin(c *gin.Context) {
	var service user.UserLoginService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	res := service.Login(c)
	c.JSON(200, res)
}

// UserRegiser 用户注册接口
func UserRegiser(c *gin.Context) {
	var service user.UserRegisterService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	res := service.Register()
	c.JSON(200, res)
}

// UserInfo 获取用户信息
func UserInfo(c *gin.Context) {
	var service user.UserInfoService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	res := service.GetUserInfo(c.Param("id"))
	c.JSON(200, res)
}

// UserMyInfo 从JWT信息中获取用户信息
func UserMyInfo(c *gin.Context) {
	var service user.UserInfoService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	userIdString := c.MustGet("UserId").(string)
	res := service.GetUserInfo(userIdString)
	c.JSON(200, res)
}

// UpdateUserInfo 更新用户昵称
func UpdateUserInfo(c *gin.Context) {
	var service user.UserUpdateService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	userId := c.MustGet("UserId").(string)
	res := service.UpdateUserInfo(userId)
	c.JSON(200, res)
}

// ConfirmUserEmail 发送确认邮件
func ConfirmUserEmail(c *gin.Context) {
	var service user.UserSendConfirmEmailService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}
	res := service.SendConfirmEmail()
	c.JSON(200, res)
}
