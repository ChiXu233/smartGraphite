package initialize

import (
	"SmartGraphite-server/global"
	"runtime"
)

func Init() {
	MongoInit()
	UnitInit()
	if runtime.GOOS != "linux" {
		return
	}
	MqttInitNew("RTBSMqtt11", &global.RTDTUMqtt)
	//MqttInit("ManagerMqttSubAlarm", &global.MqttSubAlarm)
	//MqttInit("ManagerMqttPubAlarm", &global.MqttPubAlarm)
}
