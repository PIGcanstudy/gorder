package logging

import (
	"context"
	"os"
	"time"

	"github.com/PIGcanstudy/gorder/common/tracing"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

// 本文件是为了规范日志格式

// 要么用logging.Infof, Warnf...
// 或者直接加hook，用 logrus.Infof...

func Init() {
	SetFormatter(logrus.StandardLogger())
	logrus.SetLevel(logrus.DebugLevel)
	setOutput(logrus.StandardLogger())
	logrus.AddHook(traceHook{})
}

// 设置日志的输出文件位置
func setOutput(logger *logrus.Logger) {
	var (
		folder    = "./log/"
		filePath  = "app.log"
		errorPath = "errors.log"
	)
	if err := os.MkdirAll(folder, 0750); err != nil && !os.IsExist(err) {
		panic(err)
	}
	file, err := os.OpenFile(folder+filePath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}
	_, err = os.OpenFile(folder+errorPath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}
	// 让日志输出到对应文件中
	logger.SetOutput(file)

	rotateInfo, err := rotatelogs.New(
		folder+filePath+".%Y%m%d",
		rotatelogs.WithLinkName("app.log"),       // 设置软链名为app.log
		rotatelogs.WithMaxAge(7*24*time.Hour),    // 设置日志文件最长存活时间为7天
		rotatelogs.WithRotationTime(1*time.Hour), // 设置日志切割时间为1小时
	)
	if err != nil {
		panic(err)
	}
	rotateError, err := rotatelogs.New(
		folder+errorPath+".%Y%m%d",
		rotatelogs.WithLinkName("errors.log"),    // 设置软链名为errors.log
		rotatelogs.WithMaxAge(7*24*time.Hour),    // 设置日志文件最长存活时间为7天
		rotatelogs.WithRotationTime(1*time.Hour), // 设置日志切割时间为1小时
	)
	// 定义哪个日志level对应哪个切割逻辑
	rotationMap := lfshook.WriterMap{
		logrus.DebugLevel: rotateInfo,
		logrus.InfoLevel:  rotateInfo,
		logrus.WarnLevel:  rotateError,
		logrus.ErrorLevel: rotateError,
		logrus.FatalLevel: rotateError,
		logrus.PanicLevel: rotateError,
	}
	// 它将不同的日志级别（logrus.DebugLevel, logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel）映射到不同的日志切割逻辑（rotateInfo 和 rotateError）。
	logrus.AddHook(lfshook.NewHook(rotationMap, &logrus.JSONFormatter{
		TimestampFormat: time.DateTime,
	}))
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
	// logger.SetFormatter(&prefixed.TextFormatter{
	// 	ForceColors:     true,
	// 	ForceFormatting: true,
	// 	TimestampFormat: time.RFC3339,
	// })
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
