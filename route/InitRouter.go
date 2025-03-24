package route

import (
	_ "bessGin/app/api/controller" //一定要导入这个Controller包，用来注册需要访问的方法
	_ "bessGin/app/api/controller/v1"
	"bessGin/route/middleware/jwt"
	"github.com/gin-gonic/gin"
	ginAutoRouter "github.com/xiaomo1989/gin-auto-router"
)

func InitRouter() *gin.Engine {
	//初始化路由
	r := gin.Default()
	//绑定基本路由，访问路径：/User/List
	ginAutoRouter.Bind(r)
	v1Route := r.Group("/v1")
	//加载并使用登录验证中间件
	v1Route.Use(jwt.JWT())
	{
		//绑定Group路由，访问路径：/v1/article/list
		ginAutoRouter.BindGroup(v1Route)
	}
	return r
}
