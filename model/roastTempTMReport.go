package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type RoastTempTMReport struct {
	Id         primitive.ObjectID    `json:"_id,omitempty" bson:"_id,omitempty"`               //ID
	Code       string                `json:"code" bson:"code"`                                 //自定义编号
	Name       string                `json:"name" bson:"name"`                                 //名称
	MaxMap     map[string]SensorData `json:"MaxMap" bson:"MaxMap"`                             //定义map
	StartTime  string                `json:"startTime,omitempty" bson:"startTime,omitempty"`   //下料开始时间
	EndTime    string                `json:"endTime,omitempty" bson:"endTime,omitempty"`       //下料结束时间
	CreateTime string                `json:"createTime,omitempty" bson:"createTime,omitempty"` //报表创建时间
	//Data       []DataDetail          `json:"data" bson:"data"`                                 //折线图图数据

}
type RoastTempTMReportNew struct {
	Id         primitive.ObjectID       `json:"_id,omitempty" bson:"_id,omitempty"`               //ID
	Code       string                   `json:"code" bson:"code"`                                 //自定义编号
	Name       string                   `json:"name" bson:"name"`                                 //名称
	MaxMap     map[string]DTUDataDetail `json:"MaxMap" bson:"MaxMap"`                             //定义map
	StartTime  string                   `json:"startTime,omitempty" bson:"startTime,omitempty"`   //下料开始时间
	EndTime    string                   `json:"endTime,omitempty" bson:"endTime,omitempty"`       //下料结束时间
	CreateTime string                   `json:"createTime,omitempty" bson:"createTime,omitempty"` //报表创建时间
	//Data       []DataDetail          `json:"data" bson:"data"`                                 //折线图图数据

}

// 挤压报表
type ExtrusionReport struct {
	Id         primitive.ObjectID       `json:"_id,omitempty" bson:"_id,omitempty"`               //ID
	Code       string                   `json:"code" bson:"code"`                                 //自定义编号
	Name       string                   `json:"name" bson:"name"`                                 //名称
	MaxMap     map[string]DTUDataDetail `json:"MaxMap" bson:"MaxMap"`                             //定义最大值map
	MinMap     map[string]DTUDataDetail `json:"MinMap" bson:"MinMap"`                             //定义最小值map
	AvgMap     map[string]DTUDataDetail `json:"AvgMap" bson:"AvgMap"`                             //定义平均值map
	StartTime  string                   `json:"startTime,omitempty" bson:"startTime,omitempty"`   //开始时间
	EndTime    string                   `json:"endTime,omitempty" bson:"endTime,omitempty"`       //结束时间
	Speed      string                   `json:"speed" bson:"speed"`                               //垂头速度
	Displace   string                   `json:"displace" bson:"displace"`                         //垂头位移
	CreateTime string                   `json:"createTime,omitempty" bson:"createTime,omitempty"` //报表创建时间
	Status     string                   `json:"status" bson:"status"`                             //挤压还是预压
}
