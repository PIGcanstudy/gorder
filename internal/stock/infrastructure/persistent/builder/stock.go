package builder

import (
	"github.com/PIGcanstudy/gorder/common/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 首先将整个表映射出来
type Stock struct {
	ID        []int64  `json:"ID,omitempty"`
	ProductID []string `json:"product_id,omitempty"`
	Quantity  []int32  `json:"quantity,omitempty"`
	Version   []int64  `json:"version,omitempty"`

	// extend fields
	OrderBy       string `json:"order_by,omitempty"`
	ForUpdateLock bool   `json:"for_update,omitempty"`
}

func NewStock() *Stock {
	return &Stock{}
}

// 实现序列化接口
func (s *Stock) FormatArg() (string, error) {
	return util.MarshalString(s)
}

func (s *Stock) Fill(db *gorm.DB) *gorm.DB {
	db = s.fillWhere(db)
	if s.OrderBy != "" {
		db = db.Order(s.OrderBy)
	}
	return db
}

// 填充查询条件
func (s *Stock) fillWhere(db *gorm.DB) *gorm.DB {
	if len(s.ID) > 0 {
		db = db.Where("id in (?)", s.ID)
	}
	if len(s.ProductID) > 0 {
		db = db.Where("product_id in (?)", s.ProductID)
	}
	if len(s.Version) > 0 {
		db = db.Where("version in (?)", s.Version)
	}
	if len(s.Quantity) > 0 {
		db = s.fillQuantityGT(db)
	}

	if s.ForUpdateLock { // 加forUpdate行锁
		db = db.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate})
	}
	return db
}

// 填充大于等于的库存数量查询条件到查询语句中
func (s *Stock) fillQuantityGT(db *gorm.DB) *gorm.DB {
	db = db.Where("quantity >= ?", s.Quantity)
	return db
}

// 填充id字段
func (s *Stock) IDs(v ...int64) *Stock {
	s.ID = v
	return s
}

// 填充product_id字段
func (s *Stock) ProductIDs(v ...string) *Stock {
	s.ProductID = v
	return s
}

// 填充排序字段
func (s *Stock) Order(v string) *Stock {
	s.OrderBy = v
	return s
}

// 填充version字段
func (s *Stock) Versions(v ...int64) *Stock {
	s.Version = v
	return s
}

// 填充大于等于的库存数量字段
func (s *Stock) QuantityGT(v ...int32) *Stock {
	s.Quantity = v
	return s
}

// 填充forUpdate字段
func (s *Stock) ForUpdate() *Stock {
	s.ForUpdateLock = true
	return s
}
