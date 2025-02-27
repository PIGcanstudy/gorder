package main

import (
	"fmt"

	"github.com/PIGcanstudy/gorder/common"
	client "github.com/PIGcanstudy/gorder/common/client/order"
	"github.com/PIGcanstudy/gorder/common/consts"
	"github.com/PIGcanstudy/gorder/common/convertor"
	"github.com/PIGcanstudy/gorder/common/handler/errors"
	"github.com/PIGcanstudy/gorder/order/app"
	"github.com/PIGcanstudy/gorder/order/app/command"
	"github.com/PIGcanstudy/gorder/order/app/dto"
	"github.com/PIGcanstudy/gorder/order/app/query"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	common.BaseResponse
	app app.Application
}

func (server HTTPServer) PostCustomerCustomerIdOrders(c *gin.Context, customerId string) {
	var (
		req  client.CreateOrderRequest // 获取请求信息
		resp dto.CreateOrderResponse
		err  error
	)
	defer func() {
		server.Response(c, err, &resp)
	}()

	if err := c.ShouldBindJSON(&req); err != nil {
		err = errors.NewWithError(consts.ErrnoBindRequestError, err)
		return
	}

	if err = server.validate(req); err != nil {
		err = errors.NewWithError(consts.ErrnoRequestValidateError, err)
		return
	}

	result, err := server.app.Commands.CreateOrder.Handle(c.Request.Context(), command.CreateOrder{
		CustomerID: req.CustomerId,
		Items:      convertor.NewItemWithQuantityConvertor().ClientsToEntities(req.Items),
	})

	if err != nil {
		return
	}

	resp = dto.CreateOrderResponse{
		OrderID:     result.OrderID,
		CustomerID:  req.CustomerId,
		RedirectURL: fmt.Sprintf("http://localhost:8282/success?customerID=%s&orderID=%s", req.CustomerId, result.OrderID),
	}
}

func (server HTTPServer) GetCustomerCustomerIdOrdersOrderId(c *gin.Context, customerId string, orderId string) {
	var (
		err  error
		resp interface{}
	)
	defer func() {
		server.Response(c, err, resp)
	}()

	o, err := server.app.Queries.GetCustomerOrder.Handle(c.Request.Context(), query.GetCustomerOrder{
		CustomerID: customerId,
		OrderID:    orderId,
	})
	if err != nil {
		return
	}

	resp = client.Order{
		CustomerId:  o.CustomerID,
		Id:          o.ID,
		Items:       convertor.NewItemConvertor().EntitiesToClients(o.Items),
		PaymentLink: o.PaymentLink,
		Status:      o.Status,
	}
}

// 验证请求信息的数量是否出错了
func (H HTTPServer) validate(req client.CreateOrderRequest) error {
	for _, v := range req.Items {
		if v.Quantity <= 0 {
			return fmt.Errorf("quantity must be positive, got %d from %s", v.Quantity, v.Id)
		}
	}
	return nil
}
