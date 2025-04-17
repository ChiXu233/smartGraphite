package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type DeviceData struct {
	Id         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Dataset    []Dataset          `bson:"dataset" json:"dataset"`                           //name;value;unit
	DataTime   string             `bson:"dataTime" json:"dataTime"`                         //数据测量开始时间或采集时刻
	DeviceCode string             `bson:"deviceCode" json:"deviceCode"`                     //设备唯一编码
	Signal     string             `bson:"signal" json:"signal"`                             //命令编码CN
	Payload    string             `bson:"payload" json:"payload"`                           //原始数据包
	CreateTime string             `bson:"createTime,omitempty" json:"createTime,omitempty"` //此条记录创建时间
	UpdateTime string             `bson:"updateTime,omitempty" json:"updateTime,omitempty"` //此条记录创建时间
}
type Dataset struct {
	Name  string `bson:"name" json:"name"`
	Value string `bson:"value" json:"value"`
	Unit  string `bson:"unit" json:"unit"`
}
type DeviceTimeTen struct {
	Id         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	DeviceCode string             `bson:"deviceCode" json:"deviceCode"`                     //设备唯一编码
	Dataset    []Dataset          `bson:"dataset" json:"dataset"`                           //name;value;unit
	CreateTime string             `bson:"createTime,omitempty" json:"createTime,omitempty"` //此条记录创建时间
}

// 总和表
type DayAndHour struct {
	Id         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	DeviceCode string             `bson:"deviceCode" json:"deviceCode"` //设备唯一编码
	TheyName   string             `bson:"theyName"json:"theyName"`      //设备名称
	DataOne    []DataOne          `bson:"dataOne" json:"dataOne"`
	OnlineRate string             `bson:"onlineRate"json:"onlineRate"`                      //在线率
	CreateTime string             `bson:"createTime,omitempty" json:"createTime,omitempty"` //此条记录创建时间
}
type DataOne struct {
	Average   string      `bson:"average"json:"average"`  //平均值
	Power     string      `bson:"power" json:"power"`     //功率
	CarTime   string      `bson:"carTime" json:"carTime"` //针对东西天车的总运行时间进行存储的，还有plc运行时间也存这里
	Name      string      `bson:"name" json:"name"`
	ValueData []ValueData `bson:"valueData" json:"valueData"`
}
type ValueData struct {
	Value string `bson:"value" json:"value"`
	Times string `bson:"times" json:"times"`
}
type DayAndHour2 struct {
	OnlineRate string `bson:"onlineRate"json:"onlineRate"` //在线率
	Name       string `bson:"name" json:"name"`
}
type Day3 struct {
	CreateTime string        `bson:"createTime" json:"createTime"`
	Value      []DayAndHour2 `bson:"value" json:"value"`
}
