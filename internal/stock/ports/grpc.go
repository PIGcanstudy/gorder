package ports

import (
	"context"

	"github.com/PIGcanstudy/gorder/common/genproto/stockpb"
	"github.com/PIGcanstudy/gorder/common/tracing"
	"github.com/PIGcanstudy/gorder/stock/app"
	"github.com/PIGcanstudy/gorder/stock/app/query"
	"github.com/PIGcanstudy/gorder/stock/convertor"
)

type GRPCServer struct {
	app app.Application
}

func NewGRPCServer(app app.Application) *GRPCServer {
	return &GRPCServer{app: app}
}

func (s GRPCServer) GetItems(ctx context.Context, request *stockpb.GetItemsRequest) (*stockpb.GetItemsResponse, error) {
	_, span := tracing.Start(ctx, "GetItems")
	defer span.End()

	// 获取所有对应的items
	items, err := s.app.Queries.GetItems.Handle(ctx, query.GetItems{ItemIDs: request.ItemIDs})
	if err != nil {
		return nil, err
	}

	return &stockpb.GetItemsResponse{Items: convertor.NewItemConvertor().EntitiesToProtos(items)}, nil
}

func (s GRPCServer) CheckIfItemsInStock(ctx context.Context, request *stockpb.CheckIfItemsInStockRequest) (*stockpb.CheckIfItemsInStockResponse, error) {
	_, span := tracing.Start(ctx, "CheckIfItemsInStock")
	defer span.End()

	items, err := s.app.Queries.CheckIfItemsInStock.Handle(ctx, query.CheckIfItemsInStock{
		Items: convertor.NewItemWithQuantityConvertor().ProtosToEntities(request.Items),
	})
	if err != nil {
		return nil, err
	}
	return &stockpb.CheckIfItemsInStockResponse{
		InStock: 1,
		Items:   convertor.NewItemConvertor().EntitiesToProtos(items),
	}, nil
}
