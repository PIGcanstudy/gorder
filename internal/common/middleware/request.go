package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 请求日志
func RequestLog(l *logrus.Entry) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestIn(c, l)
		defer requestOut(c, l)
		c.Next()
	}
}

// 请求响应日志
func requestOut(c *gin.Context, l *logrus.Entry) {
	response, _ := c.Get("response")
	start, _ := c.Get("request_start")
	startTime := start.(time.Time)
	l.WithContext(c.Request.Context()).WithFields(logrus.Fields{
		"proc_time_ms": time.Since(startTime).Milliseconds(),
		"response":     response,
	}).Info("__request_out")
}

// 请求发起日志
func requestIn(c *gin.Context, l *logrus.Entry) {
	c.Set("request_start", time.Now())
	body := c.Request.Body
	bodyBytes, _ := io.ReadAll(body)                          // 读取请求体后，会将c.Request.Body 的内部状态会被改变
	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes)) // 恢复请求体的状态，使得后续的处理函数能够正常读取请求体的数据
	var compactJson bytes.Buffer
	_ = json.Compact(&compactJson, bodyBytes) // 这行代码则是将读取到的请求体数据进行压缩处理（去除空白字符）
	l.WithContext(c.Request.Context()).WithFields(logrus.Fields{
		"start": time.Now().Unix(),
		"args":  compactJson.String(),
		"from":  c.RemoteIP(),
		"uri":   c.Request.RequestURI,
	}).Info("__request_in")
}
