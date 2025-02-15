package query

import (
	"context"

	"github.com/PIGcanstudy/gorder/common/decorator"
	"github.com/PIGcanstudy/gorder/common/genproto/orderpb"
	domain "github.com/PIGcanstudy/gorder/stock/domain/stock"
	"github.com/PIGcanstudy/gorder/stock/infrastructure/integration"
	"github.com/sirupsen/logrus"
)

type CheckIfItemsInStock struct {
	Items []*orderpb.ItemWithQuantity
}

type CheckIfItemsInStockHandler decorator.QueryHandler[CheckIfItemsInStock, []*orderpb.Item]

type checkIfItemsInStockHandler struct {
	stockRepo domain.Repository
	stripeAPI *integration.StripeAPI
}

func NewCheckIfItemsInStockHandler(
	stockRepo domain.Repository,
	stripeAPI *integration.StripeAPI,
	logger *logrus.Entry,
	metricClient decorator.MetricsClient,
) CheckIfItemsInStockHandler {
	if stockRepo == nil {
		panic("nil stockRepo")
	}
	if stripeAPI == nil {
		panic("nil stripeAPI")
	}
	return decorator.ApplyQueryDecorators[CheckIfItemsInStock, []*orderpb.Item](
		checkIfItemsInStockHandler{
			stockRepo: stockRepo,
			stripeAPI: stripeAPI,
		},
		logger,
		metricClient,
	)
}

// TODO: 删掉
var stub = map[string]string{
	"1": "price_1QqmtG9Q405C6ViegaDVnhiJ",
}

func (h checkIfItemsInStockHandler) Handle(ctx context.Context, query CheckIfItemsInStock) ([]*orderpb.Item, error) {
	var res []*orderpb.Item
	for _, i := range query.Items {
		// TODO: 改成从数据库或者从stripe中获取
		priceID, err := h.stripeAPI.GetPriceByProductID(ctx, i.ID)
		if err != nil {
			logrus.Warnf("GetPriceByProductID failed, Item ID = %s, err =%v", i.ID, err)
			continue
		}
		res = append(res, &orderpb.Item{
			ID:       i.ID,
			Quantity: i.Quantity,
			PriceID:  priceID,
		})
	}
	return res, nil
}

func getStubPriceID(id string) string {
	priceID, ok := stub[id]
	if !ok {
		priceID = stub["1"]
	}
	return priceID
}
