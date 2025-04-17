package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type DeviceType struct {
	Id         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Name       string             `bson:"name" json:"name"`
	Sensors    []Sensors          `bson:"sensors" json:"sensors"`                           //传感器类型
	Desc       string             `bson:"desc" json:"desc"`                                 //描述 //
	Protocol   Protocol           `bson:"protocol" json:"protocol"`                         //协议
	CreateTime string             `bson:"createTime,omitempty" json:"createTime,omitempty"` //创建时间
	UpdateTime string             `bson:"UpdateTime,omitempty" json:"UpdateTime,omitempty"` //修改时间
}
type Sensors struct {
	Code           string           `bson:"code" json:"code"` //传感器对应的标识，dtu是地址位
	Name           string           `bson:"name" json:"name"`
	DetectionValue []DetectionValue `bson:"detectionValue" json:"detectionValue"`
}
type DetectionValue struct {
	Key       string    `bson:"key" json:"key"` //检测到的值
	Unit      string    `bson:"unit" json:"unit"`
	AlarmRule AlarmRule `bson:"alarmRule" json:"alarmRule"` //报警阀值
	ValidRule ValidRule `bson:"validRule" json:"validRule"` //有效阈值
}
type AlarmRule struct {
	Min string `bson:"min" json:"min"`
	Max string `bson:"max" json:"max"`
}
type ValidRule struct {
	Min string `bson:"min" json:"min"`
	Max string `bson:"max" json:"max"`
}
type Protocol struct {
	Name string `bson:"name" json:"name"` //名称
	Key  []Key  `bson:"key" json:"key"`   //协议Key对照
}
type Key struct {
	OriginalKey string `bson:"originalKey" json:"originalKey"` //原始key
	AnalyticKey string `bson:"analyticKey" json:"analyticKey"` //解析后的key
	Unit        string `bson:"unit" json:"unit"`               //单位
}
