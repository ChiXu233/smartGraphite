package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/utils"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func ReceiveMqtt() {
	go MqttClientDTU(global.RTDTUMqtt, "rt-base-West") //焙烧
}

func MqttClientDTU(mqttClient mqtt.Client, topic string) {
	mqttClient.Subscribe(topic, 0, func(client mqtt.Client, message mqtt.Message) {
		fmt.Println(message.Topic(), message.Payload())
		utils.Try(func() {
			//解决设备数据发送错误问题
			ParseDTUDataNew(message.Payload(), len(message.Payload()))
		})
	})
}
