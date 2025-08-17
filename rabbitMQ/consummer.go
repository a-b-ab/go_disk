package rabbitMQ

import (
	"context"

	"go-cloud-disk/utils/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ConsumerMessage 消费者消息接收函数
// 参数:
//   - ctx: 上下文对象，用于控制goroutine生命周期
//   - queueName: 队列名称，指定要消费的队列
//
// 返回值:
//   - msgs: 消息通道，接收来自队列的消息
//   - err: 错误信息
func ConsumerMessage(ctx context.Context, queueName string) (msgs <-chan amqp.Delivery, err error) {
	// 创建一个新的通道
	ch, err := RabbitMq.Channel()
	if err != nil {
		logger.Log().Error("[ConsumerMessage] 打开通道失败: ", err)
		return nil, err
	}

	// 声明队列，确保队列存在
	// 参数说明: queueName(队列名), durable(持久化), autoDelete(自动删除), exclusive(排他), noWait(不等待), args(其他参数)
	q, _ := ch.QueueDeclare(queueName, true, false, false, false, nil)

	// 消息队列负载均衡
	// 设置每次只预取1条消息，确保消息在消费者之间均匀分配
	err = ch.Qos(1, 0, false)
	if err != nil {
		logger.Log().Error("[ConsumerMessage] 设置Qos失败: ", err)
		return nil, err
	}

	// 开始消费消息
	// 参数说明: queue(队列名), consumer(消费者标识), autoAck(自动确认), exclusive(排他), noLocal(不接收自己发布的消息), noWait(不等待), args(其他参数)
	return ch.Consume(q.Name, "", false, false, false, false, nil)
}
