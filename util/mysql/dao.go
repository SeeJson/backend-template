package mysqldao

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"sync"
)

type Config struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DbName   string `mapstructure:"db_name"`
	Debug    bool   `mapstructure:"debug"`
}

var cfg Config

var mysqlClient *gorm.DB
var getClientOnce sync.Once

func SetConfig(c Config) {
	cfg = c
}

type Dao struct {
}

func (d *Dao) GetClient() *gorm.DB {
	getClientOnce.Do(func() {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.User, cfg.Password,
			cfg.Host, cfg.Port, cfg.DbName)

		opt := gorm.Config{}

		db, err := gorm.Open(mysql.Open(dsn), &opt)
		if err != nil {
			log.Fatalf("fail to connect mysql server: %v", err)
		}

		if cfg.Debug {
			db = db.Debug()
		}
		mysqlClient = db
	})
	return mysqlClient
}
