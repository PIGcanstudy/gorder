package server

import (
	"net"

	"github.com/PIGcanstudy/gorder/common/logging"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_tags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func init() {
	// 新建日志实例
	logger := logrus.New()
	// 设置日志级别
	logger.SetLevel(logrus.WarnLevel)
	// 替换默认的grpc日志
	grpc_logrus.ReplaceGrpcLogger(logrus.NewEntry(logger))
}

func RunGrpcServer(serviceName string, registerServer func(server *grpc.Server)) {
	addr := viper.Sub(serviceName).GetString("grpc-addr")
	if addr == "" {
		addr = viper.GetString("fallback-grpc-addr")
	}
	RunGrpcServerOnAddr(addr, registerServer)
}

func RunGrpcServerOnAddr(addr string, registerServer func(server *grpc.Server)) {
	// 创建logrus条目实例
	logrusEntry := logrus.NewEntry(logrus.StandardLogger())
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			grpc_tags.UnaryServerInterceptor(grpc_tags.WithFieldExtractor(grpc_tags.CodeGenRequestFieldExtractor)),
			grpc_logrus.UnaryServerInterceptor(logrusEntry),
			logging.GRPCUnaryInterceptor, // 自定义的请求拦截处理
		),
		grpc.ChainStreamInterceptor(
			grpc_tags.StreamServerInterceptor(grpc_tags.WithFieldExtractor(grpc_tags.CodeGenRequestFieldExtractor)),
			grpc_logrus.StreamServerInterceptor(logrusEntry),
		),
	)

	registerServer(grpcServer)

	// 创建监听器
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		logrus.Panic(err)
	}

	logrus.Infof("grpc server strat, listening on %s", addr)
	// 启动grpc服务
	if err := grpcServer.Serve(listen); err != nil {
		logrus.Panic(err)
	}
}
