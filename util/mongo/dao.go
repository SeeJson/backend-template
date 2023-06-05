package mongodao

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	Addresses   []string `mapstructure:"addresses"`
	User        string   `mapstructure:"user"`
	Password    string   `mapstructure:"password"`
	Database    string   `mapstructure:"database"`
	Options     string   `mapstructure:"options"`
	DialTimeout int64    `mapstructure:"dial_timeout"`
}

var cfg Config

var client *mongo.Client
var getClientOnce sync.Once

func SetConfig(c Config) {
	cfg = c
}

type Dao struct {
}

func (d *Dao) GetClient() *mongo.Client {
	// 单例
	getClientOnce.Do(func() {
		var mongoURL string
		if cfg.User == "" {
			mongoURL = fmt.Sprintf("mongodb://%v/", strings.Join(cfg.Addresses, ","))
		} else {
			mongoURL = fmt.Sprintf("mongodb://%v:%v@%v/admin?%v", cfg.User, cfg.Password, strings.Join(cfg.Addresses, ","), cfg.Options)
		}

		log.Debugf("mongo url: %v", mongoURL)

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.DialTimeout)*time.Second)
		defer cancel()
		cli, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
		if err != nil {
			log.Fatalf("fail to connect mongo server: %v", err)
		}

		err = cli.Ping(context.Background(), nil)
		if err != nil {
			log.Fatalf("fail to ping mongo server: %v", err)
		}
		client = cli
	})
	return client
}

func (d *Dao) GetDatabase() *mongo.Database {
	return d.GetClient().Database(cfg.Database)
}

func (d *Dao) GetCollection(collectionName string) *mongo.Collection {
	return d.GetDatabase().Collection(collectionName)
}

/*
 * 快捷将字符串转换成objectId对象,
 * 如果输出的字符串不符合规定，将返回默认的 ObjectID("000000000000000000000000")，
 * 故，本函数仅适用于将ObjectID("000000000000000000000000")看做无效id的应用场景！！
 */
func Hex2Id(h string) primitive.ObjectID {
	id, err := primitive.ObjectIDFromHex(h)
	if err != nil {
		return primitive.NilObjectID
	}
	return id
}
