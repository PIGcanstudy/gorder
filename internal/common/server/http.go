package server

import (
	"log"

	"github.com/PIGcanstudy/gorder/common/middleware"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// 初始化并运行Http服务器
func RunHTTPServer(serviceName string, wrapper func(router *gin.Engine)) {
	// 得到服务器地址
	addr := viper.Sub(serviceName).GetString("http-addr")
	if addr == "" {
		panic("empty http address")
	}
	// 初始化并运行Http服务器
	RunHTTPServerOnAddr(addr, wrapper)
}

func RunHTTPServerOnAddr(addr string, wrapper func(*gin.Engine)) {
	apiRouter := gin.New()
	setMiddlewares(apiRouter)
	wrapper(apiRouter)
	apiRouter.Group("/api")

	log.Printf("http server strat, listening on %s\n", addr)
	if err := apiRouter.Run(addr); err != nil {
		panic(err)
	}
}

func setMiddlewares(r *gin.Engine) {
	r.Use(middleware.StructuredLog(logrus.NewEntry(logrus.StandardLogger())))
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware("default_server"))
}
