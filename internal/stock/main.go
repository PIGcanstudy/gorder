package main

import (
	"github.com/PIGcanstudy/gorder/common/genproto/stockpb"
	"github.com/PIGcanstudy/gorder/common/server"
	"github.com/PIGcanstudy/gorder/stock/ports"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func main() {
	serviceName := viper.GetString("stock.service-name")
	serviceType := viper.GetString("stock.service-to-run")

	switch serviceType {
	case "grpc":
		server.RunGrpcServer(serviceName, func(server *grpc.Server) {
			svc := ports.NewGRPCServer()
			// 此匿名函数的作用是注册服务到指定位置上
			stockpb.RegisterStockServiceServer(server, svc)
		})
	case "http":
		// 如果改用http服务可以这样写
	default:
		panic("unexpected service type")
	}
}
