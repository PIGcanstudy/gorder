package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/PIGcanstudy/gorder/common/broker"
	"github.com/PIGcanstudy/gorder/common/genproto/orderpb"
	"github.com/PIGcanstudy/gorder/common/logging"
	"github.com/PIGcanstudy/gorder/payment/app"
	"github.com/PIGcanstudy/gorder/payment/app/command"
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
	q, err := ch.QueueDeclare(broker.EventOrderCreated, true, false, false, false, nil)
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
			c.handleMessage(msg, q, ch)
		}
	}()

	// 为了让协程不退出 而导致处理消息的协程退出
	<-forever
}

func (c *Consumer) handleMessage(msg amqp.Delivery, q amqp.Queue, ch *amqp.Channel) {
	tr := otel.Tracer("rabbitmq")
	ctx, span := tr.Start(broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers), fmt.Sprintf("rabbitmq.%s.consume", q.Name))
	defer span.End()

	logging.Infof(ctx, nil, "Payment receive a message from %s, msg=%v", q.Name, string(msg.Body))
	var err error
	defer func() {
		if err != nil {
			logging.Warnf(ctx, nil, "consume failed||from=%s||msg=%+v||err=%v", q.Name, msg, err)
			_ = msg.Nack(false, false)
		} else {
			logging.Infof(ctx, nil, "%s", "consume success")
			_ = msg.Ack(false)
		}
	}()

	o := &orderpb.Order{}
	// 反序列化消息
	if err = json.Unmarshal(msg.Body, o); err != nil {
		err = errors.Wrap(err, "failed to unmarshall msg to order")
		return
	}

	// 发起创建支付连接请求并存储信息
	if _, err = c.app.Commands.CreatePayment.Handle(ctx, command.CreatePayment{Order: o}); err != nil {
		err = errors.Wrap(err, "failed to create payment")
		if err = broker.HandleRetry(ctx, ch, &msg); err != nil {
			err = errors.Wrapf(err, "retry_error, error handling retry, messageID=%s, err=%v", msg.MessageId, err)
		}
		return
	}

	span.AddEvent("payment.created")

}
