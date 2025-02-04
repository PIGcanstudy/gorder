package main

import (
	"github.com/PIGcanstudy/gorder/tree/main/internal/common/server"
	"github.com/PIGcanstudy/gorder/tree/main/internal/order/ports"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	serviceName := viper.GetString("order.service_name")

	server.RunHTTPServer(serviceName, func(router *gin.Engine) {
		ports.RegisterHandlersWithOptions(router, HTTPserver{}, ports.GinServerOptions{
			BaseURL:      "/api",
			Middlewares:  nil,
			ErrorHandler: nil,
		})
	})
}
