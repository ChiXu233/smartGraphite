package router

import (
	"SmartGraphite-server/controller"
	"github.com/gin-gonic/gin"
)

func DeviceTypeRouter(engine *gin.Engine) {
	deviceType := engine.Group("/deviceType")
	{

		deviceType.GET("/create", controller.CreateDeviceType)
		deviceType.GET("/getDeviceType", controller.GetDeviceType)

	}
}
