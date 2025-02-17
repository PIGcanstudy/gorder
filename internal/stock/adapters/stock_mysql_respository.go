package adapters

import (
	"context"

	"github.com/PIGcanstudy/gorder/stock/entity"
	"github.com/PIGcanstudy/gorder/stock/infrastructure/persistent"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MySQLStockRepository struct {
	db *persistent.MySQL
}

func NewMySQLStockRepository(db *persistent.MySQL) *MySQLStockRepository {
	return &MySQLStockRepository{db: db}
}

func (m MySQLStockRepository) GetItems(ctx context.Context, ids []string) ([]*entity.Item, error) {
	//TODO implement me
	panic("implement me")
}

func (m MySQLStockRepository) GetStock(ctx context.Context, ids []string) ([]*entity.ItemWithQuantity, error) {
	data, err := m.db.BatchGetStockByID(ctx, ids)
	if err != nil {
		return nil, errors.Wrap(err, "BatchGetStockByID error")

	}
	var result []*entity.ItemWithQuantity
	for _, d := range data {
		result = append(result, &entity.ItemWithQuantity{
			ID:       d.ProductID,
			Quantity: d.Quantity,
		})
	}
	return result, nil
}

// 删减库存函数
func (m MySQLStockRepository) UpdateStock(
	ctx context.Context,
	data []*entity.ItemWithQuantity,
	updateFn func(
		ctx context.Context,
		existing []*entity.ItemWithQuantity,
		query []*entity.ItemWithQuantity,
	) ([]*entity.ItemWithQuantity, error),
) error {
	// 开启事务（开启事务的原因是保证数据的一致性）
	return m.db.StartTransaction(func(tx *gorm.DB) (err error) {
		defer func() {
			if err != nil {
				logrus.Warnf("update stock transaction err=%v", err)
			}
		}()
		err = m.updatePessimistic(ctx, tx, data, updateFn)
		//err = m.updateOptimistic(ctx, tx, data, updateFn)
		return err
	})
}

// 乐观锁更新库存函数
func (m MySQLStockRepository) updateOptimistic(
	ctx context.Context,
	tx *gorm.DB,
	data []*entity.ItemWithQuantity,
	updateFn func(ctx context.Context, existing []*entity.ItemWithQuantity, query []*entity.ItemWithQuantity,
	) ([]*entity.ItemWithQuantity, error)) error {
	var dest []*persistent.StockModel
	if err := tx.Model(&persistent.StockModel{}).
		Where("product_id IN (?)", getIDFromEntities(data)).
		Find(&dest).Error; err != nil {
		return errors.Wrap(err, "failed to find data")
	}

	for _, queryData := range data {
		var newestRecord persistent.StockModel
		if err := tx.Model(&persistent.StockModel{}).Where("product_id = ?", queryData.ID).
			First(&newestRecord).Error; err != nil {
			return err
		}
		if err := tx.Model(&persistent.StockModel{}).
			Where("product_id = ? AND version = ? AND quantity - ? >= 0", queryData.ID, newestRecord.Version, queryData.Quantity).
			Updates(map[string]any{
				"quantity": gorm.Expr("quantity - ?", queryData.Quantity),
				"version":  newestRecord.Version + 1,
			}).Error; err != nil {
			return err
		}
	}

	return nil
}

// 悲观锁更新库存函数
func (m MySQLStockRepository) updatePessimistic(
	ctx context.Context,
	tx *gorm.DB,
	data []*entity.ItemWithQuantity,
	updateFn func(ctx context.Context, existing []*entity.ItemWithQuantity, query []*entity.ItemWithQuantity,
	) ([]*entity.ItemWithQuantity, error)) error {
	var dest []*persistent.StockModel
	// 找出与产品id对应的所有产品并加上forUpdate锁
	if err := tx.Table("o_stock").
		Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
		Where("product_id IN (?)", getIDFromEntities(data)).
		Find(&dest).Error; err != nil {

		return errors.Wrap(err, "failed to find data")
	}

	existing := m.unmarshalFromDatabase(dest)
	updated, err := updateFn(ctx, existing, data)
	if err != nil {
		return err
	}

	for _, upd := range updated {
		for _, query := range data {
			if upd.ID == query.ID {
				// 执行更新逻辑
				if err = tx.Table("o_stock").Where("product_id = ? AND quantity - ? >= 0", upd.ID, query.Quantity).
					Update("quantity", gorm.Expr("quantity - ?", query.Quantity)).Error; err != nil {
					return errors.Wrapf(err, "unable to update %s", upd.ID)
				}
			}
		}

	}
	return nil
}

// 将库存数据反序化为[]*entity.ItemWithQuantity形式
func (m MySQLStockRepository) unmarshalFromDatabase(dest []*persistent.StockModel) []*entity.ItemWithQuantity {
	var result []*entity.ItemWithQuantity
	for _, i := range dest {
		result = append(result, &entity.ItemWithQuantity{
			ID:       i.ProductID,
			Quantity: i.Quantity,
		})
	}
	return result
}

func getIDFromEntities(items []*entity.ItemWithQuantity) []string {
	var ids []string
	for _, i := range items {
		ids = append(ids, i.ID)
	}
	return ids
}
