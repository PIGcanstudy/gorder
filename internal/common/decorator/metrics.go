// 一个指标，可以用来计时，或者用来计入函数调用次数

package decorator

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type MetricsClient interface {
	Inc(key string, value int) // 用来上报指标
}
type queryMetricsDecorator[C, R any] struct {
	base   QueryHandler[C, R]
	client MetricsClient
}

// 此函数用来上报一个QueryHandler的调用的执行时间
func (q queryMetricsDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	start := time.Now()
	actionName := strings.ToLower(generateActionName(cmd))
	result, err = q.base.Handle(ctx, cmd)
	defer func() {
		end := time.Since(start)
		q.client.Inc(fmt.Sprintf("querys.%s.duration", actionName), int(end.Seconds())) // 上报查了多少秒
		if err == nil {
			q.client.Inc(fmt.Sprintf("querys.%s.success", actionName), 1) // 上报成功的次数
		} else {
			q.client.Inc(fmt.Sprintf("querys.%s.failure", actionName), 1) // 上报失败的次数
		}
	}()
	return result, err
}

// generateActionName 根据命令生成一个动作名称
func generateActionName[C any](cmd C) string {
	return strings.Split(fmt.Sprintf("%T", cmd), ".")[1]
}
