package adapters

import (
	"context"
	"strconv"
	"sync"
	"time"

	domain "github.com/PIGcanstudy/gorder/order/domain/order"
	"github.com/sirupsen/logrus"
)

type MemoryOrderRepository struct {
	lock  *sync.RWMutex
	store []*domain.Order
}

func NewMemortOrderRepository() *MemoryOrderRepository {
	// 测试用
	s := make([]*domain.Order, 0)
	s = append(s, &domain.Order{
		ID:          "fake-ID",
		CustomerID:  "fake-customerID",
		Status:      "fake-status",
		PaymentLink: "fake-paymentLink",
		Items:       nil,
	})

	return &MemoryOrderRepository{
		lock:  &sync.RWMutex{},
		store: s,
	}
}

// 往内存中创建一个order
func (m *MemoryOrderRepository) Create(_ context.Context, order *domain.Order) (*domain.Order, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	newOrder := &domain.Order{
		ID:          strconv.FormatInt(time.Now().Unix(), 10),
		CustomerID:  order.CustomerID,
		Status:      order.Status,
		PaymentLink: order.PaymentLink,
		Items:       order.Items,
	}
	m.store = append(m.store, newOrder)
	logrus.WithFields(logrus.Fields{
		"input_order":        order,
		"store_after_create": m.store,
	}).Info("memory_order_repo_create")
	return newOrder, nil
}

// 通过customerID与id从内存中获取一个order
func (m *MemoryOrderRepository) Get(_ context.Context, id, customerID string) (*domain.Order, error) {
	// 打印出所有的order
	for i, v := range m.store {
		logrus.Infof("m.store[%d] = %+v", i, v)
	}

	// 读加锁
	m.lock.RLock()
	defer m.lock.RUnlock()
	for _, o := range m.store {
		if o.ID == id && o.CustomerID == customerID {
			logrus.Infof("memory_order_repo_get||found||id=%s||customerID=%s||res=%+v", id, customerID, *o)
			return o, nil
		}
	}
	return nil, domain.NotFoundError{OrderID: id}
}

// 更新store中的某个order
func (m *MemoryOrderRepository) Update(ctx context.Context, order *domain.Order, updateFn func(context.Context, *domain.Order) (*domain.Order, error)) error {
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
