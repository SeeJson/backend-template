package service

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	radarerror "github.com/SeeJson/account/error"
	"github.com/SeeJson/account/model"
	modelbase "github.com/SeeJson/account/model/base"
	"github.com/SeeJson/account/util/crypt"
	mongodao "github.com/SeeJson/account/util/mongo"
	redisdao "github.com/SeeJson/account/util/redis"
	"github.com/mozillazg/go-pinyin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/encoding/simplifiedchinese"
)

const (
	userVersion = "user_version_%v" // userver_version_{user id}
)

type UserConfig struct {
	MinPasswordCost       int    `mapstructure:"min_password_cost"`
	UserDefaultPassword   string `mapstructure:"user_default_password"`
	MaxPoliceNumberLength int    `mapstructure:"max_police_number_length"`
	MaxNameLength         int    `mapstructure:"max_name_length"`
}

var userCfg UserConfig

func SetUserConfig(c UserConfig) {
	userCfg = c
}

type User struct {
	ME  ME
	Dao model.UserDao
}

func NewUserService(me *ME) User {
	s := User{}
	if me != nil {
		s.ME = *me
	}
	s.Dao = model.NewUserDao()
	return s
}

func (s *User) GetByAccount(account string) (model.User, *radarerror.CommonError) {
	filter := bson.M{
		model.ColUserAccount: account,
	}
	var user model.User
	err := s.Dao.Get(&user, filter)
	if err == mongo.ErrNoDocuments {
		return user, &radarerror.UserNotFound
	} else if err != nil {
		log.Errorf("fail to get user: %v", err)
		return user, &radarerror.InternalServerError
	}
	return user, nil
}

func (s *User) GetByPoliceNumber(policeNumber string) (model.User, *radarerror.CommonError) {
	filter := bson.M{
		model.ColUserPoliceNumber: policeNumber,
	}
	var user model.User
	err := s.Dao.Get(&user, filter)
	if err == mongo.ErrNoDocuments {
		return user, &radarerror.UserNotFound
	} else if err != nil {
		log.Errorf("fail to get user: %v", err)
		return user, &radarerror.InternalServerError
	}
	return user, nil
}

/*
 * 获取user信息
 */
func (s *User) GetById(id primitive.ObjectID) (model.User, *radarerror.CommonError) {
	var user model.User
	err := s.Dao.GetById(&user, id)
	if err == mongo.ErrNoDocuments {
		log.Errorf("user not found: %v", id.Hex())
		return user, &radarerror.UserNotFound
	} else if err != nil {
		log.Errorf("fail to get user: %v", err)
		return user, &radarerror.InternalServerError
	}
	return user, nil
}

type FilterUser struct {
	Department *primitive.ObjectID // 部门id
	Role       *primitive.ObjectID // 角色id
	Name       *string             // 搜索用户名、姓名或警号；模糊匹配
}

func (s *User) ConvertFilter(filter FilterUser) bson.M {
	log.Debug("filter :", filter)
	var mFilter = bson.M{}

	if filter.Department != nil {
		mFilter[model.ColUserDepartment] = *filter.Department
	}
	if filter.Role != nil {
		mFilter[model.ColUserRole] = *filter.Role
	}
	if filter.Name != nil {
		mFilter["$or"] = []bson.M{
			{model.ColUserAccount: bson.M{"$regex": fmt.Sprintf(".*%v*.", *filter.Name)}},
			{model.ColUserName: bson.M{"$regex": fmt.Sprintf(".*%v*.", *filter.Name)}},
			{model.ColUserPoliceNumber: bson.M{"$regex": fmt.Sprintf(".*%v*.", *filter.Name)}},
		}
	}
	return mFilter
}

/*
 * 根据筛选条件获取列表
 * 注意：page是从0开始
 */
func (s *User) Gets(page, pageSize int64, filter FilterUser) ([]model.User, *radarerror.CommonError) {
	mFilter := s.ConvertFilter(filter)
	opts := &options.FindOptions{}
	if pageSize > 0 {
		opts.SetLimit(pageSize)
	}
	opts.SetSkip(page * pageSize)
	opts.SetSort(bson.M{modelbase.ColId: -1})
	var models []model.User
	err := s.Dao.Gets(&models, mFilter, opts)
	if err != nil {
		log.Errorf("fail to get users: %v", err)
		return nil, &radarerror.InternalServerError
	}
	return models, nil
}

/*
 * 根据筛选你条件获取结果集总数
 */
func (s *User) GetCount(filter FilterUser) (int64, *radarerror.CommonError) {
	mFilter := s.ConvertFilter(filter)
	count, err := s.Dao.GetCount(mFilter)
	if err != nil {
		log.Errorf("fail to get user count: %v", err)
		return 0, &radarerror.InternalServerError
	}

	return count, nil
}

/*
 * 获取已逻辑删除的用户总数
 */
func (s *User) GetDeletedCount(filter FilterUser) (int64, *radarerror.CommonError) {
	mFilter := s.ConvertFilter(filter)
	(mFilter)[modelbase.ColIsDelete] = true
	count, err := s.Dao.GetCount(mFilter)
	if err != nil {
		log.Errorf("fail to get user count: %v", err)
		return 0, &radarerror.InternalServerError
	}
	return count, nil
}

/*
 * 获取已逻辑删除的用户
 * 根据筛选条件获取列表
 * 注意：page是从0开始
 */
func (s *User) GetsDeleted(page, pageSize int64, filter FilterUser) ([]model.User, *radarerror.CommonError) {
	mFilter := s.ConvertFilter(filter)
	(mFilter)[modelbase.ColIsDelete] = true
	opts := &options.FindOptions{}
	if pageSize > 0 {
		opts.SetLimit(pageSize)
	}
	opts.SetSkip(page * pageSize)
	opts.SetSort(bson.M{modelbase.ColId: -1})
	var models []model.User
	err := s.Dao.Gets(&models, mFilter, opts)
	if err != nil {
		log.Errorf("fail to get users: %v", err)
		return nil, &radarerror.InternalServerError
	}
	return models, nil
}

/*
 * 添加
 */
func (s *User) Add(user model.User) (primitive.ObjectID, *radarerror.CommonError) {

	// check name
	gbkStr, _ := simplifiedchinese.GBK.NewEncoder().String(user.Name)
	if len(gbkStr) > userCfg.MaxNameLength {
		log.Errorf("name length limit: %v, you enter:%v", userCfg.MaxNameLength, len(user.Name))
		return primitive.NilObjectID, &radarerror.InvalidArgs
	}

	// check account
	if user.Account == "" {
		log.Errorf("account cannot empty")
		return primitive.NilObjectID, &radarerror.InvalidArgs
	}
	// account去重
	_, cerr := s.GetByAccount(user.Account)
	if cerr == nil {
		log.Errorf("duplicated account: %v", user.Account)
		return primitive.NilObjectID, &radarerror.DuplicatedAccount
	} else if cerr != &radarerror.UserNotFound {
		return primitive.NilObjectID, cerr
	}

	// 加密密码
	if user.Password == "" {
		user.Password = crypt.CalMd5(userCfg.UserDefaultPassword) // 默认密码
	}
	pwd, err := bcrypt.GenerateFromPassword([]byte(user.Password), userCfg.MinPasswordCost)
	if err != nil {
		log.Errorf("fail to generate password hash: %v", err)
		return primitive.NilObjectID, &radarerror.InternalServerError
	}
	user.Password = string(pwd)

	// check police_number
	if len(user.PoliceNumber) > userCfg.MaxPoliceNumberLength {
		log.Errorf("police number length limit: %v, you enter:%v", userCfg.MaxPoliceNumberLength, len(user.PoliceNumber))
		return primitive.NilObjectID, &radarerror.InvalidArgs
	}
	// police_number 去重
	if user.PoliceNumber != "" {
		_, cerr = s.GetByPoliceNumber(user.PoliceNumber)
		if cerr == nil {
			log.Errorf("duplicated police_number: %v", user.PoliceNumber)
			return primitive.NilObjectID, &radarerror.DuplicatedPoliceNumber
		} else if cerr != &radarerror.UserNotFound {
			return primitive.NilObjectID, cerr
		}

	}

	// 默认未重设密码
	user.PasswordReset = false

	id, err := s.Dao.Add(s.ME.Id, user)
	if err != nil {
		log.Errorf("fail to add user: %v", err)
		return primitive.NilObjectID, &radarerror.InternalServerError
	}
	return id, nil
}

type SetUser struct {
	Department   *string `bson:"department"`    // 部门id
	Role         *string `bson:"role"`          // 角色id
	PoliceNumber *string `bson:"police_number"` // 警号
	Phone        *string `bson:"phone"`         // 手机号
}

/*
 * 编辑
 */
func (s *User) Update(id primitive.ObjectID, setCVs SetUser) *radarerror.CommonError {
	update := bson.M{"$set": bson.M{}}

	if setCVs.Department != nil {
		deptId := mongodao.Hex2Id(*setCVs.Department)
		update["$set"].(bson.M)[model.ColUserDepartment] = deptId
	}
	if setCVs.Role != nil {
		roleId := mongodao.Hex2Id(*setCVs.Role)
		update["$set"].(bson.M)[model.ColUserRole] = roleId
	}
	if setCVs.PoliceNumber != nil {
		// check police_number
		if len(*setCVs.PoliceNumber) > userCfg.MaxPoliceNumberLength {
			log.Errorf("exceed police number length limit: %v", *setCVs.PoliceNumber)
			return &radarerror.InvalidArgs
		}
		// police_number 去重
		if *setCVs.PoliceNumber != "" {
			_, cerr := s.GetByPoliceNumber(*setCVs.PoliceNumber)
			if cerr == nil {
				log.Errorf("duplicated police_number: %v", *setCVs.PoliceNumber)
				return &radarerror.DuplicatedPoliceNumber
			} else if cerr != &radarerror.UserNotFound {
				return cerr
			}
			update["$set"].(bson.M)[model.ColUserPoliceNumber] = *setCVs.PoliceNumber
		}

	}
	if setCVs.Phone != nil {
		update["$set"].(bson.M)[model.ColUserPhone] = *setCVs.Phone
	}

	_, err := s.Dao.UpdateById(s.ME.Id, id, update)
	if err != nil {
		log.Errorf("fail to update user: %v", err)
		return &radarerror.InternalServerError
	}
	return nil
}

/*
 * 删除
 */
func (s *User) Delete(id primitive.ObjectID) *radarerror.CommonError {
	_, err := s.Dao.DelById(s.ME.Id, id)
	if err != nil {
		log.Debugf("fail to delete user: %v", err)
		return &radarerror.InternalServerError
	}
	return nil
}

/*
 * UpdatePassword 修改密码
 */
func (s *User) UpdatePassword(id primitive.ObjectID, password string, needReset bool) *radarerror.CommonError {
	// 加密密码
	if password == "" {
		password = crypt.CalMd5(userCfg.UserDefaultPassword) // 默认密码
	}
	pwd, err := bcrypt.GenerateFromPassword([]byte(password), userCfg.MinPasswordCost)
	if err != nil {
		log.Errorf("fail to generate password hash: %v", err)
		return &radarerror.InternalServerError
	}

	update := bson.M{
		"$set": bson.M{
			model.ColUserPassword: string(pwd),
		},
	}

	// 是否需要重设密码
	update["$set"].(bson.M)[model.ColUserPasswordReset] = !needReset

	_, err = s.Dao.UpdateById(s.ME.Id, id, update)
	if err != nil {
		log.Errorf("fail to update user: %v", err)
		return &radarerror.InternalServerError
	}

	// refresh version
	RefreshSessionVersion(id)

	return nil
}

/*
 * 把给定角色下的所有用户的session版本号都更新
 */
func (s *User) RefreshSessionVersionByRole(roleId primitive.ObjectID) *radarerror.CommonError {
	users, cerr := s.Gets(0, 0, FilterUser{Role: &roleId})
	if cerr != nil {
		return cerr
	}
	for _, user := range users {
		RefreshSessionVersion(user.Id)
	}
	return nil
}

/***** 辅助函数 *****/
func CheckPassword(user model.User, password string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)
	return err == nil
}

func RefreshSessionVersion(id primitive.ObjectID) int64 {
	log.Debugf("refresh session version: %v", id)
	key := fmt.Sprintf(userVersion, id.Hex())
	return redisdao.IncrBy(key, 1)
}

func IsSessionVersionValid(id primitive.ObjectID, version int64) bool {
	key := fmt.Sprintf(userVersion, id.Hex())
	v, err := redisdao.GetInt64(key)
	if err != nil || v != version {
		return false
	}
	return true
}

func (s *User) GetCreateAccount(name string) (string, *radarerror.CommonError) {
	var nameStr string
	//检查姓名长度
	gbkStr, _ := simplifiedchinese.GBK.NewEncoder().String(name)
	if len(gbkStr) > userCfg.MaxNameLength {
		log.Errorf("name length limit: %v,you enter:%v", userCfg.MaxNameLength, len(name))
		return nameStr, &radarerror.InternalServerError
	}
	nameStr = name2PingYing(name)
	nameStr = strings.ToLower(nameStr)
	if nameStr == "" {
		return nameStr, nil
	}
	var mFilter = bson.M{}
	// $or 查询 全匹配name 和 匹配name{数字}
	mFilter["$or"] = []bson.M{
		{model.ColUserAccount: bson.M{"$regex": fmt.Sprintf("%v\\d", nameStr)}},
		{model.ColUserAccount: nameStr},
	}
	count, err := s.Dao.GetCount(mFilter)
	if err != nil {
		log.Errorf("fail to get user count: %v", err)
		return nameStr, &radarerror.InternalServerError
	}
	if count != 0 {
		nameStr += strconv.Itoa(int(count + 1))
	}
	return nameStr, nil
}

// 名字要么全是汉字，要么全是英文跟数字
func name2PingYing(name string) string {
	// 默认
	a := pinyin.NewArgs()
	a.Fallback = ProcessWithoutPingYing

	strList := pinyin.Pinyin(name, a)
	pingYing := ""
	for _, str := range strList {
		for _, val := range str {
			pingYing += val
		}

	}

	return pingYing
}

func ProcessWithoutPingYing(r rune, a pinyin.Args) []string {

	var res []string

	if unicode.IsLetter(r) {
		res = append(res, string(r))
	}
	return res
}
