package httpserver

import (
	_ "github.com/SeeJson/account/cmd/account/docs"
	handler "github.com/SeeJson/account/cmd/account/handler/http"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func getRouter() *gin.Engine {
	binding.Validator = new(defaultValidator)
	router := gin.Default()
	router.Use(gin.Recovery())
	router.Use(errorHandler())
	router.Use(RequestID())
	router.Use(Logger())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authGroup := router.Group("/api/v3", decodeJwtToken, checkPermission)

	// 登录相关
	router.POST("/api/v3/auth/login", handler.Login)
	router.GET("/api/v3/auth/captcha", handler.GenCaptcha)

	// 用户
	authGroup.POST("/user", handler.AddUser)
	authGroup.GET("/users", handler.GetUserList)
	authGroup.GET("/users/deleted", handler.GetDeletedUserList)
	authGroup.PUT("/user/:id", handler.UpdateUser)
	authGroup.DELETE("/user/:id", handler.DeleteUser)
	authGroup.PUT("/user/:id/password", handler.ResetPassword) // 超级管理员给用户重置密码
	authGroup.PUT("/user/password", handler.UpdateMyPassword)  // 用户自己修改密码
	authGroup.PUT("/user/phone", handler.UpdateMyPassword)     // 用户自己修改手机号
	authGroup.GET("/users/render", handler.GetUserRender)      // 获取用户render列表（返回的只有简要信息：id+name） 这种通常不限制权限

	return router
}
