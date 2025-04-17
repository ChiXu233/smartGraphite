package controller

import (
	"SmartGraphite-server/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CreateDeviceType(c *gin.Context) {
	//c.JSON(http.StatusOK, service.CreateDeviceType(deviceType))
}
func GetDeviceType(c *gin.Context) {
	c.JSON(http.StatusOK, service.GetDeviceType())
}
