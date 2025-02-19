package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/PIGcanstudy/gorder/common/broker"
	"github.com/PIGcanstudy/gorder/common/logging"
	"github.com/PIGcanstudy/gorder/order/app"
	"github.com/PIGcanstudy/gorder/order/app/command"
	domain "github.com/PIGcanstudy/gorder/order/domain/order"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

// payment从RabbitMQ中拿到消息，向Stripe发送创建支付连接请求

type Consumer struct {
	app app.Application
}

func NewConsumer(app app.Application) *Consumer {
	return &Consumer{
		app: app,
	}
}

func (c *Consumer) Listen(ch *amqp.Channel) {
	// 创建队列
	q, err := ch.QueueDeclare(broker.EventOrderPaid, true, false, false, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	// 绑定一个交换机给队列
	err = ch.QueueBind(q.Name, "", broker.EventOrderPaid, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	// 消费队列里的消息（从消息队列中取出消息）
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		logrus.Warnf("fail to consume: queue=%s, err=%v", q.Name, err)
	}

	var forever chan struct{}
	// 使用协程来不间断的处理取得的消息
	go func() {
		for msg := range msgs {
			c.handleMessage(ch, msg, q)
		}
	}()

	// 为了让协程不退出 而导致处理消息的协程退出
	<-forever
}

// 处理支付消息
func (c *Consumer) handleMessage(ch *amqp.Channel, msg amqp.Delivery, q amqp.Queue) {
	tr := otel.Tracer("rabbitmq")
	ctx, span := tr.Start(broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers), fmt.Sprintf("rabbitmq.%s.consume", q.Name))
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			logging.Warnf(ctx, nil, "consume failed||from=%s||msg=%+v||err=%v", q.Name, msg, err)
			_ = msg.Nack(false, false)
		} else {
			logging.Infof(ctx, nil, "%s", "consume success")
			_ = msg.Ack(false) // 确认消息以及被消费
		}
	}()

	o := &domain.Order{}
	// 反序列化消息
	if err = json.Unmarshal(msg.Body, o); err != nil {
		err = errors.Wrap(err, "error unmarshal msg.body into domain.order")
		return
	}

	// 更新数据库中的订单信息
	_, err = c.app.Commands.UpdateOrder.Handle(context.Background(), command.UpdateOrder{
		Order: o,
		UpdateFn: func(ctx context.Context, order *domain.Order) (*domain.Order, error) {
			if err := order.IsPaid(); err != nil { // 校验下订单是否被支付
				return nil, err
			}
			return order, nil
		},
	})

	if err != nil {
		logging.Errorf(ctx, nil, "error updating order||orderID=%s||err=%v", o.ID, err)
		if err = broker.HandleRetry(ctx, ch, &msg); err != nil {
			err = errors.Wrapf(err, "retry_error, error handling retry, messageID=%s||err=%v", msg.MessageId, err)
		}
		return
	}

	span.AddEvent("order.updated")
}
