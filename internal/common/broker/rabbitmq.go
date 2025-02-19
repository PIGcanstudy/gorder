package broker

import (
	"context"
	"fmt"
	"time"

	"github.com/PIGcanstudy/gorder/common/logging"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
)

const (
	DLX                = "dlx"
	DLQ                = "dlq"
	amqpRetryHeaderKey = "x-retry-count"
)

var (
	maxRetryCount = viper.GetInt64("rabbitmq.max-retry")
)

// 将重试的次数放入到消息的header中，然后加入到context中，方便同步重试次数，同时把span相关信息也放到header中，方便追踪

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
	if err = createDLX(ch); err != nil {
		logrus.Fatal(err)
	}
	return ch, conn.Close
}

func createDLX(ch *amqp.Channel) error {
	q, err := ch.QueueDeclare("share_queue", true, false, false, false, nil)
	if err != nil {
		return err
	}
	err = ch.ExchangeDeclare(DLX, "fanout", true, false, false, false, nil)
	if err != nil {
		return err
	}
	err = ch.QueueBind(q.Name, "", DLX, false, nil)
	if err != nil {
		return err
	}
	_, err = ch.QueueDeclare(DLQ, true, false, false, false, nil)
	return err
}

func HandleRetry(ctx context.Context, ch *amqp.Channel, d *amqp.Delivery) (err error) {
	fields, dLog := logging.WhenRequest(ctx, "HandleRetry", map[string]any{
		"delivery":        d,
		"max_retry_count": maxRetryCount,
	})
	defer dLog(nil, &err)
	if d.Headers == nil {
		d.Headers = amqp.Table{}
	}
	// 从消息的headers中获取消息重试次数
	retryCount, ok := d.Headers[amqpRetryHeaderKey].(int64)
	if !ok {
		retryCount = 0
	}
	retryCount++
	d.Headers[amqpRetryHeaderKey] = retryCount
	fields["retry_count"] = retryCount
	// 达到重试次数上限，放入死信队列
	if retryCount >= maxRetryCount {
		logrus.WithContext(ctx).Infof("moving message %s to dlq", d.MessageId)
		return doPublish(ctx, ch, "", DLQ, false, false, amqp.Publishing{
			Headers:      d.Headers,
			ContentType:  "application/json",
			Body:         d.Body,
			DeliveryMode: amqp.Persistent,
		})
	}
	logrus.WithContext(ctx).Debugf("retring message %s, count=%d", d.MessageId, retryCount)
	time.Sleep(time.Second * time.Duration(retryCount))
	// 消息从哪来就放到哪个队列中
	return doPublish(ctx, ch, "", DLQ, false, false, amqp.Publishing{
		Headers:      d.Headers,
		ContentType:  "application/json",
		Body:         d.Body,
		DeliveryMode: amqp.Persistent,
	})
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
