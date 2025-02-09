package query

import (
	"context"

	"github.com/PIGcanstudy/gorder/common/genproto/orderpb"
	"github.com/PIGcanstudy/gorder/common/genproto/stockpb"
)

type StockService interface {
	CheckIfItemsInStock(ctx context.Context, items []*orderpb.ItemWithQuantity) (*stockpb.CheckIfItemsInStockResponse, error)
	GetItems(ctx context.Context, itemIDs []string) ([]*orderpb.Item, error)
}
