package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/PIGcanstudy/gorder/common/broker"
	"github.com/PIGcanstudy/gorder/common/genproto/orderpb"
	"github.com/PIGcanstudy/gorder/payment/domain"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/webhook"
	"go.opentelemetry.io/otel"
)

type PanymentHandler struct {
	channel *amqp.Channel
}

func NewPaymentHandler(ch *amqp.Channel) *PanymentHandler {
	return &PanymentHandler{channel: ch}
}

func (h *PanymentHandler) RegisterRoutes(c *gin.Engine) {
	c.POST("/api/webhook", h.handleWebhook)
}

// stripe服务端发送webhook通知到payment服务端，payment服务端收到通知后，根据通知内容，更新订单状态
func (h *PanymentHandler) handleWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logrus.Infof("Error reading request body: %v\n", err)
		c.JSON(http.StatusServiceUnavailable, err.Error())
		return
	}

	log.Println("得到的Sripe secret 是", viper.GetString("ENDPOINT_STRIPE_SECRET"))

	// 验证签名和密钥，构造事件
	event, err := webhook.ConstructEventWithOptions(payload, c.Request.Header.Get("Stripe-Signature"),
		viper.GetString("ENDPOINT_STRIPE_SECRET"), webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		})

	if err != nil {
		logrus.Infof("Error verifying webhook signature: %v\n", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	switch event.Type {
	case stripe.EventTypeCheckoutSessionCompleted: // payment success
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			logrus.Infof("error unmarshal event.data.raw into session, err = %v", err)
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		if session.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid { // 如果已经支付成功，需要更改订单状态
			logrus.Infof("payment for checkout session %v success!", session.ID)
			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel()

			var items []*orderpb.Item
			_ = json.Unmarshal([]byte(session.Metadata["items"]), &items)

			// 序列化并且更改订单状态用来发给MQ
			marshalledOrder, err := json.Marshal(&domain.Order{
				ID:          session.Metadata["orderID"],
				CustomerID:  session.Metadata["customerID"],
				Status:      string(stripe.CheckoutSessionPaymentStatusPaid),
				PaymentLink: session.Metadata["paymentLink"],
				Items:       items,
			})
			if err != nil {
				logrus.Infof("error marshal domain.order, err = %v", err)
				c.JSON(http.StatusBadRequest, err.Error())
				return
			}

			tr := otel.Tracer("rabbitmq")
			mqCtx, span := tr.Start(ctx, fmt.Sprintf("rabbitmq.%s.publish", broker.EventOrderPaid))
			defer span.End()

			headers := broker.InjectRabbitMQHeaders(mqCtx)
			// 发布一个消息给MQ的exchange
			_ = h.channel.PublishWithContext(mqCtx, broker.EventOrderPaid, "", false, false, amqp.Publishing{
				ContentType:  "application/json",
				DeliveryMode: amqp.Persistent,
				Body:         marshalledOrder,
				Headers:      headers,
			})
			logrus.Infof("message published to %s, body: %s", broker.EventOrderPaid, string(marshalledOrder))
		}
	}
	c.JSON(http.StatusOK, nil)
}
