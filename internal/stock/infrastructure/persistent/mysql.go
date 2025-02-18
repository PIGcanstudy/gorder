package persistent

import (
	"context"
	"fmt"
	"time"

	"github.com/PIGcanstudy/gorder/common/logging"
	"github.com/PIGcanstudy/gorder/stock/infrastructure/persistent/builder"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MySQL struct {
	db *gorm.DB
}

func NewMySQLWithDB(db *gorm.DB) *MySQL {
	return &MySQL{db: db}
}

func NewMySQL() *MySQL {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetString("mysql.port"),
		viper.GetString("mysql.dbname"),
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.Panicf("connect to mysql failed, err=%v", err)
	}
	return &MySQL{db: db}
}

type StockModel struct {
	ID        int64     `gorm:"column:id"`
	ProductID string    `gorm:"column:product_id"`
	Quantity  int32     `gorm:"column:quantity"`
	Version   int64     `gorm:"column:version"`
	CreatedAt time.Time `gorm:"column:created_at autoCreateTime"`
	UpdateAt  time.Time `gorm:"column:updated_at autoUpdateTime"`
}

func (StockModel) TableName() string {
	return "o_stock"
}

func (m *StockModel) BeforeCreate(tx *gorm.DB) (err error) {
	m.UpdateAt = time.Now()
	return nil
}

func (d MySQL) GetStockByID(ctx context.Context, query *builder.Stock) (*StockModel, error) {
	_, deferLog := logging.WhenMySQL(ctx, "GetStockByID", query)
	var result StockModel
	tx := query.Fill(d.db.WithContext(ctx)).First(&result)
	defer deferLog(result, &tx.Error)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &result, nil
}

// 如果使用了事务就返回事务的tx，否则返回db
func (d *MySQL) UseTransaction(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return d.db
	}
	return tx
}

func (d MySQL) Update(ctx context.Context, tx *gorm.DB, cond *builder.Stock, update map[string]any) error {
	_, deferLog := logging.WhenMySQL(ctx, "BatchUpdateStock", cond)
	var returning StockModel
	res := cond.Fill(d.UseTransaction(tx).WithContext(ctx).Model(&returning).Clauses(clause.Returning{})).Updates(update)
	defer deferLog(returning, &res.Error)
	return res.Error
}

func (d MySQL) StartTransaction(fc func(tx *gorm.DB) error) error {
	return d.db.Transaction(fc)
}

// 获取库存(将查询条件全部收敛到builder上)
func (d MySQL) BatchGetStockByID(ctx context.Context, query *builder.Stock) ([]StockModel, error) {
	_, deferLog := logging.WhenMySQL(ctx, "BatchGetStockByID", query)
	var result []StockModel
	tx := query.Fill(d.db.WithContext(ctx).Clauses(clause.Returning{}).Find(&result))
	defer deferLog(result, &tx.Error)

	if tx.Error != nil {
		return nil, tx.Error
	}
	return result, nil
}

func (d MySQL) Create(ctx context.Context, tx *gorm.DB, create *StockModel) error {
	_, deferLog := logging.WhenMySQL(ctx, "Create", create)
	var returning StockModel
	err := d.db.WithContext(ctx).Model(&returning).Clauses(clause.Returning{}).Create(create).Error
	defer deferLog(returning, &err)
	return err

}
