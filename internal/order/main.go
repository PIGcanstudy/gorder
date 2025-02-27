package main

import (
	"context"

	"github.com/PIGcanstudy/gorder/common/broker"
	_ "github.com/PIGcanstudy/gorder/common/config"
	"github.com/PIGcanstudy/gorder/common/discovery"
	"github.com/PIGcanstudy/gorder/common/genproto/orderpb"
	"github.com/PIGcanstudy/gorder/common/logging"
	"github.com/PIGcanstudy/gorder/common/server"
	"github.com/PIGcanstudy/gorder/common/tracing"
	"github.com/PIGcanstudy/gorder/order/infrastructure/consumer"
	"github.com/PIGcanstudy/gorder/order/ports"
	"github.com/PIGcanstudy/gorder/order/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func init() {
	logging.Init()
}

func main() {
	serviceName := viper.GetString("order.service_name")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application, cleanup := service.NewApplication(ctx)
	defer cleanup()

	deregisterFunc, err := discovery.RegisterToConsul(ctx, serviceName)
	if err != nil {
		logrus.Fatalf("failed to register to consul: %v", err)
	}
	defer deregisterFunc()

	shutdown, err := tracing.InitJaegerProvider(viper.GetString("jaeger.url"), serviceName)
	if err != nil {
		logrus.Fatal(err)
	}
	defer shutdown(ctx)

	ch, closeCh := broker.ConnectRabbitMQ(
		viper.GetString("rabbitmq.user"),
		viper.GetString("rabbitmq.password"),
		viper.GetString("rabbitmq.host"),
		viper.GetString("rabbitmq.port"),
	)
	defer func() {
		_ = ch.Close()
		_ = closeCh()
	}()
	go consumer.NewConsumer(application).Listen(ch)

	go server.RunGrpcServer(serviceName, func(s *grpc.Server) {
		svc := ports.NewGRPCServer(application)
		orderpb.RegisterOrderServiceServer(s, svc)
	})

	server.RunHTTPServer(serviceName, func(router *gin.Engine) {
		router.StaticFile("/success", "../../public/success.html")
		ports.RegisterHandlersWithOptions(router, HTTPServer{app: application}, ports.GinServerOptions{
			BaseURL:      "/api",
			Middlewares:  nil,
			ErrorHandler: nil,
		})
	})
}
