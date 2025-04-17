package router

import (
	"SmartGraphite-server/controller"
	"github.com/gin-gonic/gin"
)

func DeviceRouter(engine *gin.Engine) {
	device := engine.Group("device")
	{
		device.GET("/create", controller.CreateDevice)
	}
}
