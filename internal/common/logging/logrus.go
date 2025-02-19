package logging

import (
	"context"
	"time"

	"github.com/PIGcanstudy/gorder/common/tracing"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

// 本文件是为了规范日志格式

// 要么用logging.Infof, Warnf...
// 或者直接加hook，用 logrus.Infof...

func Init() {
	SetFormatter(logrus.StandardLogger())
	logrus.SetLevel(logrus.DebugLevel)
	logrus.AddHook(traceHook{})
}

func SetFormatter(logger *logrus.Logger) {
	// 自定义日志格式
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyTime:  "time",
			logrus.FieldKeyMsg:   "message",
		},
	})

	// if isLocal, _ := strconv.ParseBool(os.Getenv("LOCAL_ENV")); isLocal {
	logger.SetFormatter(&prefixed.TextFormatter{
		ForceColors:     true,
		ForceFormatting: true,
		TimestampFormat: time.RFC3339,
	})
	// }
}

func logf(ctx context.Context, level logrus.Level, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Logf(level, format, args...)
}

// 打印消息（包括耗时）
func InfofWithCost(ctx context.Context, fields logrus.Fields, start time.Time, format string, args ...any) {
	fields[Cost] = time.Since(start).Milliseconds()
	Infof(ctx, fields, format, args...)
}

// 打印消息
func Infof(ctx context.Context, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Infof(format, args...)
}

// 打印错误消息
func Errorf(ctx context.Context, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Errorf(format, args...)
}

// 打印警告消息
func Warnf(ctx context.Context, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Warnf(format, args...)
}

// 打印致命错误消息
func Panicf(ctx context.Context, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Panicf(format, args...)
}

// logrus hook机制，他会在打印日志之前调用
type traceHook struct{}

// 这个函数表示在哪些日志情况下需要调用
func (t traceHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// 在写日志之前从context中获取跟踪id并加入条目中
func (t traceHook) Fire(entry *logrus.Entry) error {
	if entry.Context != nil {
		entry.Data["trace"] = tracing.TraceID(entry.Context)
		entry = entry.WithTime(time.Now())
	}
	return nil
}
