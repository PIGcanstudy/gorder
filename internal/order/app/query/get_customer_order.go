package query

import (
	"context"

	"github.com/PIGcanstudy/gorder/common/decorator"
	"github.com/PIGcanstudy/gorder/common/tracing"
	domain "github.com/PIGcanstudy/gorder/order/domain/order"
	"github.com/sirupsen/logrus"
)

type GetCustomerOrder struct {
	CustomerID string
	OrderID    string
}

// 装饰查询处理器
type GetCustomerOrderHandler decorator.QueryHandler[GetCustomerOrder, *domain.Order]

// 实现查询的一个载体（依赖接口，符合依赖倒置原则）
type getCustomerOrderHandler struct {
	orderRepo domain.Repository
}

func NewGetCustomerOrderHandler(orderRepo domain.Repository, logger *logrus.Logger, metricClient decorator.MetricsClient) GetCustomerOrderHandler {

	if orderRepo == nil { // 如果没有仓库，则 panic
		panic("nil orderRepo")
	}
	// 返回一个增强后的查询Handler
	return decorator.ApplyQueryDecorators[GetCustomerOrder, *domain.Order](
		getCustomerOrderHandler{orderRepo: orderRepo},
		logger,
		metricClient,
	)
}
func (g getCustomerOrderHandler) Handle(ctx context.Context, query GetCustomerOrder) (*domain.Order, error) {
	_, span := tracing.Start(ctx, "handle get customer order query")
	// 从仓库中获取一个order
	o, err := g.orderRepo.Get(ctx, query.OrderID, query.CustomerID)
	if err != nil {
		return nil, err
	}
	span.AddEvent("get_success")
	span.End()
	return o, nil
}
