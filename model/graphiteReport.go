package model

import "go.mongodb.org/mongo-driver/bson/primitive"

//石墨化报表数据 每一炉存一条记录
type GraReportForm struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	HeadTitle   string             `bson:"headTitle" json:"headTitle"`                       //标题
	StoveNumber string             `bson:"stoveNumber" json:"stoveNumber"`                   //炉号
	Data        [][]ReportDetail   `bson:"data" json:"data"`                                 //数据
	RunTime     string             `bson:"runTime" json:"runTime"`                           //送电时长
	StartTime   string             `bson:"startTime,omitempty" json:"startTime,omitempty"`   //开始生产时间(送电时刻)
	EndTime     string             `bson:"endTime,omitempty" json:"endTime,omitempty"`       //结束生产时间
	CreateTime  string             `bson:"createTime,omitempty" json:"createTime,omitempty"` //报表创建时间
}
type ReportDetail struct {
	Key  string         `bson:"key" json:"key"`   //变量名
	Unit string         `bson:"unit" json:"unit"` //单位
	VT   []ValueAndTime `bson:"vt" json:"vt"`     //每个变量的多条创建时间和值
}
type ValueAndTime struct {
	CreateTime string `bson:"createTime" json:"createTime"` //创建时间
	Value      string `bson:"value" json:"value"`           //变量检测值
}

//有功功率时刻数据
type GraPower struct {
	Id             primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	StartTime      string             `bson:"startTime,omitempty" json:"startTime,omitempty"`     //送电时刻
	StoveNumber    string             `bson:"stoveNumber" json:"stoveNumber"`                     //炉号
	CreateTime     string             `bson:"createTime,omitempty" json:"createTime,omitempty"`
}

//电量时刻数据
type GraElectric struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	StoveNumber string             `bson:"stoveNumber" json:"stoveNumber"` //炉号
	CreateTime  string             `bson:"createTime,omitempty" json:"createTime,omitempty"`
}
