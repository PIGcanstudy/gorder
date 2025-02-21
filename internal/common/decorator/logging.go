package decorator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PIGcanstudy/gorder/common/logging"
	"github.com/sirupsen/logrus"
)

type queryLoggingDecorator[C, R any] struct {
	logger *logrus.Logger     // 打日志用
	base   QueryHandler[C, R] // 被装饰的查询接口
}

// 为查询打日志
func (q queryLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	body, _ := json.Marshal(cmd)
	fields := logrus.Fields{
		"query":      generateActionName(cmd),
		"query_body": string(body),
	}

	defer func() {
		if err == nil {
			logging.Infof(ctx, fields, "%s", "Query execute successfully")
		} else {
			logging.Errorf(ctx, fields, "Failed to execute query, err=%v", err)
		}
	}()
	result, err = q.base.Handle(ctx, cmd)
	return result, err
}

type commandLoggingDecorator[C, R any] struct {
	logger *logrus.Logger
	base   CommandHandler[C, R]
}

func (q commandLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	body, _ := json.Marshal(cmd)
	fields := logrus.Fields{
		"command":      generateActionName(cmd),
		"command_body": string(body),
	}

	defer func() {
		if err == nil {
			logging.Infof(ctx, fields, "%s", "Query execute successfully")
		} else {
			logging.Errorf(ctx, fields, "Failed to execute query, err=%v", err)
		}
	}()
	return q.base.Handle(ctx, cmd)
}

// 执行的时候是 query.XXXXXHandler  此函数是为了打印后半段
func generateActioinName(cmd any) string {
	return strings.Split(fmt.Sprintf("%T", cmd), ".")[1]
}
