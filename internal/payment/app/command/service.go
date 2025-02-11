package command

import (
	"context"

	"github.com/PIGcanstudy/gorder/common/genproto/orderpb"
)

// 定义OrderService接口
type OrderService interface {
	UpdateOrder(ctx context.Context, order *orderpb.Order) error
}
