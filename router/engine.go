package router

import (
	"SmartGraphite-server/utils"
	"github.com/gin-gonic/gin"
)

func GetEngine() *gin.Engine {
	engine := gin.Default()
	engine.Use(utils.CORS(utils.Options{Origin: "*"})) //跨域
	//engine.Use(middleware.CustomRouterMiddle1)
	//路由分组
	//设备
	//DeviceRouter(engine)
	////设备类型
	//DeviceTypeRouter(engine)
	return engine
}
