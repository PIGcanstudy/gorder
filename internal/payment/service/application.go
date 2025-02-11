package service

// 此文件是用来初始化一些应用服务(比如grpc服务，RabbitMQ服务，以及数据库服务等)用来创建一个API应用对象
import (
	"context"

	grpcClient "github.com/PIGcanstudy/gorder/common/client"
	"github.com/PIGcanstudy/gorder/common/metrics"
	"github.com/PIGcanstudy/gorder/payment/adapters"
	"github.com/PIGcanstudy/gorder/payment/app"
	"github.com/PIGcanstudy/gorder/payment/app/command"
	"github.com/PIGcanstudy/gorder/payment/domain"
	"github.com/PIGcanstudy/gorder/payment/infrastructure/processor"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewApplication(ctx context.Context) (app.Application, func()) {
	// 初始化grpc客户端
	orderGRPCClient, closeOrderClient, err := grpcClient.NewOrderGRPCClient(ctx)
	if err != nil {
		panic(err)
	}
	orderGRPC := adapters.NewOrderGRPC(orderGRPCClient)

	// 促使话Stripe处理器
	stripeProcessor := processor.NewStripeProcessor(viper.GetString("STRIPE_KEY"))

	return newApplication(ctx, orderGRPC, stripeProcessor), func() {
		_ = closeOrderClient()
	}
}

// 创建一个API应用对象
func newApplication(ctx context.Context, orderGRPC command.OrderService, processor domain.Processor) app.Application {
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{
			CreatePayment: command.NewCreatePaymentHandler(processor, orderGRPC, logger, metricClient),
		},
	}
}
