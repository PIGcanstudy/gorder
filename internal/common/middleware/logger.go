package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StructuredLog(l *logrus.Entry) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		// 之前是还没开始请求的处理
		c.Next()
		// C.Next之后即为请求执行完的收尾
		elapsed := time.Since(t)
		l.WithFields(logrus.Fields{
			"time_elapsed_ms": elapsed.Milliseconds(),
			"request_uri":     c.Request.RequestURI,
			"remote_addr":     c.RemoteIP(),
			"client_ip":       c.ClientIP(),
			"full_path":       c.FullPath(),
		}).Info("request_out")
	}
}
