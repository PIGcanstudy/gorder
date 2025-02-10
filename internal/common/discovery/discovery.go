package discovery

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type Registry interface {
	// 注册服务，instanceID为服务实例ID，serviceName为服务名称，hostPort为服务地址
	Register(ctx context.Context, instanceID, serviceName, hostPort string) error
	// 注销服务，instanceID为服务实例ID，serviceName为服务名称
	Deregister(ctx context.Context, instanceID, serviceName string) error
	// 发现服务，serviceName为服务名称，返回服务地址列表
	Discover(ctx context.Context, serviceName string) ([]string, error)
	// 健康检查，instanceID为服务实例ID，serviceName为服务名称（探活）
	HealthCheck(instanceID, serviceName string) error
}

// 生成实例ID
func GenerateInstanceID(serviceName string) string {
	x := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	return fmt.Sprintf("%s-%d", serviceName, x)
}
