package main

import (
	"fmt"
	"net/http"

	"github.com/PIGcanstudy/gorder/common/genproto/orderpb"
	"github.com/PIGcanstudy/gorder/order/app"
	"github.com/PIGcanstudy/gorder/order/app/command"
	"github.com/PIGcanstudy/gorder/order/app/query"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	app app.Application
}

func (server HTTPServer) GetCustomerCustomerIdOrdersOrderId(c *gin.Context, customerId string, orderId string) {
	o, err := server.app.Queries.GetCustomerOrder.Handle(c, query.GetCustomerOrder{
		CustomerID: customerId,
		OrderID:    orderId,
	})
	if err != nil {
		c.JSON(200, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"Order": o,
		},
	})
}

func (server HTTPServer) PostCustomerCustomerIdOrders(c *gin.Context, customerId string) {
	// 获取请求信息
	var req orderpb.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	result, err := server.app.Commands.CreateOrder.Handle(c, command.CreateOrder{
		CustomerID: req.CustomerID,
		Items:      req.Items,
	})

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "success",
		"customer_id":  req.CustomerID,
		"order_id":     result.OrderID,
		"redirect_url": fmt.Sprintf("http://localhost:8282/success?customerID=%s&orderID=%s", req.CustomerID, result.OrderID),
	})
}
