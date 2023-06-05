package httpserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	handler "github.com/SeeJson/account/cmd/account/handler/http"
	radarerror "github.com/SeeJson/account/error"
	"github.com/SeeJson/account/service"
	"github.com/SeeJson/account/util/jwt"
	mstring "github.com/SeeJson/account/util/string"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	HeaderRequestId = "X-Request-ID"
	LogRequestId    = "request_id"
)

// errorHandler 对错误结果统一处理
func errorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		errs := c.Errors.ByType(gin.ErrorTypeAny)
		if len(errs) > 0 {
			err := errs.Last().Err
			switch cErr := err.(type) {
			case *radarerror.CommonError:
				c.JSON(cErr.HttpStatus(), cErr.Response())
				return
			default:
				log.Errorf("unknown err type: %v", err)
			}
		}
	}
}

/*
 * 解析jwt token，获得当前操作人信息
 */
func decodeJwtToken(c *gin.Context) {
	tokenFields := strings.Fields(c.GetHeader("Authorization"))
	if len(tokenFields) != 2 || tokenFields[0] != "Bearer" {
		log.Errorf("invalid authorization header: %v", c.GetHeader("Authorization"))
		c.Error(&radarerror.Unauthorized)
		c.Abort()
		return
	}
	b64Token := tokenFields[1]

	jwtClaim, err := jwt.DecodeB64Token(b64Token)

	if err != nil {
		c.Error(&radarerror.Unauthorized)
		c.Abort()
		return
	}

	// check token timeout
	if jwtClaim.Exp < time.Now().Unix() {
		log.Errorf("token expired: %v", b64Token)
		c.Error(&radarerror.Unauthorized)
		c.Abort()
		return
	}

	me, err := service.LoadME(jwtClaim.Payload)
	if err != nil {
		log.Errorf("fail to decode session: %v", err)
		c.Error(&radarerror.Unauthorized)
		c.Abort()
		return
	}

	log.Debugf("me: %+v", me)

	// check user version
	if !service.IsSessionVersionValid(me.Id, me.Version) {
		log.Errorf("token version invalid: %v", me.Version)
		c.Error(&radarerror.Unauthorized)
		c.Abort()
		return
	}

	//todo 区分token类型，client的怎么处理？？？，检查token过期，检查权限
	// 校验version，当登录、修改密码、注销、重置密码都会递增版本号

	// todo client_id
	// todo scope

	// todo role

	c.Set(handler.SessME, *me)

	c.Next()
}

// 格式：map[uri][method][]handler.Auth
var apiAuthMap map[string]map[string][]handler.Auth = map[string]map[string][]handler.Auth{
	"/api/v3/user/:id/password": {
		"PUT": []handler.Auth{
			{Obj: handler.AuthObjUser, Act: handler.AuthActUpdate},
		},
	},
	"/api/v3/roles": {
		"GET": []handler.Auth{
			{Obj: handler.AuthObjRole, Act: handler.AuthActGet},
		},
	},
	"/api/v3/role": {
		"POST": []handler.Auth{
			{Obj: handler.AuthObjRole, Act: handler.AuthActAdd},
		},
	},
	"/api/v3/role/:id": {
		"PUT": []handler.Auth{
			{Obj: handler.AuthObjRole, Act: handler.AuthActUpdate},
		},
		"DELETE": []handler.Auth{
			{Obj: handler.AuthObjRole, Act: handler.AuthActDelete},
		},
	},
	"/api/v3/departments": {
		"GET": []handler.Auth{
			{Obj: handler.AuthObjDepartment, Act: handler.AuthActGet},
		},
	},
	"/api/v3/department": {
		"POST": []handler.Auth{
			{Obj: handler.AuthObjDepartment, Act: handler.AuthActAdd},
		},
	},
	"/api/v3/department/:id": {
		"PUT": []handler.Auth{
			{Obj: handler.AuthObjDepartment, Act: handler.AuthActUpdate},
		},
		"DELETE": []handler.Auth{
			{Obj: handler.AuthObjDepartment, Act: handler.AuthActDelete},
		},
	},
	"/api/v3/users": {
		"GET": []handler.Auth{
			{Obj: handler.AuthObjUser, Act: handler.AuthActGet},
		},
	},
	"/api/v3/users/deleted": {
		"GET": []handler.Auth{
			{Obj: handler.AuthObjUser, Act: handler.AuthActGet},
		},
	},
	"/api/v3/user": {
		"POST": []handler.Auth{
			{Obj: handler.AuthObjUser, Act: handler.AuthActAdd},
		},
	},
	"/api/v3/user/:id": {
		"PUT": []handler.Auth{
			{Obj: handler.AuthObjUser, Act: handler.AuthActUpdate},
		},
		"DELETE": []handler.Auth{
			{Obj: handler.AuthObjUser, Act: handler.AuthActDelete},
		},
	},
}

/*
 * 对个别请求做权限校验
 */
func checkPermission(c *gin.Context) {
	uri := c.Request.URL.Path
	method := c.Request.Method

	// session
	ss, ok := c.Get(handler.SessME)
	if !ok {
		log.Errorf("need login")
		c.Error(&radarerror.Unauthorized)
		c.Abort()
		return
	}
	me := ss.(service.ME)

	if _, ok := apiAuthMap[uri]; !ok {
		c.Next()
		return
	}

	if _, ok := apiAuthMap[uri][method]; !ok {
		c.Next()
		return
	}

	if !handler.CheckAuth(&me, apiAuthMap[uri][method]) {
		log.Errorf("exceed authority: %v %v %v", me.Id.Hex(), uri, method)
		c.Error(&radarerror.ExceedAuthority)
		c.Abort()
		return
	}

	// check passwd_reset
	if !me.PasswordReset &&
		(uri != "/api/v3/user/password" || method != "PUT") {
		log.Errorf("need to reset password: %v %v %v", me.Id.Hex(), uri, method)
		c.Error(&radarerror.NeedResetPwd)
		c.Abort()
		return
	}

	c.Next()
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := c.Request.Header.Get(HeaderRequestId)
		if requestId == "" {
			requestId = mstring.GetUUID()
			c.Header(HeaderRequestId, requestId)
		}
		c.Set(HeaderRequestId, requestId)
		c.Next()
	}
}

var (
	WriteSizeLimit = 1024 * 10
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
	size int
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	if w.size+len(b) < WriteSizeLimit {
		n, err := w.body.Write(b)
		if err != nil {
			log.Warnf("write log body failed: %v", err)
		} else {
			w.size += n
		}
	}
	return w.ResponseWriter.Write(b)
}

func Logger() gin.HandlerFunc {

	return func(c *gin.Context) {
		var readCloser io.ReadCloser

		buf, _ := c.GetRawData()
		if c.ContentType() == gin.MIMEJSON {
			readCloser = ioutil.NopCloser(bytes.NewBuffer(buf))
		}

		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(buf))

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		start := time.Now().UnixNano() / 1e6
		c.Next()
		end := time.Now().UnixNano() / 1e6
		latency := end - start

		request := c.Request.URL.RawQuery
		if c.ContentType() == gin.MIMEJSON {
			request = ReadBody(readCloser)
		}

		if len(request) >= WriteSizeLimit {
			request = "..."
		}

		response := "..."
		contentType := c.Writer.Header().Get("Content-Type")
		if strings.Contains(contentType, gin.MIMEJSON) {
			if blw.size <= WriteSizeLimit {
				response = blw.body.String()
			}
		}

		log.WithFields(log.Fields{
			LogRequestId: c.GetString(HeaderRequestId),
			"request":    request,
			"response":   response,
			"ip":         c.ClientIP(),
			"path":       c.Request.URL.Path,
			"method":     c.Request.Method,
			"status":     c.Writer.Status(),
			"latency":    fmt.Sprintf("%v ms", latency),
		}).Info("")
	}
}

func ReadBody(reader io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)

	newBuf := new(bytes.Buffer)
	if err := json.Compact(newBuf, buf.Bytes()); err != nil {
		return buf.String()
	}

	return newBuf.String()
}
