package main

import (
	"SmartGraphite-server/controller"
	"SmartGraphite-server/initialize"
	"SmartGraphite-server/router"
	"SmartGraphite-server/service"
)

func init() {
	initialize.Init()
}
func main() {
	//go service.SocketServerDTU()

	go service.SocketServerNew() //动态协议
	go service.SocketServer3()
	go service.SocketServer2()
	go service.CornTimer()
	go controller.TokenTimer()
	go service.TokenTimer2()
	service.ReceiveMqtt() //mqtt接收焙烧数据
	engine := router.GetEngine()
	if err := engine.Run(":8090"); err != nil {
		panic(err)
	}
}
