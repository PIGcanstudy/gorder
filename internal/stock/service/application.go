package service

import (
	"context"

	"github.com/PIGcanstudy/gorder/common/metrics"
	"github.com/PIGcanstudy/gorder/stock/adapters"
	"github.com/PIGcanstudy/gorder/stock/app"
	"github.com/PIGcanstudy/gorder/stock/app/query"
	"github.com/PIGcanstudy/gorder/stock/infrastructure/integration"
	"github.com/sirupsen/logrus"
)

func NewApplication(ctx context.Context) app.Application {
	stockRepo := adapters.NewMemoryOrderRepository()
	stripeAPI := integration.NewStripeAPI()
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricClient := metrics.TodoMetrics{}

	return app.Application{
		Commands: app.Commands{},
		Queries: app.Queries{
			CheckIfItemsInStock: query.NewCheckIfItemsInStockHandler(stockRepo, stripeAPI, logger, metricClient),
			GetItems:            query.NewGetItemsHandler(stockRepo, logger, metricClient),
		},
	}
}
