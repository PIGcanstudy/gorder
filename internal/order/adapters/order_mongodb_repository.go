package adapters

import (
	"context"
	"time"

	_ "github.com/PIGcanstudy/gorder/common/config"
	"github.com/PIGcanstudy/gorder/common/entity"
	domain "github.com/PIGcanstudy/gorder/order/domain/order"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	dbName   = viper.GetString("mongo.db-name")
	collName = viper.GetString("mongo.coll-name")
)

type OrderRepositoryMongo struct {
	db *mongo.Client
}

func NewOrderRepositoryMongo(db *mongo.Client) *OrderRepositoryMongo {
	return &OrderRepositoryMongo{db: db}
}

// collection 理解为mongdb中的一个表格
func (r *OrderRepositoryMongo) collection() *mongo.Collection {
	return r.db.Database(dbName).Collection(collName)
}

type orderModel struct {
	MongoID     primitive.ObjectID `bson:"_id"` // mongodb创建一个document自动创建的id
	ID          string             `bson:"id"`  // orderid
	CustomerID  string             `bson:"customer_id"`
	Status      string             `bson:"status"`
	PaymentLink string             `bson:"payment_link"`
	Items       []*entity.Item     `bson:"items"`
}

func (r *OrderRepositoryMongo) Create(ctx context.Context, order *domain.Order) (created *domain.Order, err error) {
	defer r.logWithTag("create", err, order, created)
	writeModel := r.marshalToModel(order)
	// collection 理解为mysql中的一个表格
	res, err := r.collection().InsertOne(ctx, writeModel)
	if err != nil {
		return nil, err
	}
	created = order
	created.ID = res.InsertedID.(primitive.ObjectID).Hex()
	return created, nil
}

func (r *OrderRepositoryMongo) logWithTag(tag string, err error, input *domain.Order, result interface{}) {
	l := logrus.WithFields(logrus.Fields{
		"tag":            "order_repository_mongo",
		"input_order":    input,
		"performed_time": time.Now().Unix(),
		"err":            err,
		"result":         result,
	})
	if err != nil {
		l.Infof("%s_fail", tag)
	} else {
		l.Infof("%s_success", tag)
	}
}

func (r *OrderRepositoryMongo) Get(ctx context.Context, id, customerID string) (got *domain.Order, err error) {
	defer r.logWithTag("get", err, nil, got)
	readModel := &orderModel{}
	mongoID, _ := primitive.ObjectIDFromHex(id) // 转换成mongodb的id类型
	cond := bson.M{"_id": mongoID}
	if err = r.collection().FindOne(ctx, cond).Decode(readModel); err != nil {
		return
	}
	if readModel == nil {
		return nil, domain.NotFoundError{OrderID: id}
	}
	got = r.unmarshal(readModel)
	return got, nil
}

// Update 先查找对应的order，然后apply updateFn，再写入回去
func (r *OrderRepositoryMongo) Update(
	ctx context.Context,
	order *domain.Order,
	updateFn func(context.Context, *domain.Order,
	) (*domain.Order, error)) (err error) {
	defer r.logWithTag("update", err, order, nil)
	if order == nil {
		panic("got nil order")
	}
	// 使用mongodb事务
	session, err := r.db.StartSession()
	if err != nil {
		return
	}
	defer session.EndSession(ctx)

	// 开启事务
	if err = session.StartTransaction(); err != nil {
		return err
	}
	defer func() {
		if err == nil {
			_ = session.CommitTransaction(ctx)
		} else {
			_ = session.AbortTransaction(ctx)
		}
	}()

	// inside transaction事务里执行的操作:
	oldOrder, err := r.Get(ctx, order.ID, order.CustomerID)
	if err != nil {
		return
	}
	updated, err := updateFn(ctx, oldOrder)
	if err != nil {
		return
	}
	logrus.Infof("update||oldOrder=%+v||updated=%+v", oldOrder, updated)
	mongoID, _ := primitive.ObjectIDFromHex(oldOrder.ID)
	// 第二个参数表示查询的条件，第三个参数表示更新操作的内容
	res, err := r.collection().UpdateOne(
		ctx,
		bson.M{"_id": mongoID, "customer_id": oldOrder.CustomerID},
		bson.M{"$set": bson.M{
			"status":       updated.Status,
			"payment_link": updated.PaymentLink,
		}},
	)
	if err != nil {
		return
	}
	r.logWithTag("finish_update", err, order, res)
	return
}

func (r *OrderRepositoryMongo) marshalToModel(order *domain.Order) *orderModel {
	return &orderModel{
		MongoID:     primitive.NewObjectID(),
		ID:          order.ID,
		CustomerID:  order.CustomerID,
		Status:      order.Status,
		PaymentLink: order.PaymentLink,
		Items:       order.Items,
	}
}

func (r *OrderRepositoryMongo) unmarshal(m *orderModel) *domain.Order {
	return &domain.Order{
		ID:          m.MongoID.Hex(),
		CustomerID:  m.CustomerID,
		Status:      m.Status,
		PaymentLink: m.PaymentLink,
		Items:       m.Items,
	}
}
