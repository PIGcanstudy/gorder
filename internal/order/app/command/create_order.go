package command

import (
	"context"
	"errors"

	"github.com/PIGcanstudy/gorder/common/decorator"
	"github.com/PIGcanstudy/gorder/common/genproto/orderpb"
	"github.com/PIGcanstudy/gorder/order/app/query"
	domain "github.com/PIGcanstudy/gorder/order/domain/order"
	"github.com/sirupsen/logrus"
)

type CreateOrder struct {
	CustomerID string
	Items      []*orderpb.ItemWithQuantity
}

type CreateOrderResult struct {
	OrderID string
}

type CreateOrderHandler decorator.CommandHandler[CreateOrder, *CreateOrderResult]

type createOrderHandler struct {
	orderRepo domain.Repository
	stockGRPC query.StockService // 使用接口而不是具体的实现
}

func NewCreateOrderHandler(
	orderRepo domain.Repository,
	stockGRPC query.StockService,
	logger *logrus.Entry,
	metricClient decorator.MetricsClient,
) CreateOrderHandler {
	if orderRepo == nil {
		panic("orderRepo is nil")
	}
	return decorator.ApplyCommandDecorators[CreateOrder, *CreateOrderResult](
		createOrderHandler{orderRepo: orderRepo, stockGRPC: stockGRPC},
		logger,
		metricClient,
	)
}

func (h createOrderHandler) Handle(ctx context.Context, cmd CreateOrder) (*CreateOrderResult, error) {
	validItems, err := h.validate(ctx, cmd.Items)
	if err != nil {
		return nil, err
	}

	order, err := h.orderRepo.Create(ctx, &domain.Order{
		CustomerID: cmd.CustomerID,
		Items:      validItems,
	})
	if err != nil {
		return nil, err
	}

	return &CreateOrderResult{
		OrderID: order.ID,
	}, nil
}

func (c createOrderHandler) validate(ctx context.Context, items []*orderpb.ItemWithQuantity) ([]*orderpb.Item, error) {
	if len(items) == 0 { //首先检验长度是不是0
		return nil, errors.New("must have at least one item")
	}
	// 将前端传来的数据中重复的部分合并
	items = packItems(items)

	// 检验库存是否足够
	resp, err := c.stockGRPC.CheckIfItemsInStock(ctx, items)
	if err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func packItems(items []*orderpb.ItemWithQuantity) []*orderpb.ItemWithQuantity {
	merged := make(map[string]int32)
	for _, item := range items {
		merged[item.ID] += item.Quantity
	}
	// 合并后的数据
	var res []*orderpb.ItemWithQuantity
	for id, quantity := range merged {
		res = append(res, &orderpb.ItemWithQuantity{
			ID:       id,
			Quantity: quantity,
		})
	}
	return res
}
