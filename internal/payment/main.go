package main

import (
	"context"

	"github.com/PIGcanstudy/gorder/common/broker"
	"github.com/PIGcanstudy/gorder/common/config"
	"github.com/PIGcanstudy/gorder/common/logging"
	"github.com/PIGcanstudy/gorder/common/server"
	"github.com/PIGcanstudy/gorder/payment/infrastructure/consumer"
	"github.com/PIGcanstudy/gorder/payment/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	logging.Init()
	if err := config.NewViperConfig(); err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serviceType := viper.GetString("payment.server_to_run")
	serviceName := viper.GetString("payment.service_name")
	application, cleanup := service.NewApplication(ctx)
	defer cleanup()

	ch, closeConnFn := broker.ConnectRabbitMQ(
		viper.GetString("rabbitmq.user"),
		viper.GetString("rabbitmq.password"),
		viper.GetString("rabbitmq.host"),
		viper.GetString("rabbitmq.port"),
	)

	paymentHandler := NewPaymentHandler(ch)

	defer func() {
		_ = ch.Close()
		_ = closeConnFn()
	}()

	// 启动协程不断监听 RabbitMQ 队列的消息
	go consumer.NewConsumer(application).Listen(ch)

	switch serviceType {
	case "http":
		server.RunHTTPServer(serviceName, paymentHandler.RegisterRoutes)
	case "grpc":
		logrus.Panic("unsupported server type: grpc")
	default:
		logrus.Panic("unreachable code")
	}
}
