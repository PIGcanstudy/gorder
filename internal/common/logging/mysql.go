package logging

// 本函数用于标准化mysql输出日志

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	Method   = "method"
	Args     = "args"
	Cost     = "cost_ms"
	Response = "response"
	Error    = "err"
)

// ...any 在底层会把接收到的any放到一个切片上
// 放回logrus.Fields类型，以及一个函数，这个函数是用来将查询结果和错误放入传进来的fields中，并输出到日志中
func WhenMySQL(ctx context.Context, method string, args ...any) (logrus.Fields, func(any, *error)) {
	fields := logrus.Fields{
		Method: method,
		Args:   formatMySQLArgs(args),
	}
	start := time.Now()
	return fields, func(resp any, err *error) {
		level, msg := logrus.InfoLevel, "mysql_success"
		fields[Cost] = time.Since(start).Milliseconds()
		fields[Response] = resp

		if err != nil && (*err != nil) {
			level, msg = logrus.ErrorLevel, "mysql_error"
			fields[Error] = (*err).Error()
		}

		logrus.WithContext(ctx).WithFields(fields).Logf(level, "%s", msg)
	}
}

func formatMySQLArgs(args []any) string {
	var item []string
	for _, arg := range args {
		item = append(item, formatMySQLArg(arg))
	}
	// 将切片中的元素用"||"连接，拼成一个字符串
	return strings.Join(item, "||")
}

// 此函数功能是把any转换为json string类型
func formatMySQLArg(arg any) string {
	switch v := arg.(type) {
	default:
		// 序列化为Json
		bytes, err := json.Marshal(v)
		if err != nil {
			return "unsupported type in formatMySQLArg||err=" + err.Error()
		}
		// 转换为Json字符串
		return string(bytes)
	}
}
