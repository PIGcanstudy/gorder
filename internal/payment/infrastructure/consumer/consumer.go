package consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/PIGcanstudy/gorder/common/broker"
	"github.com/PIGcanstudy/gorder/common/genproto/orderpb"
	"github.com/PIGcanstudy/gorder/payment/app"
	"github.com/PIGcanstudy/gorder/payment/app/command"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
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
	logrus.Infof("Payment receive a message from %s, msg=%v", q.Name, string(msg.Body))

	o := &orderpb.Order{}
	// 反序列化消息
	if err := json.Unmarshal(msg.Body, o); err != nil {
		logrus.Infof("failed to unmarshall msg to order, err=%v", err)
		_ = msg.Nack(false, false) // 发送没有确认消息
		return
	}
	log.Printf("order=%v", o)
	// 发起创建支付连接请求并存储信息
	if _, err := c.app.Commands.CreatePayment.Handle(context.TODO(), command.CreatePayment{Order: o}); err != nil {
		// TODO: retry
		logrus.Infof("failed to create order, err=%v", err)
		_ = msg.Nack(false, false)
		return
	}

	// 处理消息后发送确认消息
	_ = msg.Ack(false)
	logrus.Info("consume success")
}
