package broker

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
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

// 实现propagation库里面的carrier接口
type RabbitMQHeaderCarrier map[string]interface{}

func (r RabbitMQHeaderCarrier) Get(key string) string {
	value, ok := r[key]
	if !ok {
		return ""
	}
	return value.(string)
}

func (r RabbitMQHeaderCarrier) Set(key string, value string) {
	r[key] = value
}

func (r RabbitMQHeaderCarrier) Keys() []string {
	keys := make([]string, len(r))
	i := 0
	for key := range r {
		keys[i] = key
		i++
	}
	return keys
}

// 将context信息注入到Carrier中
func InjectRabbitMQHeaders(ctx context.Context) map[string]interface{} {
	carrier := make(RabbitMQHeaderCarrier)
	//将context信息注入到carrier中
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return carrier
}

// 将header里的东西加入到context里返回一个新的context
func ExtractRabbitMQHeaders(ctx context.Context, headers map[string]interface{}) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, RabbitMQHeaderCarrier(headers))
}
