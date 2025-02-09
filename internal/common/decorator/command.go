package decorator

import (
	"context"

	"github.com/sirupsen/logrus"
)

// 使用泛型定义CommandHandler接口 接受一个泛型C，返回一个泛型R
type CommandHandler[C any, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}

// 应用CommandHandler装饰器（用来初始化CommandHandler）
// 由于queryLoggingDecorator[C, R any]结构体实现了CommandHandler[C, R]接口，因此可以直接赋值给CommandHandler[C, R]接口变量
func ApplyCommandDecorators[H any, R any](handler CommandHandler[H, R], logger *logrus.Entry, metricsClient MetricsClient) CommandHandler[H, R] {
	return queryLoggingDecorator[H, R]{
		logger: logger,
		base: queryMetricsDecorator[H, R]{
			base:   handler,
			client: metricsClient,
		},
	}
}
