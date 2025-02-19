package logging

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// 执行请求后打印响应日志是通过go的返回值机制加上return处理函数来获得的
// 此函数用于拦截grpc的请求，并记录日志
func GRPCUnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	fields := logrus.Fields{
		Args: req,
	}
	defer func() {
		fields[Response] = resp
		if err != nil {
			fields[Error] = err.Error()
			logf(ctx, logrus.ErrorLevel, fields, "%s", "_grpc_request_out")
		}
	}()
	// 获取grpc中的元数据
	md, exist := metadata.FromIncomingContext(ctx)
	if exist {
		fields["grpc_metadata"] = md
	}

	// 打印请求日志
	logf(ctx, logrus.InfoLevel, fields, "%s", "_grpc_request_in")
	return handler(ctx, req)
}
