package service

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 会话里的操作人信息
type ME struct {
	Id      primitive.ObjectID `json:"id"`       // 用户id
	Version int64              `json:"version"`  // 会话版本号
	AuthMp  map[int64]int64    `json:"auth_map"` // 所拥有的权限集 map的key是权限对象的二进制掩码，value是权限动作的二进制掩码取或

	Account        string             `json:"account"`         // 登录账号
	Name           string             `json:"name"`            // 显示名
	PasswordReset  bool               `json:"password_reset"`  // 是否已重设密码
	Department     primitive.ObjectID `json:"department"`      // 部门id
	DepartmentName string             `json:"department_name"` // 部门名
	Role           primitive.ObjectID `json:"role"`            // 角色id
	RoleName       string             `json:"role_name"`       // 角色名
	PoliceNumber   string             `json:"police_number"`   // 警号
	Phone          string             `json:"phone"`           // 手机号
}

func (m ME) Json() string {
	s, err := json.Marshal(m)
	if err != nil {
		log.Fatal("service.ME to json error")
	}
	return string(s)
}

func LoadME(jsonStr string) (*ME, error) {
	var me ME
	err := json.Unmarshal([]byte(jsonStr), &me)
	if err != nil {
		return nil, err
	}
	return &me, nil
}

type TimeRange struct {
	Start int64 `json:"start" ` // 开始时间戳
	End   int64 `json:"end"`    // 结束时间戳
}
