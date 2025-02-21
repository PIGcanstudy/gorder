package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/PIGcanstudy/gorder/common/broker"
	"github.com/PIGcanstudy/gorder/common/consts"
	"github.com/PIGcanstudy/gorder/common/convertor"
	"github.com/PIGcanstudy/gorder/common/entity"
	"github.com/PIGcanstudy/gorder/common/genproto/orderpb"
	"github.com/PIGcanstudy/gorder/common/logging"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

type OrderService interface {
	UpdateOrder(ctx context.Context, request *orderpb.Order) error
}

type Consumer struct {
	orderGRPC OrderService
}

func NewConsumer(orderGRPC OrderService) *Consumer {
	return &Consumer{
		orderGRPC: orderGRPC,
	}
}

func (c *Consumer) Listen(ch *amqp.Channel) {
	q, err := ch.QueueDeclare("", true, false, true, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}
	if err = ch.QueueBind(q.Name, "", broker.EventOrderPaid, false, nil); err != nil {
		logrus.Fatal(err)
	}
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		logrus.Warnf("fail to consume: queue=%s, err=%v", q.Name, err)
	}

	var forever chan struct{}
	go func() {
		for msg := range msgs {
			c.handleMessage(ch, msg, q)
		}
	}()
	<-forever
}

func (c *Consumer) handleMessage(ch *amqp.Channel, msg amqp.Delivery, q amqp.Queue) {
	tr := otel.Tracer("rabbitmq")
	ctx, span := tr.Start(broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers), fmt.Sprintf("rabbitmq.%s.consume", q.Name))
	defer span.End()

	var err error
	mqCtx, span := tr.Start(ctx, fmt.Sprintf("rabbitmq.%s.consume", q.Name))
	defer func() {
		span.End()
		if err != nil {
			logging.Warnf(ctx, nil, "consume failed||from=%s||msg=%+v||err=%v", q.Name, msg, err)
			_ = msg.Nack(false, false)
		} else {
			logging.Infof(ctx, nil, "%s", "consume success")
			_ = msg.Ack(false)
		}
	}()

	o := &entity.Order{}
	if err = json.Unmarshal(msg.Body, o); err != nil {
		err = errors.Wrap(err, "error unmarshal msg.body into order")
		return
	}
	if o.Status != "paid" {
		err = errors.New("order not paid, cannot cook")
		return
	}
	cook(ctx, o)
	span.AddEvent(fmt.Sprintf("order_cook: %v", o))
	if err := c.orderGRPC.UpdateOrder(mqCtx, &orderpb.Order{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      consts.OrderStatusReady,
		Items:       convertor.NewItemConvertor().EntitiesToProtos(o.Items),
		PaymentLink: o.PaymentLink,
	}); err != nil {
		logging.Errorf(ctx, nil, "error updating order||orderID=%s||err=%v", o.ID, err)
		if err = broker.HandleRetry(ctx, ch, &msg); err != nil {
			err = errors.Wrapf(err, "retry_error, error handling retry, messageID=%s||err=%v", msg.MessageId, err)
		}
	}
	span.AddEvent("kitchen.order.finished.updated")
}

// 模拟烹饪过程
func cook(ctx context.Context, o *entity.Order) {
	logrus.WithContext(ctx).Printf("cooking order: %s", o.ID)
	time.Sleep(5 * time.Second)
	logrus.WithContext(ctx).Printf("order %s done!", o.ID)
}
