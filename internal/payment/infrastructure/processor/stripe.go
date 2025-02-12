package processor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/PIGcanstudy/gorder/common/genproto/orderpb"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/checkout/session"
)

type StripeProcessor struct {
	apiKey string // 配置里的stripe key
}

func NewStripeProcessor(apiKey string) *StripeProcessor {
	if apiKey == "" {
		panic("empty api key")
	}
	stripe.Key = apiKey
	return &StripeProcessor{apiKey: apiKey}
}

const (
	successURL = "http://localhost:8282/success" // 成功创建连接后的支付连接地址
)

// stripe的创建支付连接逻辑
func (s StripeProcessor) CreatePaymentLink(ctx context.Context, order *orderpb.Order) (string, error) {
	var items []*stripe.CheckoutSessionLineItemParams
	for _, item := range order.Items {
		items = append(items, &stripe.CheckoutSessionLineItemParams{
			Price:    stripe.String(item.PriceID),
			Quantity: stripe.Int64(int64(item.Quantity)),
		})
	}

	marshalledItems, _ := json.Marshal(order.Items)
	metadata := map[string]string{
		"orderID":     order.ID,
		"customerID":  order.CustomerID,
		"status":      order.Status,
		"items":       string(marshalledItems),
		"paymentLink": order.PaymentLink,
	}
	params := &stripe.CheckoutSessionParams{
		Metadata:   metadata, // 订单的附加信息
		LineItems:  items,    // 商品的详细信息
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(fmt.Sprintf("%s?customerID=%s&orderID=%s", successURL, order.CustomerID, order.ID)),
	}

	result, err := session.New(params) // 创建一个新的Session
	if err != nil {
		return "", err
	}
	return result.URL, nil // 返回支付连接的地址
}
