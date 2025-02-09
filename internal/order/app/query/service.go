package query

import (
	"context"

	"github.com/PIGcanstudy/gorder/common/genproto/orderpb"
)

type StockService interface {
	CheckIfItemsInStock(ctx context.Context, items []*orderpb.ItemWithQuantity) ([]*orderpb.Item, error)
	GetItems(ctx context.Context, itemIDs []string) ([]*orderpb.Item, error)
}
