package task

import (
	"context"
	"encoding/json"

	"go-cloud-disk/rabbitMQ"
	"go-cloud-disk/utils"
	"go-cloud-disk/utils/logger"
)

type SendConfirmEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func RunSendConfirmEmail(ctx context.Context) error {
	msgs, err := rabbitMQ.ConsumerMessage(ctx, rabbitMQ.RabbitMqSendEmailQueue)
	if err != nil {
		return err
	}
	var forever chan struct{}

	go func() {
		for msg := range msgs {
			logger.Log().Info("[RunSendConfirmEmail] 收到消息: ", string(msg.Body))

			sendConirmEmailReq := SendConfirmEmailRequest{}
			err = json.Unmarshal(msg.Body, &sendConirmEmailReq)
			if err != nil {
				logger.Log().Error("[RunSendConfirmEmail] 解析消息错误: ", err)
			}

			err = utils.SendConfirmMessage(sendConirmEmailReq.Email, sendConirmEmailReq.Code)
			if err != nil {
				logger.Log().Error("[RunSendConfirmEmail] 发送确认邮件错误: ", err)
			}

			msg.Ack(false)
		}
	}()

	logger.Log().Info("发送确认邮件服务已启动")
	<-forever
	return nil
}
