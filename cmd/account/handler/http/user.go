package httphandler

import (
	"net/http"

	radarerror "github.com/SeeJson/account/error"
	"github.com/SeeJson/account/model"
	"github.com/SeeJson/account/service"
	"github.com/SeeJson/account/util/captcha"
	"github.com/SeeJson/account/util/jwt"
	mongodao "github.com/SeeJson/account/util/mongo"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Request: Login
type ReqLogin struct {
	Account       string `json:"account" binding:"required"`                   // 账号
	Password      string `json:"password" binding:"required"`                  // 密码
	CaptchaId     string `json:"captcha_id,omitempty" binding:"omitempty"`     // 验证码ID
	CaptchaAnswer string `json:"captcha_result,omitempty" binding:"omitempty"` // 验证码
}

// Response: Login
type RspLogin struct {
	NeedReset bool `json:"need_reset"` // 是否需要重设密码
}

// @Summary 登录
// @Description
// @Tags 登录相关
// @Accept application/json
// @Produce application/json
// @Param body body  ReqLogin  true "查询参数"
// @Success 200  {object} radarerror.ResponseWithData{data=RspLogin}
// @Router /api/v3/auth/login [post]
func Login(c *gin.Context) {
	// param
	var req ReqLogin
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Errorf("fail to bind param: %v", err)
		c.Error(&radarerror.InvalidArgs)
		return
	}

	//验证验证码 fixme! 由前端判断是否要验证验证码
	if req.CaptchaId != "" || req.CaptchaAnswer != "" {
		if ok := captcha.Verify(req.CaptchaId, req.CaptchaAnswer); !ok {
			log.Errorf("captcha verify failed!")
			c.Error(&radarerror.InvalidCaptcha)
			return
		}
	}

	var cerr *radarerror.CommonError
	svcUser := service.NewUserService(nil)

	user, cerr := svcUser.GetByAccount(req.Account)
	if cerr == &radarerror.UserNotFound {
		log.Errorf("account not found: %v", err)
		c.Error(&radarerror.AccountNotFound)
		return
	} else if cerr != nil {
		c.Error(cerr)
		return
	}

	// check password
	if !service.CheckPassword(user, req.Password) {
		log.Errorf("password not match")
		c.Error(&radarerror.InvalidPassword)
		return
	}

	// refresh version
	version := service.RefreshSessionVersion(user.Id)

	// session
	authMp := make(map[int64]int64)
	me := service.ME{
		Id:      user.Id,
		Version: version,
		AuthMp:  authMp,

		Account:        user.Account,
		Name:           user.Name,
		PasswordReset:  user.PasswordReset,
		Department:     user.Department,
		DepartmentName: "",
		Role:           user.Role,
		RoleName:       "",
		PoliceNumber:   user.PoliceNumber,
		Phone:          user.Phone,
	}

	// jwt token
	token, err := jwt.GenBase64Token(me.Json())
	if err != nil {
		c.Error(&radarerror.InternalServerError)
		return
	}
	c.Header("Authorization", "Bearer "+token)

	c.JSON(http.StatusOK,
		radarerror.Success.ResponseWithData(RspLogin{
			NeedReset: !user.PasswordReset,
		}),
	)

}

// Response: GenCaptcha
type RspGenCaptcha struct {
	CaptchaId     string `json:"captcha_id"`     // 验证码id
	CaptchaBase64 string `json:"captcha_base64"` // 验证码图片
}

// @Summary 获取验证码
// @Description
// @Tags 登录相关
// @Accept application/json
// @Produce application/json
// @Success 200  {object} radarerror.ResponseWithData{data=RspGenCaptcha}
// @Router /api/v3/auth/captcha [get]
func GenCaptcha(c *gin.Context) {
	id, b64s, err := captcha.Generate()
	if err != nil {
		log.Errorf("fail to generate captcha: %v", err)
		c.Error(&radarerror.InternalServerError)
		return
	}
	c.JSON(http.StatusOK,
		radarerror.Success.ResponseWithData(RspGenCaptcha{
			CaptchaId:     id,
			CaptchaBase64: b64s,
		}),
	)
}

func Logout(c *gin.Context) {
	// todo
}

// @Summary 超管给用户重置初始密码
// @Description 超级管理员重置指定用户的密码为默认密码，用户登录时需要重设密码
// @Tags 用户
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path int true "用户id"
// @Success 200  {object} radarerror.Response
// @Router /api/v3/user/:id/password [put]
func ResetPassword(c *gin.Context) {
	// param
	userId := mongodao.Hex2Id(c.Param("id"))
	if userId == primitive.NilObjectID {
		log.Errorf("invalid id: %v", c.Param("id"))
		c.Error(&radarerror.InvalidArgs)
		return
	}

	// session
	ss, ok := c.Get(SessME)
	if !ok {
		log.Errorf("need login")
		c.Error(&radarerror.Unauthorized)
		return
	}
	me := ss.(service.ME)

	var cerr *radarerror.CommonError
	svcUser := service.NewUserService(&me)

	user, cerr := svcUser.GetById(userId)
	if cerr != nil {
		c.Error(cerr)
		return
	}

	// 检查跨部门权限
	if !CheckAuth(&me, []Auth{{Obj: AuthObjTransDepartment, Act: AuthActGet}}) {
		// 检查是否本部门
		if me.Department != user.Department {
			log.Debugf("different department, me: %v, user: %v", me.Department.Hex(), user.Department.Hex())
			c.Error(&radarerror.ExceedAuthority)
			return
		}
	}

	cerr = svcUser.UpdatePassword(userId, "", true)
	if cerr != nil {
		c.Error(cerr)
		return
	}

	c.JSON(http.StatusOK, radarerror.Success.Response())
}

// Request: UpdateMyPassword
type ReqUpdateMyPassword struct {
	Password string `json:"password" binding:"required"` // 新密码
}

// @Summary 修改个人密码
// @Description
// @Tags 用户
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param body body  ReqUpdateMyPassword  true "请求参数"
// @Success 200  {object} radarerror.Response
// @Router /api/v3/user/password [put]
func UpdateMyPassword(c *gin.Context) {
	// param
	var req ReqUpdateMyPassword
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Errorf("fail to bind param: %v", err)
		c.Error(&radarerror.InvalidArgs)
		return
	}

	// session
	ss, ok := c.Get(SessME)
	if !ok {
		log.Errorf("need login")
		c.Error(&radarerror.Unauthorized)
		return
	}
	me := ss.(service.ME)

	var cerr *radarerror.CommonError
	svcUser := service.NewUserService(&me)
	cerr = svcUser.UpdatePassword(svcUser.ME.Id, req.Password, false)
	if err != nil {
		c.Error(cerr)
		return
	}

	c.JSON(http.StatusOK, radarerror.Success.Response())
}

// Request: UpdateMyPhone
type ReqUpdateMyPhone struct {
	Phone string `json:"phone" binding:"required,phone"` // 新手机号
}

// @Summary 修改个人手机号
// @Description
// @Tags 用户
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param body body  ReqUpdateMyPhone  true "请求参数"
// @Success 200  {object} radarerror.Response
// @Router /api/v3/user/phone [put]
func UpdateMyPhone(c *gin.Context) {
	// param
	var req ReqUpdateMyPhone
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Errorf("fail to bind param: %v", err)
		c.Error(&radarerror.InvalidArgs)
		return
	}

	// session
	ss, ok := c.Get(SessME)
	if !ok {
		log.Errorf("need login")
		c.Error(&radarerror.Unauthorized)
		return
	}
	me := ss.(service.ME)

	var cerr *radarerror.CommonError
	svcUser := service.NewUserService(&me)

	cerr = svcUser.Update(svcUser.ME.Id, service.SetUser{Phone: &req.Phone})
	if err != nil {
		c.Error(cerr)
		return
	}

	c.JSON(http.StatusOK, radarerror.Success.Response())
}

// Request: GetUserList
type ReqGetUserList struct {
	Page       int64   `form:"page"  binding:"required,gte=1"`      // 分页数，默认1页开始
	PageSize   int64   `form:"page_size"  binding:"required,gte=0"` // 每页数量，传0代表返回全部
	Department *string `form:"department" binding:"omitempty"`      // 部门id
	Role       *string `form:"role" binding:"omitempty" `           // 角色id
	Name       *string `form:"name" binding:"omitempty" `           // 搜索用户名、姓名或警号；模糊匹配
}

// Request: GetUserList
type RspGetUserList struct {
	List         []RspUserData `json:"list"`
	Total        int64         `json:"total"`         // 结果集总数
	TotalDeleted int64         `json:"total_deleted"` // 已删除总数
}

// RspUserData
type RspUserData struct {
	Id           string `json:"id"`            // 主键
	Account      string `json:"account"`       // 登录账号
	Name         string `json:"name"`          // 显示名
	Department   string `json:"department"`    // 部门名
	DepartmentId string `json:"department_id"` // 部门名
	Role         string `json:"role"`          // 角色名
	RoleId       string `json:"role_id"`       // 角色ID
	PoliceNumber string `json:"police_number"` // 警号
	Phone        string `json:"phone"`         // 手机号
	Creator      string `json:"creator"`       // 创建者姓名
	CreateTime   int64  `json:"create_time"`   // 创建时间-时间戳
	Updator      string `json:"updator"`       // 修改者姓名
	UpdateTime   int64  `json:"update_time"`   // 修改时间-时间戳
}

// @Tags 用户
// @Summary 用户列表
// @Description
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param page query int true "第几页，默认从1开始"
// @Param page_size query int true "每页结果数"
// @Param name query string false "筛选条件：用户名、姓名或警号（模糊匹配）"
// @Param department query string false "筛选条件：部门id"
// @Param role query string false "筛选条件：角色id"
// @Success 200  {object} radarerror.ResponseWithData{data=RspGetUserList}
// @Router /api/v3/users [get]
func GetUserList(c *gin.Context) {
	// param
	req := ReqGetUserList{
		PageSize: cfg.DefaultPageSize,
	}
	err := c.ShouldBind(&req)
	if err != nil {
		log.Errorf("fail to bind param: %v", err)
		c.Error(&radarerror.InvalidArgs)
		return
	}
	req.Page = req.Page - 1

	// session
	ss, ok := c.Get(SessME)
	if !ok {
		log.Errorf("need login")
		c.Error(&radarerror.Unauthorized)
		return
	}
	me := ss.(service.ME)

	var cerr *radarerror.CommonError
	svcUser := service.NewUserService(&me)

	filter := service.FilterUser{
		Name: req.Name,
	}
	if req.Department != nil {
		deptId := mongodao.Hex2Id(*req.Department)
		filter.Department = &deptId
	}
	if req.Role != nil {
		roleId := mongodao.Hex2Id(*req.Role)
		filter.Department = &roleId
	}

	// 检查跨部门权限
	if !CheckAuth(&me, []Auth{{Obj: AuthObjTransDepartment, Act: AuthActGet}}) {
		if filter.Department != nil {
			// 检查是否本部门
			if me.Department != *filter.Department {
				log.Debugf("different department, me: %v, user: %v", me.Department.Hex(), (*filter.Department).Hex())
				c.Error(&radarerror.ExceedAuthority)
				return
			}
		} else {
			filter.Department = &me.Department
		}
	}

	// 获取列表
	users, cerr := svcUser.Gets(req.Page, req.PageSize, filter)
	if cerr != nil {
		c.Error(cerr)
		return
	}

	// 结果集总数
	total, cerr := svcUser.GetCount(filter)
	if cerr != nil {
		c.Error(cerr)
		return
	}

	// 已删除用户总数量
	totalDeleted, cerr := svcUser.GetDeletedCount(filter)
	if cerr != nil {
		c.Error(cerr)
		return
	}

	userId2Name := make(map[primitive.ObjectID]string)
	userId2Name[primitive.NilObjectID] = ""

	list := make([]RspUserData, 0, len(users))
	for _, user := range users {
		if _, ok := userId2Name[user.Creator]; !ok {
			user, cerr := svcUser.GetById(user.Creator)
			if cerr != nil {
				c.Error(cerr)
				return
			}
			userId2Name[user.Creator] = user.Name
		}
		if _, ok := userId2Name[user.Updator]; !ok {
			user, cerr := svcUser.GetById(user.Updator)
			if cerr != nil {
				c.Error(cerr)
				return
			}
			userId2Name[user.Updator] = user.Name
		}
		data := RspUserData{
			Id:           user.Id.Hex(),
			Account:      user.Account,
			Name:         user.Name,
			Department:   "",
			DepartmentId: user.Department.Hex(),
			Role:         "",
			RoleId:       user.Role.Hex(),
			PoliceNumber: user.PoliceNumber,
			Phone:        user.Phone,
			Creator:      userId2Name[user.Creator],
			CreateTime:   user.CreateTime.Unix(),
			Updator:      userId2Name[user.Updator],
			UpdateTime:   user.UpdateTime.Unix(),
		}
		list = append(list, data)
	}

	c.JSON(http.StatusOK, radarerror.Success.ResponseWithData(RspGetUserList{
		List:         list,
		Total:        total,
		TotalDeleted: totalDeleted,
	}))
}

// Request: ReqAddUser
type ReqAddUser struct {
	Account      string `json:"account" binding:"required" `       // 账号
	Name         string `json:"name" binding:"required" `          // 用户名
	Department   string `json:"department" binding:"required"`     // 部门id
	Role         string `json:"role" binding:"required"`           // 角色id
	PoliceNumber string `json:"police_number" binding:"omitempty"` // 警号
	Phone        string `json:"phone" binding:"omitempty,phone"`   // 手机号
}

// Response: RspAddUser
type RspAddUser struct {
	Id string `json:"id"` // 用户id
}

// @Tags 用户
// @Summary 新增用户
// @Description
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param body body  ReqAddUser  true "请求参数"
// @Success 200  {object} radarerror.ResponseWithData{data=RspAddUser}
// @Router /api/v3/user [post]
func AddUser(c *gin.Context) {
	// param
	var req ReqAddUser
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Errorf("fail to bind param: %v", err)
		c.Error(&radarerror.InvalidArgs)
		return
	}

	// session
	ss, ok := c.Get(SessME)
	if !ok {
		log.Errorf("need login")
		c.Error(&radarerror.Unauthorized)
		return
	}
	me := ss.(service.ME)

	user := model.User{
		Account:      req.Account,
		Name:         req.Name,
		Department:   mongodao.Hex2Id(req.Department),
		Role:         mongodao.Hex2Id(req.Role),
		PoliceNumber: req.PoliceNumber,
		Phone:        req.Phone,
	}

	// 检查跨部门权限
	if !CheckAuth(&me, []Auth{{Obj: AuthObjTransDepartment, Act: AuthActGet}}) {
		// 检查是否本部门
		if me.Department != user.Department {
			log.Debugf("different department, me: %v, user: %v", me.Department.Hex(), user.Department.Hex())
			c.Error(&radarerror.ExceedAuthority)
			return
		}
	}

	var cerr *radarerror.CommonError
	svcUser := service.NewUserService(&me)

	id, cerr := svcUser.Add(user)
	if cerr != nil {
		c.Error(cerr)
		return
	}

	c.JSON(http.StatusOK,
		radarerror.Success.ResponseWithData(RspAddUser{
			Id: id.Hex(),
		}),
	)
}

// Request: UpdateUser
type ReqUpdateUser struct {
	Set service.SetUser `json:"set" binding:"required"` // 增量修改
}

// @Tags 用户
// @Summary 编辑用户
// @Description
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param body body  ReqUpdateUser true "请求参数"
// @Param id path int true "用户id"
// @Success 200  {object} radarerror.Response
// @Router /api/v3/user/:id [put]
func UpdateUser(c *gin.Context) {
	// param
	var req ReqUpdateUser
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Errorf("fail to bind param: %v", err)
		c.Error(&radarerror.InvalidArgs)
		return
	}
	userId := mongodao.Hex2Id(c.Param("id"))
	if userId == primitive.NilObjectID {
		log.Errorf("invalid id: %v", c.Param("id"))
		c.Error(&radarerror.InvalidArgs)
		return
	}

	// session
	ss, ok := c.Get(SessME)
	if !ok {
		log.Errorf("need login")
		c.Error(&radarerror.Unauthorized)
		return
	}
	me := ss.(service.ME)

	if req.Set.Department != nil {
		// 检查跨部门权限
		if !CheckAuth(&me, []Auth{{Obj: AuthObjTransDepartment, Act: AuthActGet}}) {
			// 检查是否本部门
			if me.Department.Hex() != *req.Set.Department {
				log.Debugf("different department, me: %v, user: %v", me.Department.Hex(), *req.Set.Department)
				c.Error(&radarerror.ExceedAuthority)
				return
			}
		}
	}

	var cerr *radarerror.CommonError
	svcUser := service.NewUserService(&me)

	cerr = svcUser.Update(userId, req.Set)
	if cerr != nil {
		c.Error(cerr)
		return
	}

	c.JSON(http.StatusOK, radarerror.Success.Response())
}

// @Tags 用户
// @Summary 删除用户
// @Description
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path int true "用户id"
// @Success 200  {object} radarerror.Response
// @Router /api/v3/user/:id [delete]
func DeleteUser(c *gin.Context) {
	// param
	userId := mongodao.Hex2Id(c.Param("id"))
	if userId == primitive.NilObjectID {
		log.Errorf("invalid id: %v", c.Param("id"))
		c.Error(&radarerror.InvalidArgs)
		return
	}

	// session
	ss, ok := c.Get(SessME)
	if !ok {
		log.Errorf("need login")
		c.Error(&radarerror.Unauthorized)
		return
	}
	me := ss.(service.ME)

	var cerr *radarerror.CommonError
	svcUser := service.NewUserService(&me)

	user, cerr := svcUser.GetById(userId)
	if cerr != nil {
		c.Error(cerr)
		return
	}

	// 检查跨部门权限
	if !CheckAuth(&me, []Auth{{Obj: AuthObjTransDepartment, Act: AuthActGet}}) {
		// 检查是否本部门
		if me.Department != user.Department {
			log.Debugf("different department, me: %v, user: %v", me.Department.Hex(), user.Department.Hex())
			c.Error(&radarerror.ExceedAuthority)
			return
		}
	}

	cerr = svcUser.Delete(userId)
	if cerr != nil {
		c.Error(cerr)
		return
	}

	c.JSON(http.StatusOK, radarerror.Success.Response())
}

// Request: GetDeletedUserList
type ReqGetDeletedUserList struct {
	Page       int64   `form:"page"  binding:"required,gte=1"`      // 分页数 默认1页开始
	PageSize   int64   `form:"page_size"  binding:"required,gte=0"` // 每页数量，传0代表返回全部
	Department *string `form:"department" binding:"omitempty"`      // 部门id
	Role       *string `form:"role" binding:"omitempty" `           // 角色id
	Name       *string `form:"name" binding:"omitempty" `           // 搜索用户名、姓名或警号；模糊匹配
}

// Request: GetDeletedUserList
type RspGetDeletedUserList struct {
	List  []RspDeletedUserData `json:"list"`
	Total int64                `json:"total"` // 结果集总数
}

// RspDeletedUserData
type RspDeletedUserData struct {
	Id           string `json:"id"`            // 主键（务必设置omitempty，让驱动自动生成）
	Account      string `json:"account"`       // 登录账号
	Name         string `json:"name"`          // 显示名
	Department   string `json:"department"`    // 部门名
	Role         string `json:"role"`          // 角色名
	PoliceNumber string `json:"police_number"` // 警号
	Phone        string `json:"phone"`         // 手机号
	Updator      string `json:"updator"`       // 删除者姓名
	UpdateTime   int64  `json:"update_time"`   // 删除时间-时间戳
}

// @Tags 用户
// @Summary 已删除用户列表
// @Description
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param page query int true "第几页，默认从1开始"
// @Param page_size query int true "每页结果数"
// @Param name query string false "筛选条件：用户名（模糊匹配）"
// @Param department query string false "筛选条件：部门id"
// @Param role query string false "筛选条件：角色id"
// @Success 200  {object} radarerror.ResponseWithData{data=RspGetDeletedUserList}
// @Router /api/v3/users/deleted [get]
func GetDeletedUserList(c *gin.Context) {
	// param
	req := ReqGetDeletedUserList{
		PageSize: cfg.DefaultPageSize,
	}
	err := c.ShouldBind(&req)
	if err != nil {
		log.Errorf("fail to bind param: %v", err)
		c.Error(&radarerror.InvalidArgs)
		return
	}
	req.Page = req.Page - 1

	// session
	ss, ok := c.Get(SessME)
	if !ok {
		log.Errorf("need login")
		c.Error(&radarerror.Unauthorized)
		return
	}
	me := ss.(service.ME)

	var cerr *radarerror.CommonError
	svcUser := service.NewUserService(&me)

	filter := service.FilterUser{
		Name: req.Name,
	}
	if req.Department != nil {
		deptId := mongodao.Hex2Id(*req.Department)
		filter.Department = &deptId
	}
	if req.Role != nil {
		roleId := mongodao.Hex2Id(*req.Role)
		filter.Department = &roleId
	}

	// 检查跨部门权限
	if !CheckAuth(&me, []Auth{{Obj: AuthObjTransDepartment, Act: AuthActGet}}) {
		if filter.Department != nil {
			// 检查是否本部门
			if me.Department != *filter.Department {
				log.Debugf("different department, me: %v, user: %v", me.Department.Hex(), (*filter.Department).Hex())
				c.Error(&radarerror.ExceedAuthority)
				return
			}
		} else {
			filter.Department = &me.Department
		}
	}

	// 获取列表
	users, cerr := svcUser.GetsDeleted(req.Page, req.PageSize, filter)
	if cerr != nil {
		c.Error(cerr)
		return
	}

	// 结果集总数
	total, cerr := svcUser.GetDeletedCount(filter)
	if cerr != nil {
		c.Error(cerr)
		return
	}

	userId2Name := make(map[primitive.ObjectID]string)
	userId2Name[primitive.NilObjectID] = ""

	list := make([]RspDeletedUserData, 0, len(users))
	for _, user := range users {
		if _, ok := userId2Name[user.Updator]; !ok {
			user, cerr := svcUser.GetById(user.Updator)
			if cerr != nil {
				c.Error(cerr)
				return
			}
			userId2Name[user.Updator] = user.Name
		}
		data := RspDeletedUserData{
			Id:           user.Id.Hex(),
			Account:      user.Account,
			Name:         user.Name,
			Department:   "",
			Role:         "",
			PoliceNumber: user.PoliceNumber,
			Phone:        user.Phone,
			Updator:      userId2Name[user.Updator],
			UpdateTime:   user.UpdateTime.Unix(),
		}
		list = append(list, data)
	}

	c.JSON(http.StatusOK, radarerror.Success.ResponseWithData(RspGetDeletedUserList{
		List:  list,
		Total: total,
	}))
}

// Request: GetUserRender
type ReqGetUserRender struct {
	Page       int64   `form:"page"  binding:"required,gte=1"`      // 分页数，默认1页开始
	PageSize   int64   `form:"page_size"  binding:"required,gte=0"` // 每页数量，传0代表返回全部
	Department *string `form:"department" binding:"omitempty"`      // 部门id
	Role       *string `form:"role" binding:"omitempty" `           // 角色id
	Name       *string `form:"name" binding:"omitempty" `           // 搜索用户名、姓名或警号；模糊匹配
}

// Request: GetUserRender
type RspGetUserRender struct {
	List  []RspUserRenderData `json:"list"`
	Total int64               `json:"total"` // 结果集总数
}

// RspUserRenderData
type RspUserRenderData struct {
	Id   string `json:"id"`   // 主键
	Name string `json:"name"` // 显示名
}

// @Tags 用户
// @Summary 获取用户render列表
// @Description 获取简要信息，用于下拉框等场景
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param page query int true "第几页，默认从1开始"
// @Param page_size query int true "每页结果数"
// @Param name query string false "筛选条件：用户名、姓名或警号（模糊匹配）"
// @Param department query string false "筛选条件：部门id"
// @Param role query string false "筛选条件：角色id"
// @Success 200  {object} radarerror.ResponseWithData{data=RspGetUserRender}
// @Router /api/v3/users/render [get]
func GetUserRender(c *gin.Context) {
	// param
	req := ReqGetUserRender{
		PageSize: cfg.DefaultPageSize,
	}
	err := c.ShouldBind(&req)
	if err != nil {
		log.Errorf("fail to bind param: %v", err)
		c.Error(&radarerror.InvalidArgs)
		return
	}
	req.Page = req.Page - 1

	// session
	ss, ok := c.Get(SessME)
	if !ok {
		log.Errorf("need login")
		c.Error(&radarerror.Unauthorized)
		return
	}
	me := ss.(service.ME)

	var cerr *radarerror.CommonError
	svcUser := service.NewUserService(&me)

	filter := service.FilterUser{
		Name: req.Name,
	}
	if req.Department != nil {
		deptId := mongodao.Hex2Id(*req.Department)
		filter.Department = &deptId
	}
	if req.Role != nil {
		roleId := mongodao.Hex2Id(*req.Role)
		filter.Department = &roleId
	}

	// 检查跨部门权限
	if !CheckAuth(&me, []Auth{{Obj: AuthObjTransDepartment, Act: AuthActGet}}) {
		if filter.Department != nil {
			// 检查是否本部门
			if me.Department != *filter.Department {
				log.Debugf("different department, me: %v, user: %v", me.Department.Hex(), (*filter.Department).Hex())
				c.Error(&radarerror.ExceedAuthority)
				return
			}
		} else {
			filter.Department = &me.Department
		}
	}

	// 获取列表
	users, cerr := svcUser.Gets(req.Page, req.PageSize, filter)
	if cerr != nil {
		c.Error(cerr)
		return
	}

	// 结果集总数
	total, cerr := svcUser.GetCount(filter)
	if cerr != nil {
		c.Error(cerr)
		return
	}

	list := make([]RspUserRenderData, 0, len(users))
	for _, user := range users {
		data := RspUserRenderData{
			Id:   user.Id.Hex(),
			Name: user.Name,
		}
		list = append(list, data)
	}

	c.JSON(http.StatusOK, radarerror.Success.ResponseWithData(RspGetUserRender{
		List:  list,
		Total: total,
	}))
}

// Request: ReqGenAccountByName
type ReqGenAccountByName struct {
	Name string `json:"name" binding:"required,min=1" ` // 姓名
}

// Response: RspGetGenAccountByName
type RspGetGenAccountByName struct {
	Account string `json:"account"` // 账号
}

// @Tags 用户
// @Summary 获取用户名生成账号
// @Description
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param body body  ReqGenAccountByName  true "请求参数"
// @Success 200  {object} radarerror.ResponseWithData{data=RspGetGenAccountByName}
// @Router /api/v3/user/create/account_name [post]
func GetGenAccountByName(c *gin.Context) {
	// param
	var req ReqGenAccountByName
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Errorf("fail to bind param: %v", err)
		c.Error(&radarerror.InvalidArgs)
		return
	}

	var cerr *radarerror.CommonError
	svcUser := service.NewUserService(nil)
	// 获取
	name, cerr := svcUser.GetCreateAccount(req.Name)
	if cerr != nil {
		c.Error(cerr)
		return
	}

	c.JSON(http.StatusOK,
		radarerror.Success.ResponseWithData(RspGetGenAccountByName{
			Account: name,
		}),
	)
}
