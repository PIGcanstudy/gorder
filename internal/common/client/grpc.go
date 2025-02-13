package client

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/PIGcanstudy/gorder/common/discovery"
	"github.com/PIGcanstudy/gorder/common/genproto/orderpb"
	"github.com/PIGcanstudy/gorder/common/genproto/stockpb"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// 用来创建调用stock的grpc客户端
func NewStockGRPCClient(ctx context.Context) (Client stockpb.StockServiceClient, close func() error, err error) {
	if !WaitForStockGRPCServer(viper.GetDuration("dial-grpc-timeout") * time.Second) {
		return nil, nil, errors.New("stock grpc not available")
	}
	grpcAddr, err := discovery.GetServiceAddr(ctx, viper.GetString("stock.service_name"))
	if err != nil {
		return nil, func() error { return nil }, err
	}
	if grpcAddr == "" {
		logrus.Warn("empty grpc addr for stock grpc")
	}
	opts := grpcDialOpts()

	conn, err := grpc.NewClient(grpcAddr, opts...)
	if err != nil {
		return nil, func() error { return nil }, err
	}
	return stockpb.NewStockServiceClient(conn), conn.Close, nil
}

func NewOrderGRPCClient(ctx context.Context) (Client orderpb.OrderServiceClient, close func() error, err error) {
	if !WaitForOrderGRPCServer(viper.GetDuration("dial-grpc-timeout") * time.Second) {
		return nil, nil, errors.New("order grpc not available")
	}
	logrus.Debug("creating order grpc client")
	// 首先需要知道order服务的grpc地址
	grpcAddr, err := discovery.GetServiceAddr(ctx, viper.GetString("order.service_name"))

	// 如果order服务的grpc地址为空，则返回错误
	if err != nil {
		return nil, func() error { return nil }, err
	}

	if grpcAddr == "" {
		logrus.Warn("empty grpc addr for order grpc")
	}

	opts := grpcDialOpts()

	// 调用生成的grpc代码创建连接
	conn, err := grpc.NewClient(grpcAddr, opts...)
	if err != nil {
		return nil, func() error { return nil }, err
	}
	// 返回生成的grpc客户端代码，以及连接关闭函数
	return orderpb.NewOrderServiceClient(conn), conn.Close, nil
}
func grpcDialOpts() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}
}

func WaitForStockGRPCServer(timeout time.Duration) bool {
	logrus.Infof("waiting for stock grpc client, timeout: %v seconds", timeout.Seconds())
	return waitFor(viper.GetString("stock.grpc-addr"), timeout)
}

func WaitForOrderGRPCServer(timeout time.Duration) bool {
	logrus.Infof("waiting for order grpc client, timeout: %v seconds", timeout.Seconds())
	return waitFor(viper.GetString("order.grpc-addr"), timeout)
}

func waitFor(addr string, timeout time.Duration) bool {
	portAlivable := make(chan struct{})
	timeoutCh := time.After(timeout)
	go func() {
		for {
			select {
			case <-timeoutCh:
				return
			default:
				// 继续执行
			}

			_, err := net.Dial("tcp", addr)

			if err == nil {
				close(portAlivable) // 外部的select语句会被唤醒
				return
			}
			time.Sleep(200 * time.Millisecond)
		}
	}()

	select {
	case <-portAlivable:
		return true
	case <-timeoutCh:
		return false
	}
}
