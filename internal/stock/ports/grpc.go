package ports

import (
	"context"

	"github.com/PIGcanstudy/gorder/common/genproto/stockpb"
	"github.com/PIGcanstudy/gorder/stock/app"
)

type GRPCServer struct {
	app app.Application
}

func NewGRPCServer(app app.Application) *GRPCServer {
	return &GRPCServer{app: app}
}

func (s GRPCServer) GetItems(context.Context, *stockpb.GetItemsRequest) (*stockpb.GetItemsResponse, error) {
	panic("implement me")
}

func (s GRPCServer) CheckIfItemsInStock(context.Context, *stockpb.CheckIfItemsInStockRequest) (*stockpb.CheckIfItemsInStockResponse, error) {
	panic("implement me")
}
