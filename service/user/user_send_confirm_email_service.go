package user

import (
	"context"
	"encoding/json"
	"math/rand"
	"strconv"
	"time"

	"go-cloud-disk/cache"
	"go-cloud-disk/model"
	"go-cloud-disk/rabbitMQ"
	"go-cloud-disk/rabbitMQ/task"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils"
	"go-cloud-disk/utils/logger"
)

type UserSendConfirmEmailService struct {
	UserEmail string `json:"email" form:"email" binding:"required"`
}

func getConfirmCode() string {
	var confirmCode int
	for i := 0; i < 6; i++ {
		confirmCode = confirmCode*10 + (rand.Intn(9) + 1)
	}
	confirmCodeStr := strconv.Itoa(confirmCode)
	return confirmCodeStr
}

func (service *UserSendConfirmEmailService) SendConfirmEmail() serializer.Response {
	// 检查邮箱格式
	if !utils.VerifyEmailFormat(service.UserEmail) {
		return serializer.ParamsErr("NotEmail", nil)
	}
	// 检查用户最近发送邮件的次数限制
	if cache.RedisClient.Get(context.Background(), cache.RecentSendUserKey(service.UserEmail)).Val() != "" {
		return serializer.ParamsErr("HasSendCode", nil)
	}

	// 检查邮箱是否已注册
	var emailNum int64
	if err := model.DB.Model(&model.User{}).Where("user_name = ?", service.UserEmail).Count(&emailNum).Error; err != nil {
		logger.Log().Error("[UserSendConfirmEmailService.SendConfirmEmail] 查找用户失败: ", err)
		return serializer.DBErr("", err)
	}
	if emailNum > 0 {
		return serializer.ParamsErr("HasRegister", nil)
	}

	code := getConfirmCode()
	cache.RedisClient.Set(context.Background(), cache.EmailCodeKey(service.UserEmail), code, time.Minute*30)

	if err := service.sendConfirmEmailToMQ(service.UserEmail, code); err != nil {
		return serializer.InternalErr("", err)
	}
	// 限制3分钟内每个邮箱最多请求1次确认邮件
	cache.RedisClient.Set(context.Background(), cache.RecentSendUserKey(service.UserEmail), code, time.Minute*3)

	return serializer.Success(nil)
}

// sendConfirmEmailToMQ 将确认邮件发送到消息队列
func (service *UserSendConfirmEmailService) sendConfirmEmailToMQ(targetEmail string, code string) error {
	// 限制1秒超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	sendConfirmEmailReq := task.SendConfirmEmailRequest{
		Email: targetEmail,
		Code:  code,
	}

	body, err := json.Marshal(sendConfirmEmailReq)
	if err != nil {
		logger.Log().Error("[UserSendConfirmEmailService.SendConfirmEmailToMQ] 序列化请求失败: ", err)
		return err
	}
	err = rabbitMQ.SendMessageToMQ(ctx, rabbitMQ.RabbitMqSendEmailQueue, body)
	if err != nil {
		return err
	}

	return nil
}
