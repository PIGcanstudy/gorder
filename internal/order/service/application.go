package service

import (
	"context"
	"fmt"
	"time"

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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
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
	mongoClient := newMongoClient()
	orderRepo := adapters.NewOrderRepositoryMongo(mongoClient)

	metricClient := metrics.TodoMetrics{}

	return app.Application{
		Commands: app.Commands{
			CreateOrder: command.NewCreateOrderHandler(orderRepo, stockGRPC, ch, logrus.StandardLogger(), metricClient),
			UpdateOrder: command.NewUpdateOrderHandler(orderRepo, logrus.StandardLogger(), metricClient),
		},
		Queries: app.Queries{
			GetCustomerOrder: query.NewGetCustomerOrderHandler(orderRepo, logrus.StandardLogger(), metricClient),
		},
	}
}

func newMongoClient() *mongo.Client {
	uri := fmt.Sprintf(
		"mongodb://%s:%s@%s:%s",
		viper.GetString("mongo.user"),
		viper.GetString("mongo.password"),
		viper.GetString("mongo.host"),
		viper.GetString("mongo.port"),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := mongo.Connect(ctx, options.Client().ApplyURI(uri)) // 连接mongodb
	if err != nil {
		panic(err)
	}
	if err = c.Ping(ctx, readpref.Primary()); err != nil { // ping mongodb的client
		panic(err)
	}
	return c
}
