package service

import (
	"context"

	grpcClient "github.com/PIGcanstudy/gorder/common/client"
	"github.com/PIGcanstudy/gorder/common/metrics"
	"github.com/PIGcanstudy/gorder/order/adapters"
	"github.com/PIGcanstudy/gorder/order/adapters/grpc"
	"github.com/PIGcanstudy/gorder/order/app"
	"github.com/PIGcanstudy/gorder/order/app/command"
	"github.com/PIGcanstudy/gorder/order/app/query"
	"github.com/sirupsen/logrus"
)

func NewApplication(ctx context.Context) (app.Application, func()) {
	stockGrpcClient, closeGrpcClient, err := grpcClient.NewGRPCClient(ctx)
	if err != nil {
		panic(err)
	}

	// 用户自定义的 继承了stockpb.StockServiceClient的结构体，并实现了服务
	stockGRPC := grpc.NewStockGRPC(stockGrpcClient)

	return newApplication(ctx, stockGRPC), func() {
		_ = closeGrpcClient()
	}
}

func newApplication(ctx context.Context, stockGRPC query.StockService) app.Application {
	orderInmemRepo := adapters.NewMemortOrderRepository()
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.TodoMetrics{}

	return app.Application{
		Commands: app.Commands{
			CreateOrder: command.NewCreateOrderHandler(orderInmemRepo, stockGRPC, logger, metricsClient),
			UpdateOrder: command.NewUpdateOrderHandler(orderInmemRepo, logger, metricsClient),
		},
		Queries: app.Queries{
			GetCustomerOrder: query.NewGetCustomerOrderHandler(orderInmemRepo, logger, metricsClient),
		},
	}
}
