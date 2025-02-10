package broker

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

// 连接rabbitmq
func ConnectRabbitMQ(user, password, host, port string) (*amqp.Channel, func() error) {
	// 拼接服务地址
	address := fmt.Sprintf("amqp://%s:%s@%s:%s", user, password, host, port)
	// 建立连接
	conn, err := amqp.Dial(address)
	if err != nil {
		logrus.Fatal(err)
	}
	// 返回连接的Channel
	ch, err := conn.Channel()
	if err != nil {
		logrus.Fatal(err)
	}

	// 定义两个Exchange，分别用于订单创建（直连点对点）和订单支付（广播）
	err = ch.ExchangeDeclare(EventOrderCreated, "direct", true, false, false, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}
	err = ch.ExchangeDeclare(EventOrderPaid, "fanout", true, false, false, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}
	return ch, conn.Close
}
