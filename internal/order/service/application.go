package service

import (
	"context"

	"github.com/PIGcanstudy/gorder/common/metrics"
	"github.com/PIGcanstudy/gorder/order/adapters"
	"github.com/PIGcanstudy/gorder/order/app"
	"github.com/PIGcanstudy/gorder/order/app/command"
	"github.com/PIGcanstudy/gorder/order/app/query"
	"github.com/sirupsen/logrus"
)

func NewApplication(ctx context.Context) app.Application {
	orderInmemRepo := adapters.NewMemortOrderRepository()
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{
			CreateOrder: command.NewCreateOrderHandler(orderInmemRepo, logger, metricsClient),
			UpdateOrder: command.NewUpdateOrderHandler(orderInmemRepo, logger, metricsClient),
		},
		Queries: app.Queries{
			GetCustomerOrder: query.NewGetCustomerOrderHandler(orderInmemRepo, logger, metricsClient),
		},
	}
}
