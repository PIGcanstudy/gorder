package server

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
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
	wrapper(apiRouter)
	apiRouter.Group("/api")

	log.Printf("http server strat, listening on %s\n", addr)
	if err := apiRouter.Run(addr); err != nil {
		panic(err)
	}
}
