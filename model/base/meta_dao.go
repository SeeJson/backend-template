package modelbase

import (
	"context"

	mongodao "github.com/SeeJson/account/util/mongo"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
 * 元数据库表
 * 不拥有创建者、更新者字段
 */
type MetaModel struct {
	Uid      int64 `json:"uid" bson:"uid"`             // meta表唯一标识
	IsDelete bool  `json:"is_delete" bson:"is_delete"` // 是否已逻辑删除
}

type MetaDao struct {
	mongodao.Dao
	Coll ICollection
}

func (d *MetaDao) GetCollection() *mongo.Collection {
	return d.GetDatabase().Collection(d.Coll.GetCollectionName())
}

// 参数model传指针
func (d *MetaDao) Get(model interface{}, filter bson.M) error {
	if _, ok := filter[ColIsDelete]; !ok {
		filter[ColIsDelete] = false
	}
	log.Debugf("filter: %+v", filter)

	result := d.GetCollection().FindOne(context.Background(), filter)
	if result.Err() != nil {
		return result.Err()
	}
	err := result.Decode(model)
	if err != nil {
		return err
	}

	return nil
}

// 参数models传数组指针
func (d *MetaDao) Gets(models interface{}, filter bson.M, opts ...*options.FindOptions) error {
	if _, ok := filter[ColIsDelete]; !ok {
		filter[ColIsDelete] = false
	}
	log.Debugf("filter: %+v", filter)

	cursor, err := d.GetCollection().Find(context.Background(), filter, opts...)
	if err != nil {
		return err
	}
	err = cursor.All(context.Background(), models)
	if err != nil {
		return err
	}

	return nil
}

// 参数model传指针
func (d *MetaDao) GetByUid(model interface{}, uid int64) error {
	filter := bson.M{
		ColUid:      uid,
		ColIsDelete: false,
	}
	log.Debugf("filter: %+v", filter)

	result := d.GetCollection().FindOne(context.Background(), filter)
	if result.Err() != nil {
		return result.Err()
	}
	err := result.Decode(model)
	if err != nil {
		return err
	}

	return nil
}

// 参数models传数组指针
func (d *MetaDao) GetByUids(models interface{}, uids []int64) error {
	filter := bson.M{
		ColUid: bson.M{
			"$in": uids,
		},
	}
	return d.Gets(models, filter)
}

func (d *MetaDao) GetCount(filter bson.M) (int64, error) {
	if _, ok := filter[ColIsDelete]; !ok {
		filter[ColIsDelete] = false
	}
	log.Debugf("filter: %+v", filter)
	count, err := d.GetCollection().CountDocuments(context.Background(), filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (d *MetaDao) Add(model interface{}) (uid int64, err error) {
	doc := d.Coll.ToBsonM(model)
	doc[ColIsDelete] = false
	log.Debugf("doc: %+v", model)
	uid = doc[ColUid].(int64)

	_, err = d.GetCollection().InsertOne(context.Background(), doc)
	if err != nil {
		return
	}
	return
}

func (d *MetaDao) AddMany(models []interface{}) (uids []int64, err error) {
	docs := make([]interface{}, 0, len(models))
	uids = make([]int64, 0, len(models))
	for _, model := range models {
		doc := d.Coll.ToBsonM(model)
		doc[ColIsDelete] = false
		docs = append(docs, doc)
		uids = append(uids, doc[ColUid].(int64))
	}

	_, err = d.GetCollection().InsertMany(context.Background(), docs)
	if err != nil {
		return
	}
	return
}

func (d *MetaDao) Update(filter bson.M, update bson.M) (modifiedCount int64, err error) {
	if _, ok := filter[ColIsDelete]; !ok {
		filter[ColIsDelete] = false
	}

	res, err := d.GetCollection().UpdateMany(
		context.Background(),
		filter,
		update,
	)
	if err != nil {
		return
	}
	modifiedCount = res.ModifiedCount
	return
}

func (d *MetaDao) UpdateByUid(uid int64, update bson.M) (modifiedCount int64, err error) {
	filter := bson.M{ColUid: uid}
	modifiedCount, err = d.Update(filter, update)
	return
}

func (d *MetaDao) UpdateByUids(uids []int64, update bson.M) (modifiedCount int64, err error) {
	filter := bson.M{ColUid: bson.M{"$in": uids}}
	modifiedCount, err = d.Update(filter, update)
	return
}

/*
 * @Param incCVs: 举例 bson.M{"num1": 2, "num2":3}
 */
func (d *MetaDao) Increase(filter bson.M, incCVs bson.M) (modifiedCount int64, err error) {
	update := bson.M{"$inc": incCVs}
	modifiedCount, err = d.Update(filter, update)
	return
}

/*
 * @Param incCVs: 举例 bson.M{"num1": 2, "num2":3}
 */
func (d *MetaDao) IncreaseByUid(uid int64, incCVs bson.M) (modifiedCount int64, err error) {
	filter := bson.M{ColUid: uid}
	modifiedCount, err = d.Increase(filter, incCVs)
	return
}

/*
 * @Param incCVs: 举例 bson.M{"num1": 2, "num2":3}
 */
func (d *MetaDao) IncreaseByUids(uids []int64, incCVs bson.M) (modifiedCount int64, err error) {
	filter := bson.M{ColId: bson.M{"$in": uids}}
	modifiedCount, err = d.Increase(filter, incCVs)
	return
}

// 创建索引
func (d *MetaDao) CreateIndex(index []mongo.IndexModel) error {
	_, err := d.GetCollection().Indexes().CreateMany(context.Background(), index)
	if err != nil {
		return err
	}
	return nil
}
