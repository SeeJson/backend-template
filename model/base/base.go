package modelbase

import (
	"go.mongodb.org/mongo-driver/bson"
)

const (
	// data collection
	ColId         = "_id"
	ColCreator    = "creator"
	ColCreateTime = "create_time"
	ColUpdator    = "updator"
	ColUpdateTime = "update_time"
	ColIsDelete   = "is_delete"

	// meta collection
	ColUid = "uid"
)

type ICollection interface {
	GetCollectionName() string
	ToBsonM(model interface{}) bson.M
}
