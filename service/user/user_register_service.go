package user

import (
	"context"

	"go-cloud-disk/cache"
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils"
)

type UserRegisterService struct {
	NickName string `form:"nickname" json:"nickname" binding:"required,min=2,max=30"`
	UserName string `form:"username" json:"username" binding:"required,min=3,max=80"`
	Password string `form:"password" json:"password" binding:"required,min=3,max=40"`
	Code     string `form:"code" json:"code" binding:"required,min=6,max=6"`
}

type registerResponse struct {
	Token string `json:"token"`
	serializer.User
}

// vaild 检查注册信息是否正确
func (service *UserRegisterService) vaild() *serializer.Response {
	if service.Code != cache.RedisClient.Get(context.Background(), cache.EmailCodeKey(service.UserName)).Val() {
		return &serializer.Response{
			Code: serializer.CodeParamsError,
			Msg:  "验证码不正确",
		}
	}
	// 检查昵称
	count := int64(0)
	model.DB.Model(&model.User{}).Where("nick_name = ?", service.NickName).Count(&count)
	if count > 0 {
		return &serializer.Response{
			Code: serializer.CodeParamsError,
			Msg:  "昵称已被占用",
		}
	}

	// 检查用户名
	count = 0
	model.DB.Model(&model.User{}).Where("user_name = ?", service.UserName).Count(&count)
	if count > 0 {
		return &serializer.Response{
			Code: 40001,
			Msg:  "用户名已被占用",
		}
	}

	return nil
}

// Register 检查注册信息是否正确。如果正确，
// 将用户注册到数据库。否则，返回错误消息
func (service *UserRegisterService) Register() serializer.Response {
	user := model.User{
		NickName: service.NickName,
		UserName: service.UserName,
		Status:   model.StatusActiveUser,
	}

	// 检查用户有效性
	if err := service.vaild(); err != nil {
		return *err
	}

	// 加密密码
	if err := user.SetPassword(service.Password); err != nil {
		return serializer.Err(serializer.CodeError, "密码加密错误", err)
	}

	// 创建用户
	if err := user.CreateUser(); err != nil {
		return serializer.ParamsErr("创建用户错误", err)
	}

	// 生成JWT令牌
	token, err := utils.GenToken("crow", 24, &user)
	if err != nil {
		return serializer.Err(serializer.CodeError, "生成token错误", err)
	}

	return serializer.Success(registerResponse{
		Token: token,
		User:  serializer.BuildUser(user),
	})
}
