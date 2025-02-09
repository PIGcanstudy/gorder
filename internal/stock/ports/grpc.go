package ports

import (
	"context"

	"github.com/PIGcanstudy/gorder/common/genproto/orderpb"
	"github.com/PIGcanstudy/gorder/common/genproto/stockpb"
	"github.com/PIGcanstudy/gorder/stock/app"
	"github.com/sirupsen/logrus"
)

type GRPCServer struct {
	app app.Application
}

func NewGRPCServer(app app.Application) *GRPCServer {
	return &GRPCServer{app: app}
}

func (s GRPCServer) GetItems(context.Context, *stockpb.GetItemsRequest) (*stockpb.GetItemsResponse, error) {
	logrus.Info("rpc_request_in, stock.GetItems")
	defer func() {
		logrus.Info("rpc_request_out, stock.GetItems")
	}()
	fake := []*orderpb.Item{
		{
			ID:       "fake_id",
			Name:     "fake_name",
			Quantity: 100,
			PriceID:  "fake_price_id",
		},
	}
	return &stockpb.GetItemsResponse{Items: fake}, nil
}

func (s GRPCServer) CheckIfItemsInStock(context.Context, *stockpb.CheckIfItemsInStockRequest) (*stockpb.CheckIfItemsInStockResponse, error) {
	logrus.Info("rpc_request_in, stock.CheckIfItemsInStock")
	defer func() {
		logrus.Info("rpc_request_out, stock.CheckIfItemsInStock")
	}()

	fake := []*orderpb.Item{
		{
			ID:       "fake_id",
			Name:     "fake_name",
			Quantity: 100,
			PriceID:  "fake_price_id",
		},
	}
	return &stockpb.CheckIfItemsInStockResponse{Items: fake, InStock: 100}, nil
}
