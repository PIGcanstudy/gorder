package stock

import (
	"context"
	"fmt"
	"strings"

	"github.com/PIGcanstudy/gorder/stock/entity"
)

type Repository interface {
	GetItems(ctx context.Context, ids []string) ([]*entity.Item, error)             // 通过id列表获得商品列表
	GetStock(ctx context.Context, ids []string) ([]*entity.ItemWithQuantity, error) // 通过id列表获得库存列表
	UpdateStock(
		ctx context.Context,
		data []*entity.ItemWithQuantity,
		updateFn func(
			ctx context.Context,
			existing []*entity.ItemWithQuantity,
			query []*entity.ItemWithQuantity,
		) ([]*entity.ItemWithQuantity, error),
	) error
}

type NotFoundError struct {
	Missing []string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("these items not found in stock: %s", strings.Join(e.Missing, ","))
}

// 库存不足错误
type ExceedStockError struct {
	FailedOn []struct {
		ID   string // 商品id
		Want int32  // 需求量
		Have int32  // 库存量
	}
}

func (e ExceedStockError) Error() string {
	var info []string
	for _, v := range e.FailedOn {
		info = append(info, fmt.Sprintf("product_id=%s, want %d, have %d", v.ID, v.Want, v.Have))
	}
	return fmt.Sprintf("not enough stock for [%s]", strings.Join(info, ","))
}
