package discovery

import (
	"context"
	"time"

	"github.com/PIGcanstudy/gorder/common/discovery/consul"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// 通用的函数其作用是向consul注册服务
func RegisterToConsul(ctx context.Context, serviceName string) (func() error, error) {
	// 首先创建Consul客户端
	registry, err := consul.New(viper.GetString("consul.addr"))
	if err != nil {
		return func() error { return nil }, err
	}

	instanceID := GenerateInstanceID(serviceName)
	// 获取grpc地址（多个微服务之间通信使用的是grpc地址）
	grpcAddr := viper.Sub(serviceName).GetString("grpc-addr")
	// 向consul注册服务
	if err := registry.Register(ctx, instanceID, serviceName, grpcAddr); err != nil {
		return func() error { return nil }, err
	}

	// 注册成功后需要进行心跳检测，为了让consul服务端知道服务还处于健康状态
	go func() {
		for {
			if err := registry.HealthCheck(instanceID, serviceName); err != nil {
				logrus.Panicf("no heartbeat from %s to registry, err=%v", serviceName, err)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	logrus.WithFields(logrus.Fields{
		"serviceName": serviceName,
		"addr":        grpcAddr,
	}).Info("registered to consul")

	// 返回一个注销服务的函数
	return func() error {
		return registry.Deregister(ctx, instanceID, serviceName)
	}, nil
}
