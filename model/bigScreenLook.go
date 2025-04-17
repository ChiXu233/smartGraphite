package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type BigScreenLook struct {
	Id         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Name       string             `bson:"name" json:"name"`
	Code       string             `bson:"code" json:"code"`
	Data       []DataDetail       `bson:"data" json:"data"`
	Url        string             `bson:"url" json:"url"`                                   //路由
	Method     string             `bson:"method" json:"method"`                             //方法
	Desc       string             `bson:"desc" json:"desc"`                                 //描述
	UpdateTime string             `bson:"updateTime,omitempty" json:"updateTime,omitempty"` //此条记录创建时间
}

type InfoDataDetail struct {
	Data []string `json:"data" bson:"data"`
	Time []string `json:"time" bson:"time"`
	Unit string   `json:"unit" bson:"unit"`
}

type DataDetail struct { //声明model.BigScreenLook中Data的类型
	Name string                    `json:"name" bson:"name"`
	Code string                    `json:"code" bson:"code"`
	Info map[string]InfoDataDetail `json:"info" bson:"info"`
}
