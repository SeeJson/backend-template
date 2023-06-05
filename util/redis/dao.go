package redisdao

import (
	"sync"
	"time"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	PoolSize int    `mapstructure:"pool_size"`
}

var cfg Config

var client *redis.Client
var getClientOnce sync.Once

func SetConfig(c Config) {
	cfg = c
}

func GetClient() *redis.Client {
	// 单例
	getClientOnce.Do(func() {
		log.Debugf("redis address: %v", cfg.Address)

		cli := redis.NewClient(&redis.Options{
			Addr:         cfg.Address,
			Password:     cfg.Password,
			PoolSize:     cfg.PoolSize,
			MaxRetries:   2,
			DialTimeout:  10 * time.Second,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			PoolTimeout:  20 * time.Second,
		})

		if _, err := cli.Ping().Result(); err != nil {
			log.Fatalf("fail to ping redis server: %v", err)
		}

		client = cli
	})
	return client
}

func IncrBy(key string, num int64) int64 {
	newNum, _ := GetClient().IncrBy(key, num).Result()
	return newNum
}

func GetInt64(key string) (int64, error) {
	return GetClient().Get(key).Int64()
}
