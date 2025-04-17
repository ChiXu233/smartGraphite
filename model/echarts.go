package model

import "go.mongodb.org/mongo-driver/bson/primitive"

//echarts
type Echarts struct {
	Id         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`               //Id
	Name       string             `json:"name" bson:"name"`                                 //名称
	Code       string             `json:"code" bson:"code"`                                 //编号,可自定义
	Data       [][]string         `json:"data" bson:"data"`                                 //数据
	Unit       map[string]string  `json:"unit" bson:"unit"`                                 //数据单位
	Describe   string             `json:"describe" bson:"describe"`                         //详细描述
	CreateTime string             `json:"createTime,omitempty" bson:"createTime,omitempty"` //创建时间
	UpdateTime string             `json:"updateTime,omitempty" bson:"updateTime,omitempty"` //更新时间
}
