### 应用
- 独立的进程或工具
- 例如：
	- web进程
	- 初始化部署工具
	- 后台POLLER

### 生成swagger文档
- 安装需要用到的包
	- go get -u github.com/swaggo/swag/cmd/swag  //github.com/swaggo/swag/cmd/swag@v1.6.7
	- go get -u github.com/swaggo/gin-swagger
	- go get -u github.com/swaggo/gin-swagger/swaggerFiles
- 接口代码支持swagger (查看文档的http)
	- 在路由中引用
		```
		import (
			ginSwagger "github.com/swaggo/gin-swagger"
			"github.com/swaggo/gin-swagger/swaggerFiles"
			_ "github.com/SeeJson/account/cmd/account/docs"  // {account/docs}替换成自己项目的
		)
		//	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		```
	- 在接口里通过注释构建api文档
		```
		// @Tags 登录				(标题)
		// @Summary Login     		(副标题)
		// @Description 			(描述)
		// @Description | 错误代码 | 价格 | 数量 |
		// @Description | :-------- | --------:| :--: |
		// @Description | iPhone | 6000 元 | 5 |
		// @Description | iPad | 3800 元 | 12 |
		// @Description | iMac | 10000 元 | 234 |
		// @Accept application/json
		// @Produce application/json
		// @Param Authorization header string false "Bearer 用户令牌"
		// @Param body body  LoginRequest  true "查询参数"
		// @Success 200  {object} radarerror.CommonData
		// @Failure 400  {object} radarerror.CommonError --> 失败后返回数据结构
		// @Failure 500  {object} radarerror.CommonError --> 失败后返回数据结构
		// @Router /api/v3/auth/login [get]
		```

		- success 成功响应 格式: [ 状态码 {数据类型} 数据类型 备注 ]
			```
			@Success 200 {object} Response "返回空对象"
			```
		- failure 失败响应 格式: [ 状态码 {数据类型} 数据类型 备注 ]
			```
			@Failure 400 {object} ResponseError
			```
		- header 请求头字段 格式: [ 状态码 {数据类型} 数据类型 备注 ]
			```
			@Header 200 {string} Token "qwerty"
			```
		- 多字段定义时不能跨字段
		
- 命令 (在进程中使用命令如下)
	- `swag init --parseDependency true` (因为项目cmd多个进程,进程中引用外部依赖的go文件,属于需要设置parseDependency)