package global

import mqtt "github.com/eclipse/paho.mqtt.golang"

var (
	// MqttSubAlarm mqtt.Client //接收报警客户端
	// MqttPubAlarm mqtt.Client //发送报警客户端
	RTDTUMqtt mqtt.Client
)
