package service

import (
	"context"

	"github.com/PIGcanstudy/gorder/common/metrics"
	"github.com/PIGcanstudy/gorder/stock/adapters"
	"github.com/PIGcanstudy/gorder/stock/app"
	"github.com/PIGcanstudy/gorder/stock/app/query"
	"github.com/PIGcanstudy/gorder/stock/infrastructure/integration"
	"github.com/PIGcanstudy/gorder/stock/infrastructure/persistent"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewApplication(ctx context.Context) app.Application {
	//stockRepo := adapters.NewMemoryOrderRepository()
	db := persistent.NewMySQL()
	stockRepo := adapters.NewMySQLStockRepository(db)
	stripeAPI := integration.NewStripeAPI()
	metricsClient := metrics.NewPrometheusMetricsClient(&metrics.PrometheusMetricsClientConfig{
		Host:        viper.GetString("stock.metrics_export_addr"),
		ServiceName: viper.GetString("stock.service-name"),
	})

	return app.Application{
		Commands: app.Commands{},
		Queries: app.Queries{
			CheckIfItemsInStock: query.NewCheckIfItemsInStockHandler(stockRepo, stripeAPI, logrus.StandardLogger(), metricsClient),
			GetItems:            query.NewGetItemsHandler(stockRepo, logrus.StandardLogger(), metricsClient),
		},
	}
}
