package error

import (
	"encoding/json"
	"net/http"
)

// CommonError
type CommonError struct {
	Code    int    `json:"code"` //code 错误码
	Message string `json:"msg"`  //msg 消息
}

// Response
type Response struct {
	Code    int    `json:"code"` //code 错误码
	Message string `json:"msg"`  //msg 消息
}

// ResponseWithData
type ResponseWithData struct {
	Code    int         `json:"code"` //code 错误码
	Message string      `json:"msg"`  //msg 消息
	Data    interface{} `json:"data"` //数据信息
}

func (err *CommonError) Error() string {
	b, _ := json.Marshal(err)
	return string(b)
}

func (err *CommonError) HttpStatus() int {
	switch err.Code {
	case Success.Code,
		NeedResetPwd.Code:
		return http.StatusOK
	case InternalServerError.Code:
		return http.StatusInternalServerError
	case Unauthorized.Code:
		return http.StatusUnauthorized
	case ForbiddenAccess.Code:
		return http.StatusForbidden
	case InvalidArgs.Code:
		return http.StatusBadRequest
	default:
		return http.StatusBadRequest
	}
}

func (err *CommonError) Response() interface{} {
	return Response{
		Code:    err.Code,
		Message: err.Message,
	}
}

func (err *CommonError) ResponseWithData(data interface{}) interface{} {
	return ResponseWithData{
		Code:    err.Code,
		Message: err.Message,
		Data:    data,
	}
}

// Public
var (
	Success CommonError = CommonError{0, "success"}
)

// Account Error  20001 ~ 29999
var (
	InternalServerError      CommonError = CommonError{20001, "internal server error"}
	Unauthorized             CommonError = CommonError{20002, "need login"}
	ForbiddenAccess          CommonError = CommonError{20003, "forbidden access"}
	InvalidArgs              CommonError = CommonError{20004, "invalid args"}
	InvalidCaptcha           CommonError = CommonError{20005, "invalid captcha"}
	AccountNotFound          CommonError = CommonError{20006, "account not found"}
	InvalidPassword          CommonError = CommonError{20007, "invalid password"}
	DuplicatedAccount        CommonError = CommonError{20008, "duplicated account"}
	UserNotFound             CommonError = CommonError{20009, "user not found"}
	DuplicatedRoleName       CommonError = CommonError{20010, "duplicated role name"}
	RoleNotFound             CommonError = CommonError{20011, "role  not found"}
	InvalidAuths             CommonError = CommonError{20012, "invalid auths"}
	NeedResetPwd             CommonError = CommonError{20013, "need to reset password"}
	DepartmentNotFound       CommonError = CommonError{20014, "department not found"}
	DuplicatedDepartmentName CommonError = CommonError{20015, "duplicated department name"}
	InvalidRegions           CommonError = CommonError{20016, "invalid regions"}
	DuplicatedPoliceNumber   CommonError = CommonError{20017, "duplicated police number"}
	ExceedAuthority          CommonError = CommonError{20018, "exceed your authority"} // 越权行为
)
