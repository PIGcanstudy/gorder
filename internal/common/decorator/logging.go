package decorator

import (
	"context"
	"encoding/json"
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
	body, _ := json.Marshal(cmd)
	logger := q.logger.WithFields(logrus.Fields{
		"query":      generateActioinName(cmd),
		"query_body": string(body),
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

type commandLoggingDecorator[C, R any] struct {
	logger *logrus.Entry
	base   CommandHandler[C, R]
}

func (q commandLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	body, _ := json.Marshal(cmd)
	logger := q.logger.WithFields(logrus.Fields{
		"command":      generateActionName(cmd),
		"command_body": string(body),
	})
	logger.Debug("Executing command")
	defer func() {
		if err == nil {
			logger.Info("Command execute successfully")
		} else {
			logger.Error("Failed to execute command", err)
		}
	}()
	return q.base.Handle(ctx, cmd)
}

// 执行的时候是 query.XXXXXHandler  此函数是为了打印后半段
func generateActioinName(cmd any) string {
	return strings.Split(fmt.Sprintf("%T", cmd), ".")[1]
}
