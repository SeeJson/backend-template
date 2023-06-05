package modelbase

import (
	"context"
	"time"

	mongodao "github.com/SeeJson/account/util/mongo"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
 * 实体数据库表
 * 拥有创建者、更新者字段
 */
type DataModel struct {
	Id         primitive.ObjectID `json:"_id" bson:"_id,omitempty"`       // 主键（务必设置omitempty，让驱动自动生成）
	Creator    primitive.ObjectID `json:"creator" bson:"creator"`         // 创建者
	CreateTime time.Time          `json:"create_time" bson:"create_time"` // 创建时间
	Updator    primitive.ObjectID `json:"updator" bson:"updator"`         // 更新者
	UpdateTime time.Time          `json:"update_time" bson:"update_time"` // 更新时间
	IsDelete   bool               `json:"is_delete" bson:"is_delete"`     // 是否已逻辑删除
}

type DataDao struct {
	mongodao.Dao
	Coll ICollection
}

func (d *DataDao) GetCollection() *mongo.Collection {
	return d.GetDatabase().Collection(d.Coll.GetCollectionName())
}

// 参数model传指针
func (d *DataDao) Get(model interface{}, filter bson.M) error {
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
func (d *DataDao) Gets(models interface{}, filter bson.M, opts ...*options.FindOptions) error {
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
func (d *DataDao) GetById(model interface{}, id primitive.ObjectID) error {
	filter := bson.M{
		ColId:       id,
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
func (d *DataDao) GetByIds(models interface{}, ids []primitive.ObjectID) error {
	filter := bson.M{
		ColId: bson.M{
			"$in": ids,
		},
	}
	return d.Gets(models, filter)
}

func (d *DataDao) GetCount(filter bson.M) (int64, error) {
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

func (d *DataDao) Add(meId primitive.ObjectID, model interface{}) (primitive.ObjectID, error) {
	doc := d.Coll.ToBsonM(model)
	doc[ColIsDelete] = false
	doc[ColCreateTime] = time.Now()
	if meId != primitive.NilObjectID {
		doc[ColCreator] = meId
	}
	log.Debugf("doc: %+v", model)

	result, err := d.GetCollection().InsertOne(context.Background(), doc)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func (d *DataDao) AddMany(meId primitive.ObjectID, models []interface{}) ([]primitive.ObjectID, error) {
	docs := make([]interface{}, 0, len(models))
	for _, model := range models {
		doc := d.Coll.ToBsonM(model)

		doc[ColIsDelete] = false
		doc[ColCreateTime] = time.Now()
		if meId != primitive.NilObjectID {
			doc[ColCreator] = meId
		}
		docs = append(docs, doc)
	}

	result, err := d.GetCollection().InsertMany(context.Background(), docs)
	if err != nil {
		return nil, err
	}
	ids := make([]primitive.ObjectID, 0, len(result.InsertedIDs))
	for _, id := range result.InsertedIDs {
		ids = append(ids, id.(primitive.ObjectID))
	}
	return ids, nil
}

func (d *DataDao) Update(meId primitive.ObjectID, filter bson.M, update bson.M) (modifiedCount int64, err error) {
	if _, ok := filter[ColIsDelete]; !ok {
		filter[ColIsDelete] = false
	}
	if meId != primitive.NilObjectID {
		if _, ok := update["$set"]; !ok {
			update["$set"] = bson.M{}
		}
		update["$set"].(bson.M)[ColUpdator] = meId
		update["$set"].(bson.M)[ColUpdateTime] = time.Now()
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

func (d *DataDao) UpdateById(meId primitive.ObjectID, id primitive.ObjectID, update bson.M) (modifiedCount int64, err error) {
	filter := bson.M{ColId: id}
	modifiedCount, err = d.Update(meId, filter, update)
	return
}

func (d *DataDao) UpdateByIds(meId primitive.ObjectID, ids []primitive.ObjectID, update bson.M) (modifiedCount int64, err error) {
	filter := bson.M{ColId: bson.M{"$in": ids}}
	modifiedCount, err = d.Update(meId, filter, update)
	return
}

func (d *DataDao) Del(meId primitive.ObjectID, filter bson.M) (deletedCount int64, err error) {
	update := bson.M{"$set": bson.M{ColIsDelete: true}}
	deletedCount, err = d.Update(meId, filter, update)
	return
}

func (d *DataDao) DelById(meId primitive.ObjectID, id primitive.ObjectID) (deletedCount int64, err error) {
	filter := bson.M{ColId: id}
	deletedCount, err = d.Del(meId, filter)
	return
}

func (d *DataDao) DelByIds(meId primitive.ObjectID, ids []primitive.ObjectID) (deletedCount int64, err error) {
	filter := bson.M{ColId: bson.M{"$in": ids}}
	deletedCount, err = d.Del(meId, filter)
	return
}

/*
 * @Param incCVs: 举例 bson.M{"num1": 2, "num2":3}
 */
func (d *DataDao) Increase(meId primitive.ObjectID, filter bson.M, incCVs bson.M) (modifiedCount int64, err error) {
	update := bson.M{"$inc": incCVs}
	modifiedCount, err = d.Update(meId, filter, update)
	return
}

/*
 * @Param incCVs: 举例 bson.M{"num1": 2, "num2":3}
 */
func (d *DataDao) IncreaseById(meId primitive.ObjectID, id primitive.ObjectID, incCVs bson.M) (modifiedCount int64, err error) {
	filter := bson.M{ColId: id}
	modifiedCount, err = d.Increase(meId, filter, incCVs)
	return
}

/*
 * @Param incCVs: 举例 bson.M{"num1": 2, "num2":3}
 */
func (d *DataDao) IncreaseByIds(meId primitive.ObjectID, ids []primitive.ObjectID, incCVs bson.M) (modifiedCount int64, err error) {
	filter := bson.M{ColId: bson.M{"$in": ids}}
	modifiedCount, err = d.Increase(meId, filter, incCVs)
	return
}

// 创建索引
func (d *DataDao) CreateIndex(index []mongo.IndexModel) error {
	_, err := d.GetCollection().Indexes().CreateMany(context.Background(), index)
	if err != nil {
		return err
	}
	return nil
}
