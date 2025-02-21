package order

import (
	"errors"
	"fmt"
	"slices"

	"github.com/PIGcanstudy/gorder/common/consts"
	"github.com/PIGcanstudy/gorder/common/entity"
	"github.com/stripe/stripe-go/v79"
)

type Order struct {
	ID          string
	CustomerID  string
	Status      string
	PaymentLink string
	Items       []*entity.Item
}

func (o *Order) UpdatePaymentLink(paymentLink string) error {
	//if paymentLink == "" {
	//	return errors.New("cannot update empty paymentLink")
	//}
	o.PaymentLink = paymentLink
	return nil
}

func (o *Order) UpdateItems(items []*entity.Item) error {
	o.Items = items
	return nil
}

func (o *Order) UpdateStatus(to string) error {
	if !o.isValidStatusTransition(to) {
		return fmt.Errorf("cannot transit from '%s' to '%s'", o.Status, to)
	}
	o.Status = to
	return nil
}

// 状态校验
func (o *Order) isValidStatusTransition(to string) bool {
	switch o.Status {
	default:
		return false
	case consts.OrderStatusPending: // 判断是否在可达状态里
		return slices.Contains([]string{consts.OrderStatusWaitingForPayment}, to)
	case consts.OrderStatusWaitingForPayment:
		return slices.Contains([]string{consts.OrderStatusPaid}, to)
	case consts.OrderStatusPaid:
		return slices.Contains([]string{consts.OrderStatusReady}, to)
	}
}

func NewOrder(id, customerID, status, paymentLink string, items []*entity.Item) (*Order, error) {
	if id == "" {
		return nil, errors.New("empty id")
	}
	if customerID == "" {
		return nil, errors.New("empty customerID")
	}
	if status == "" {
		return nil, errors.New("empty status")
	}
	if items == nil {
		return nil, errors.New("empty items")
	}
	return &Order{
		ID:          id,
		CustomerID:  customerID,
		Status:      status,
		PaymentLink: paymentLink,
		Items:       items,
	}, nil
}

// 新建一个正在等待的订单
func NewPendingOrder(customerId string, items []*entity.Item) (*Order, error) {
	if customerId == "" {
		return nil, errors.New("empty customerID")
	}
	if items == nil {
		return nil, errors.New("empty items")
	}
	return &Order{
		CustomerID: customerId,
		Status:     consts.OrderStatusPending,
		Items:      items,
	}, nil
}

func (o *Order) IsPaid() error {
	if o.Status == string(stripe.CheckoutSessionPaymentStatusPaid) {
		return nil
	}
	return fmt.Errorf("order status not paid, order id = %s, status = %s", o.ID, o.Status)
}
