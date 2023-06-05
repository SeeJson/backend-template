package model

import (
	modelbase "github.com/SeeJson/account/model/base"
	"github.com/naamancurtis/mongo-go-struct-to-bson/mapper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	CollectionUser = "user"

	ColUserAccount       = "account"
	ColUserPassword      = "password"
	ColUserName          = "name"
	ColUserPasswordReset = "password_reset"
	ColUserDepartment    = "department"
	ColUserRole          = "role"
	ColUserPoliceNumber  = "police_number"
	ColUserPhone         = "phone"
)

type User struct {
	modelbase.DataModel `bson:",inline,flatten"` // data类 inline,flatten（必须有） 字段将使嵌套结构中的所有字段在地图中上移一级，以位于更高的级别

	Account       string             `bson:"account"`        // 登录账号
	Password      string             `bson:"password"`       // 登录密码
	Name          string             `bson:"name"`           // 显示名
	PasswordReset bool               `bson:"password_reset"` // 是否已重设密码 ture)已重设 false)未
	Department    primitive.ObjectID `bson:"department"`     // 部门id
	Role          primitive.ObjectID `bson:"role"`           // 角色id
	PoliceNumber  string             `bson:"police_number"`  // 警号
	Phone         string             `bson:"phone"`          // 手机号
}

func NewUserDao() UserDao {
	d := UserDao{}
	d.Coll = &d
	return d
}

// implement interface modelbase.ICollection
type UserDao struct {
	modelbase.DataDao
}

// implement interface modelbase.ICollection
func (d *UserDao) GetCollectionName() string {
	return CollectionUser
}

// implement interface modelbase.ICollection
func (d *UserDao) ToBsonM(model interface{}) bson.M {
	m := model.(User)
	result := mapper.ConvertStructToBSONMap(m, nil)
	return result
}
