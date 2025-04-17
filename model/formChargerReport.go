package model

import "go.mongodb.org/mongo-driver/bson/primitive"

//压型配料报表
type FormChargerReport struct {
	Id         primitive.ObjectID        `json:"_id,omitempty" bson:"_id,omitempty"`
	Data       []FormChargerReportDetail `json:"data" bson:"data"`             //详细数据
	Crucible   string                    `json:"crucible" bson:"crucible"`     //锅
	CreateTime string                    `json:"createTime" bson:"createTime"` //创建时间
}

//详细数据
type FormChargerReportDetail struct {
	OriginKey string `json:"originKey" bson:"originKey"` //原名称
	Key       string `json:"key" bson:"key"`             //名称
	Value     string `json:"value" bson:"value"`         //值
	Unit      string `json:"unit" bson:"unit"`           //单位
}
