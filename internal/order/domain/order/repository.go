package order

import (
	"context"
	"fmt"
)

type Repository interface {
	Create(context.Context, *Order) (*Order, error)
	Get(ctx context.Context, id, customerID string) (*Order, error)
	Update(
		ctx context.Context,
		o *Order,
		updateFn func(context.Context, *Order) (*Order, error),
	) error // 参数中有一个更新函数，用于更新订单
}

type NotFoundError struct {
	OrderID string
}

// 实现error的接口
func (e NotFoundError) Error() string {
	return fmt.Sprintf("order %s not found", e.OrderID)
}
