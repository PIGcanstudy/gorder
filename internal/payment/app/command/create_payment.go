package command

import (
	"context"

	"github.com/PIGcanstudy/gorder/common/decorator"
	"github.com/PIGcanstudy/gorder/common/genproto/orderpb"
	"github.com/PIGcanstudy/gorder/payment/domain"
	"github.com/sirupsen/logrus"
)

type CreatePayment struct {
	Order *orderpb.Order
}

type CreatePaymentHandler decorator.CommandHandler[CreatePayment, string]

type createPaymentHandler struct {
	processor domain.Processor
	orderGRPC OrderService
}

func (c createPaymentHandler) Handle(ctx context.Context, cmd CreatePayment) (string, error) {
	// 首先请求支付连接
	link, err := c.processor.CreatePaymentLink(ctx, cmd.Order)
	if err != nil {
		return "", err
	}
	logrus.Infof("create payment link for order: %s success, payment link: %s", cmd.Order.ID, link)

	// 更新订单状态
	newOrder := &orderpb.Order{
		ID:          cmd.Order.ID,
		CustomerID:  cmd.Order.CustomerID,
		Status:      "waiting_for_payment",
		Items:       cmd.Order.Items,
		PaymentLink: link,
	}

	// 向orderGRPC发送更新订单的请求
	err = c.orderGRPC.UpdateOrder(ctx, newOrder)
	return link, err
}

// 创建创建支付命令处理器（包括日志、指标）
func NewCreatePaymentHandler(
	processor domain.Processor,
	orderGRPC OrderService,
	logger *logrus.Entry,
	metricClient decorator.MetricsClient,
) CreatePaymentHandler {
	return decorator.ApplyCommandDecorators[CreatePayment, string](
		createPaymentHandler{processor: processor, orderGRPC: orderGRPC},
		logger,
		metricClient,
	)
}
