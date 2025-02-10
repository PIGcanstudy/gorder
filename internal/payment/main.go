package main

import (
	"github.com/PIGcanstudy/gorder/common/broker"
	"github.com/PIGcanstudy/gorder/common/config"
	"github.com/PIGcanstudy/gorder/common/logging"
	"github.com/PIGcanstudy/gorder/common/server"
	"github.com/PIGcanstudy/gorder/payment/infrastructure/consumer"

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
	serviceType := viper.GetString("payment.server_to_run")
	serviceName := viper.GetString("payment.service_name")

	paymentHandler := NewPaymentHandler()

	ch, closeConnFn := broker.ConnectRabbitMQ(
		viper.GetString("rabbitmq.user"),
		viper.GetString("rabbitmq.password"),
		viper.GetString("rabbitmq.host"),
		viper.GetString("rabbitmq.port"),
	)

	defer func() {
		_ = ch.Close()
		_ = closeConnFn()
	}()

	// 启动协程不断监听 RabbitMQ 队列的消息
	go consumer.NewConsumer().Listen(ch)

	switch serviceType {
	case "http":
		server.RunHTTPServer(serviceName, paymentHandler.RegisterRoutes)
	case "grpc":
		logrus.Panic("unsupported server type: grpc")
	default:
		logrus.Panic("unreachable code")
	}
}
