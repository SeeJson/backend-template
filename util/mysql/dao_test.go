package mysqldao

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestDao_GetClient(t *testing.T) {
	dbcfg := Config{
		Host:     "127.0.0.1",
		Port:     3306,
		User:     "root",
		Password: "123456",
		DbName:   "gozero",
		Debug:    true,
	}
	SetConfig(dbcfg)
	dao := Dao{}
	dbClint := dao.GetClient()
	var mapArr []map[string]interface{}
	err := dbClint.Raw("select * from book ").Find(&mapArr).Error
	if err != nil {
		log.Errorf("err:%s", err)
	}
	fmt.Println(mapArr)
}
