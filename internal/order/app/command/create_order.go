package command

import (
	"context"
	"fmt"

	"github.com/PIGcanstudy/gorder/common/broker"
	"github.com/PIGcanstudy/gorder/common/convertor"
	"github.com/PIGcanstudy/gorder/common/decorator"
	"github.com/PIGcanstudy/gorder/common/entity"
	"github.com/PIGcanstudy/gorder/common/logging"
	"github.com/PIGcanstudy/gorder/order/app/query"
	domain "github.com/PIGcanstudy/gorder/order/domain/order"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/status"

	"github.com/sirupsen/logrus"
)

type CreateOrder struct {
	CustomerID string
	Items      []*entity.ItemWithQuantity
}

type CreateOrderResult struct {
	OrderID string
}

type CreateOrderHandler decorator.CommandHandler[CreateOrder, *CreateOrderResult]

type createOrderHandler struct {
	orderRepo domain.Repository
	stockGRPC query.StockService // 使用接口而不是具体的实现
	channel   *amqp.Channel
}

func NewCreateOrderHandler(
	orderRepo domain.Repository,
	stockGRPC query.StockService,
	channel *amqp.Channel,
	logger *logrus.Entry,
	metricClient decorator.MetricsClient,
) CreateOrderHandler {
	if orderRepo == nil {
		panic("orderRepo is nil")
	}
	if stockGRPC == nil {
		panic("nil stockGRPC")
	}
	if channel == nil {
		panic("nil channel ")
	}
	return decorator.ApplyCommandDecorators[CreateOrder, *CreateOrderResult](
		createOrderHandler{
			orderRepo: orderRepo,
			stockGRPC: stockGRPC,
			channel:   channel,
		},
		logger,
		metricClient,
	)
}

func (h createOrderHandler) Handle(ctx context.Context, cmd CreateOrder) (*CreateOrderResult, error) {
	var err error
	defer logging.WhenCommandExecute(ctx, "CreateOrderHandler", cmd, err)

	// 首先需要创建消息队列

	t := otel.Tracer("rabbitmq")
	ctx, span := t.Start(ctx, fmt.Sprintf("rabbitmq.%s.publish", broker.EventOrderCreated))
	defer span.End()
	// 验证数据
	validItems, err := h.validate(ctx, cmd.Items)
	if err != nil {
		return nil, err
	}

	// 在发给payment服务之前状态为pending状态
	pendingOrder, err := domain.NewPendingOrder(cmd.CustomerID, validItems)
	if err != nil {
		return nil, err
	}

	order, err := h.orderRepo.Create(ctx, pendingOrder)

	// 开始向RabbitMQ发送消息

	// 序列化要发送的订单消息
	// 发布一个消息到exchange，并指定队列的name
	err = broker.PublishEvent(ctx, broker.PublishEventReq{
		Channel:  h.channel,
		Routing:  broker.Direct,
		Queue:    broker.EventOrderCreated,
		Exchange: "",
		Body:     order,
	})

	if err != nil {
		return nil, errors.Wrapf(err, "publish event error q.Name=%s", broker.EventOrderCreated)
	}

	return &CreateOrderResult{
		OrderID: order.ID,
	}, nil
}

func (c createOrderHandler) validate(ctx context.Context, items []*entity.ItemWithQuantity) ([]*entity.Item, error) {
	if len(items) == 0 { //首先检验长度是不是0
		return nil, errors.New("must have at least one item")
	}
	// 将前端传来的数据中重复的部分合并
	items = packItems(items)
	// 检验库存是否足够
	resp, err := c.stockGRPC.CheckIfItemsInStock(ctx, convertor.NewItemWithQuantityConvertor().EntitiesToProtos(items))
	if err != nil {
		return nil, status.Convert(err).Err()
	}
	return convertor.NewItemConvertor().ProtosToEntities(resp.Items), nil
}

// 将key重复的项目合并
func packItems(items []*entity.ItemWithQuantity) []*entity.ItemWithQuantity {
	merged := make(map[string]int32)
	for _, item := range items {
		merged[item.ID] += item.Quantity
	}
	// 合并后的数据
	var res []*entity.ItemWithQuantity
	for id, quantity := range merged {
		res = append(res, entity.NewItemWithQuantity(id, quantity))
	}
	return res
}
