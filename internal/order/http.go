package main

import (
	"github.com/PIGcanstudy/gorder/order/app"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	app app.Application
}

func (server HTTPServer) GetCustomerCustomerIdOrdersOrderId(c *gin.Context, customerId string, orderId string) {
}

func (server HTTPServer) PostCustomerCustomerIdOrders(c *gin.Context, customerId string) {

}
