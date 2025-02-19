package adapters

import (
	"context"
	"sync"

	"github.com/PIGcanstudy/gorder/common/entity"
	domain "github.com/PIGcanstudy/gorder/stock/domain/stock"
)

type MemoryStockRepository struct {
	lock  *sync.RWMutex
	store map[string]*entity.Item
}

func (m MemoryStockRepository) GetStock(ctx context.Context, ids []string) ([]*entity.ItemWithQuantity, error) {
	//TODO implement me
	panic("implement me")
}

func (m MemoryStockRepository) UpdateStock(ctx context.Context, data []*entity.ItemWithQuantity, updateFn func(ctx context.Context, existing []*entity.ItemWithQuantity, query []*entity.ItemWithQuantity) ([]*entity.ItemWithQuantity, error)) error {
	//TODO implement me
	panic("implement me")
}

// 假设里面存储了item_id
var stub = map[string]*entity.Item{
	"item_id": {
		ID:       "foo_item",
		Name:     "stub item",
		Quantity: 10000,
		PriceID:  "stub_item_price_id",
	},
	"item1": {
		ID:       "item1",
		Name:     "stub item1",
		Quantity: 10000,
		PriceID:  "stub_item1_price_id",
	},
	"item2": {
		ID:       "item2",
		Name:     "stub item2",
		Quantity: 10000,
		PriceID:  "stub_item2_price_id",
	},
	"item3": {
		ID:       "item3",
		Name:     "stub item3",
		Quantity: 10000,
		PriceID:  "stub_item3_price_id",
	},
}

func NewMemoryOrderRepository() *MemoryStockRepository {
	return &MemoryStockRepository{
		lock:  &sync.RWMutex{},
		store: stub,
	}
}
func (m MemoryStockRepository) GetItems(ctx context.Context, ids []string) ([]*entity.Item, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	var (
		res     []*entity.Item // 找到的订单数据
		missing []string       // 未找到的订单id
	)
	for _, id := range ids {
		if item, exist := m.store[id]; exist {
			res = append(res, item)
		} else {
			missing = append(missing, id)
		}
	}
	if len(res) == len(ids) {
		return res, nil
	}
	return res, domain.NotFoundError{Missing: missing}
}
