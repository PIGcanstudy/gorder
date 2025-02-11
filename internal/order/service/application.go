package service

import (
	"context"

	"github.com/PIGcanstudy/gorder/common/broker"
	grpcClient "github.com/PIGcanstudy/gorder/common/client"
	"github.com/PIGcanstudy/gorder/common/metrics"
	"github.com/PIGcanstudy/gorder/order/adapters"
	"github.com/PIGcanstudy/gorder/order/adapters/grpc"
	"github.com/PIGcanstudy/gorder/order/app"
	"github.com/PIGcanstudy/gorder/order/app/command"
	"github.com/PIGcanstudy/gorder/order/app/query"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewApplication(ctx context.Context) (app.Application, func()) {
	// 创建grpc客户端
	stockGrpcClient, closeGrpcClient, err := grpcClient.NewStockGRPCClient(ctx)
	if err != nil {
		logrus.Panicf("in NewApplication, NewStockGRPCClient error: %v", err)
	}

	// 连接rabbitmq
	ch, closeConnFn := broker.ConnectRabbitMQ(
		viper.GetString("rabbitmq.user"),
		viper.GetString("rabbitmq.password"),
		viper.GetString("rabbitmq.host"),
		viper.GetString("rabbitmq.port"),
	)

	// 用户自定义的 继承了stockpb.StockServiceClient的结构体，并实现了服务
	stockGRPC := grpc.NewStockGRPC(stockGrpcClient)

	return newApplication(ctx, stockGRPC, ch), func() {
		_ = closeGrpcClient()
		_ = closeConnFn()
		_ = ch.Close()
	}
}

func newApplication(ctx context.Context, stockGRPC query.StockService, ch *amqp.Channel) app.Application {
	orderInmemRepo := adapters.NewMemortOrderRepository()
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.TodoMetrics{}

	return app.Application{
		Commands: app.Commands{
			CreateOrder: command.NewCreateOrderHandler(orderInmemRepo, stockGRPC, ch, logger, metricsClient),
			UpdateOrder: command.NewUpdateOrderHandler(orderInmemRepo, logger, metricsClient),
		},
		Queries: app.Queries{
			GetCustomerOrder: query.NewGetCustomerOrderHandler(orderInmemRepo, logger, metricsClient),
		},
	}
}
