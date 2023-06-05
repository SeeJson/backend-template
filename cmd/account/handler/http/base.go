package httphandler

import "github.com/SeeJson/account/service"

type Config struct {
	DefaultPageSize int64 `mapstructure:"default_page_size"`
}

var cfg Config

func SetConfig(c Config) {
	cfg = c
}

const (
	SessME = "me"
)

const (
	// 权限对象的bit-mark
	AuthObjTransDepartment  = 1  // 跨部门
	AuthObjLogoLibPublic    = 4  // 图标库（公共）
	AuthObjFaceLibPublic    = 6  // 人脸库（公共）
	AuthObjImageLibPublic   = 8  // 人脸库（公共）
	AuthObjKeywordLibPublic = 10 // 关键词库（公共）
	AuthObjDepartment       = 17 // 部门管理
	AuthObjRole             = 18 // 角色管理
	AuthObjUser             = 19 // 用户管理

	// 权限动作的bit-mark
	AuthActGet      = 1  // 2^0
	AuthActAdd      = 2  // 2^1
	AuthActUpdate   = 4  // 2^2
	AuthActDelete   = 8  // 2^3
	AuthActDownload = 16 // 2^4
	AuthActUpload   = 32 // 2^5
	AuthActFeedBack = 64 // 2^6 提交错误反馈
)

type Auth struct {
	Obj int64 // 权限对象的二进制掩码 model.auth_obj.bit_mark
	Act int64 // 权限动作的二进制掩码 model.auth_act.bit_mark
}

/*
 * 检查是否拥有指定的权限（可以要求同时拥有多个）
 */
func CheckAuth(me *service.ME, auths []Auth) bool {
	for _, auth := range auths {
		acts, ok := me.AuthMp[auth.Obj]
		if !ok || (acts&auth.Act) == 0 {
			return false
		}
	}
	return true
}
