package command

import (
	"context"

	"github.com/PIGcanstudy/gorder/common/consts"
	"github.com/PIGcanstudy/gorder/common/convertor"
	"github.com/PIGcanstudy/gorder/common/decorator"
	"github.com/PIGcanstudy/gorder/common/entity"
	"github.com/PIGcanstudy/gorder/common/logging"
	"github.com/PIGcanstudy/gorder/payment/domain"
	"github.com/sirupsen/logrus"
)

type CreatePayment struct {
	Order *entity.Order
}

type CreatePaymentHandler decorator.CommandHandler[CreatePayment, string]

type createPaymentHandler struct {
	processor domain.Processor
	orderGRPC OrderService
}

func (c createPaymentHandler) Handle(ctx context.Context, cmd CreatePayment) (string, error) {
	var err error
	defer logging.WhenCommandExecute(ctx, "CreatePaymentHandler", cmd, err)

	// 首先请求支付连接
	link, err := c.processor.CreatePaymentLink(ctx, cmd.Order)
	if err != nil {
		return "", err
	}

	// 更新订单状态
	newOrder, err := entity.NewValidOrder(
		cmd.Order.ID,
		cmd.Order.CustomerID,
		consts.OrderStatusWaitingForPayment,
		link,
		cmd.Order.Items,
	)
	if err != nil {
		return "", err
	}

	// 向orderGRPC发送更新订单的请求
	err = c.orderGRPC.UpdateOrder(ctx, convertor.NewOrderConvertor().EntityToProto(newOrder))
	return link, err
}

// 创建创建支付命令处理器（包括日志、指标）
func NewCreatePaymentHandler(
	processor domain.Processor,
	orderGRPC OrderService,
	logger *logrus.Logger,
	metricClient decorator.MetricsClient,
) CreatePaymentHandler {
	return decorator.ApplyCommandDecorators[CreatePayment, string](
		createPaymentHandler{processor: processor, orderGRPC: orderGRPC},
		logger,
		metricClient,
	)
}
