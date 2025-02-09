package decorator

import (
	"context"

	"github.com/sirupsen/logrus"
)

// 使用泛型定义QueryHandler接口 接受一个泛型Q，返回一个泛型R
type QueryHandler[Q any, R any] interface {
	Handle(ctx context.Context, query Q) (R, error)
}

// 应用QueryHandler装饰器（用来初始化QueryHandler）
// 由于queryLoggingDecorator[C, R any]结构体实现了QueryHandler[C, R]接口，因此可以直接赋值给QueryHandler[C, R]接口变量
func ApplyQueryDecorators[H any, R any](handler QueryHandler[H, R], logger *logrus.Entry, metricsClient MetricsClient) QueryHandler[H, R] {
	return queryLoggingDecorator[H, R]{
		logger: logger,
		base: queryMetricsDecorator[H, R]{
			base:   handler,
			client: metricsClient,
		},
	}
}
