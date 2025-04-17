package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Alarm struct {
	Id      primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Device  Device             `json:"device" bson:"device"`   //设备
	Type    string             `json:"type" bson:"type"`       //离线，环保，生产
	Content string             `json:"content" bson:"content"` //内容
	Time    string             `json:"time" bson:"time"`
}
