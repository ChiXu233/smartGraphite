package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Device struct {
	Id           primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Name         string             `bson:"name" json:"name"`
	DeviceTypeId primitive.ObjectID `bson:"deviceTypeId" json:"deviceTypeId"`                 //设备类型id 对应设备类型表
	Code         string             `bson:"code" json:"code"`                                 //设备唯一标识
	IsValid      bool               `bson:"isValid" json:"isValid"`                           //是否被删除
	Status       string             `bson:"status" json:"status"`                             //状态 正常 异常
	IsCustom     bool               `bson:"isCustom" json:"isCustom"`                         //报警规则是否自定义，
	Sensors      []Sensors          `bson:"sensors" json:"sensors"`                           //具体数据信息规则继承设备类型
	Desc         string             `bson:"desc" json:"desc"`                                 //描述
	Loc          Loc                `bson:"loc" json:"loc"`                                   //设备地址
	Img          string             `bson:"img" json:"img"`                                   //图床地址
	CreateTime   string             `bson:"createTime,omitempty" json:"createTime,omitempty"` //创建时间
	UpdateTime   string             `bson:"UpdateTime,omitempty" json:"UpdateTime,omitempty"` //修改时间
}
type Loc struct {
}
