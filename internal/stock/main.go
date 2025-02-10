package main

import (
	"context"
	"log"

	"github.com/PIGcanstudy/gorder/common/config"
	"github.com/PIGcanstudy/gorder/common/discovery"
	"github.com/PIGcanstudy/gorder/common/genproto/stockpb"
	"github.com/PIGcanstudy/gorder/common/server"
	"github.com/PIGcanstudy/gorder/stock/ports"
	"github.com/PIGcanstudy/gorder/stock/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func init() {
	if err := config.NewViperConfig(); err != nil {
		logrus.Fatalf("failed to load config: %v", err)
	}
}

func main() {
	serviceName := viper.GetString("stock.service_name")
	serviceType := viper.GetString("stock.server_to_run")

	log.Printf("starting %s service...", serviceName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deregisterFunc, err := discovery.RegisterToConsul(ctx, serviceName)
	if err != nil {
		logrus.Fatal(err)
	}
	defer deregisterFunc()

	application := service.NewApplication(ctx)
	switch serviceType {
	case "grpc":
		server.RunGrpcServer(serviceName, func(server *grpc.Server) {
			svc := ports.NewGRPCServer(application)
			// 此匿名函数的作用是将实现的服务注册到grpc的服务器上
			stockpb.RegisterStockServiceServer(server, svc)
		})
	case "http":
		// 如果改用http服务可以这样写
	default:
		panic("unexpected service type")
	}
}
