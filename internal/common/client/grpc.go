package client

import (
	"context"

	"github.com/PIGcanstudy/gorder/common/discovery"
	"github.com/PIGcanstudy/gorder/common/genproto/orderpb"
	"github.com/PIGcanstudy/gorder/common/genproto/stockpb"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// 用来创建调用stock的grpc客户端
func NewStockGRPCClient(ctx context.Context) (Client stockpb.StockServiceClient, close func() error, err error) {
	grpcAddr, err := discovery.GetServiceAddr(ctx, viper.GetString("stock.service_name"))
	if err != nil {
		return nil, func() error { return nil }, err
	}
	if grpcAddr == "" {
		logrus.Warn("empty grpc addr for stock grpc")
	}
	opts, err := grpcDialOpts()
	if err != nil {
		return nil, func() error { return nil }, err
	}
	conn, err := grpc.NewClient(grpcAddr, opts...)
	if err != nil {
		return nil, func() error { return nil }, err
	}
	return stockpb.NewStockServiceClient(conn), conn.Close, nil
}

func NewOrderGRPCClient(ctx context.Context) (Client orderpb.OrderServiceClient, close func() error, err error) {
	// 首先需要知道order服务的grpc地址
	grpcAddr, err := discovery.GetServiceAddr(ctx, viper.GetString("order.service_name"))

	// 如果order服务的grpc地址为空，则返回错误
	if err != nil {
		return nil, func() error { return nil }, err
	}

	if grpcAddr == "" {
		logrus.Warn("empty grpc addr for order grpc")
	}

	opts, err := grpcDialOpts()
	if err != nil {
		return nil, func() error { return nil }, err
	}
	// 调用生成的grpc代码创建连接
	conn, err := grpc.NewClient(grpcAddr, opts...)
	if err != nil {
		return nil, func() error { return nil }, err
	}
	// 返回生成的grpc客户端代码，以及连接关闭函数
	return orderpb.NewOrderServiceClient(conn), conn.Close, nil
}
func grpcDialOpts() ([]grpc.DialOption, error) {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}, nil
}
