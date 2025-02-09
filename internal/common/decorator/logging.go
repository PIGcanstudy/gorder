package decorator

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type queryLoggingDecorator[C, R any] struct {
	logger *logrus.Entry      // 打日志用
	base   QueryHandler[C, R] // 被装饰的查询接口
}

// 为查询打日志
func (q queryLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	logger := q.logger.WithFields(logrus.Fields{
		"query":      generateActioinName(cmd),
		"query_body": fmt.Sprintf("%#v", cmd),
	})
	logger.Debug("Executing query")
	defer func() {
		if err == nil {
			logger.Info("Query executed successfully")
		} else {
			logger.Error("Query execution failed ", err)
		}
	}()
	result, err = q.base.Handle(ctx, cmd)
	return result, err
}

// 执行的时候是 query.XXXXXHandler  此函数是为了打印后半段
func generateActioinName(cmd any) string {
	return strings.Split(fmt.Sprintf("%T", cmd), ".")[1]
}
