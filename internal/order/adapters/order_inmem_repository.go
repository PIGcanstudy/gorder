package adapters

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/PIGcanstudy/gorder/common/logging"
	domain "github.com/PIGcanstudy/gorder/order/domain/order"
)

type MemoryOrderRepository struct {
	lock  *sync.RWMutex
	store []*domain.Order
}

func NewMemortOrderRepository() *MemoryOrderRepository {
	// 测试用
	s := []*domain.Order{
		{
			ID:          "fake-ID",
			CustomerID:  "fake-customer-id",
			Status:      "fake-status",
			PaymentLink: "fake-payment-link",
			Items:       nil,
		},
	}

	return &MemoryOrderRepository{
		lock:  &sync.RWMutex{},
		store: s,
	}
}

// 往内存中创建一个order
func (m *MemoryOrderRepository) Create(ctx context.Context, order *domain.Order) (created *domain.Order, err error) {
	_, deferLog := logging.WhenRequest(ctx, "MemoryOrderRepository.Create", map[string]any{"order": order})
	defer deferLog(created, &err)
	m.lock.Lock()
	defer m.lock.Unlock()
	newOrder := &domain.Order{
		ID:          strconv.FormatInt(time.Now().Unix(), 10),
		CustomerID:  order.CustomerID,
		Status:      order.Status,
		PaymentLink: order.PaymentLink,
		Items:       order.Items,
	}
	return newOrder, nil
}

// 通过customerID与id从内存中获取一个order
func (m *MemoryOrderRepository) Get(ctx context.Context, id, customerID string) (got *domain.Order, err error) {
	_, deferLog := logging.WhenRequest(ctx, "MemoryOrderRepository.Get", map[string]any{
		"id":         id,
		"customerID": customerID,
	})
	defer deferLog(got, &err)
	// 读加锁
	m.lock.RLock()
	defer m.lock.RUnlock()
	for _, o := range m.store {
		if o.ID == id && o.CustomerID == customerID {
			return o, nil
		}
	}
	return nil, domain.NotFoundError{OrderID: id}
}

// 更新store中的某个order
func (m *MemoryOrderRepository) Update(
	ctx context.Context,
	order *domain.Order,
	updateFn func(context.Context, *domain.Order) (*domain.Order, error),
) (err error) {
	_, deferLog := logging.WhenRequest(ctx, "MemoryOrderRepository.Update", map[string]any{
		"order": order,
	})
	defer deferLog(nil, &err)

	m.lock.Lock()
	defer m.lock.Unlock()
	found := false
	for i, o := range m.store {
		if o.ID == order.ID && o.CustomerID == order.CustomerID {
			found = true
			updatedOrder, err := updateFn(ctx, order) // 这里的作用是什么？可能更新的时候会有连锁的更新
			if err != nil {
				return err
			}
			m.store[i] = updatedOrder
		}
	}
	if !found {
		return domain.NotFoundError{OrderID: order.ID}
	}
	return nil
}
