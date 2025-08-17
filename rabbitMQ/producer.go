package rabbitMQ

import (
	"context"

	"go-cloud-disk/utils/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

// 发送消息到MQ
func SendMessageToMQ(ctx context.Context, queueName string, body []byte) (err error) {
	ch, err := RabbitMq.Channel()
	if err != nil {
		logger.Log().Error("[SendMessageToMQ] 打开通道失败: %s", err)
		return
	}

	q, _ := ch.QueueDeclare(queueName, true, false, false, false, nil)
	err = ch.PublishWithContext(ctx, "", q.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         body,
	})
	if err != nil {
		logger.Log().Error("[SendMessageToMQ] 发布消息失败: %s", err)
		return
	}
	return
}
