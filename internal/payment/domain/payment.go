package domain

import (
	"context"

	"github.com/PIGcanstudy/gorder/common/genproto/orderpb"
)

type Processor interface {
	CreatePaymentLink(context.Context, *orderpb.Order) (string, error)
}
