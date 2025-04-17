package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type DTU struct {
	Id         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`               //objectId
	DTUId      string             `bson:"DTUId" json:"DTUId"`                               //dtu上的唯一标识 设备表中的code
	DTUData    []DTUData          `bson:"DTUData" json:"DTUData"`                           //数据数组
	Payload    string             `bson:"payload" json:"payload"`                           //原始数据
	CreateTime string             `bson:"createTime,omitempty" json:"createTime,omitempty"` //创建时间
	UpdateTime string             `bson:"updateTime,omitempty" json:"updateTime,omitempty"` //更新时间
}
type DTUData struct {
	SensorId      string          `bson:"sensorId" json:"sensorId"`           //相当于传感器在dtu上的地址位
	SensorName    string          `bson:"sensorName" json:"sensorName"`       //传感器名称
	DTUDataDetail []DTUDataDetail `bson:"DTUDataDetail" json:"DTUDataDetail"` //详细数据信息
}
type DTUDataDetail struct {
	Key   string `bson:"key" json:"key"`     //传感器能检测到的key值
	Value string `bson:"value" json:"value"` //value值
	Unit  string `bson:"unit" json:"unit"`   //单位
}

type SensorData struct {
	Code          string          `bson:"sensorId" json:"sensorId"`                         //相当于传感器在dtu上的地址位
	Name          string          `bson:"sensorName" json:"sensorName"`                     //传感器名称
	DTUDataDetail []DTUDataDetail `bson:"DTUDataDetail" json:"DTUDataDetail"`               //详细数据信息
	CreateTime    string          `bson:"createTime,omitempty" json:"createTime,omitempty"` //创建时间
}
