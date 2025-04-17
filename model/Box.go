package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Box struct {
	Id           primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	DeviceTypeId primitive.ObjectID `bson:"deviceTypeId" json:"deviceTypeId"` //设备类型id 对应设备类型表
	BoxId        string             `bson:"boxId" json:"boxId"`               //Box上唯一表示
	Data         []BoxData          `bson:"data" json:"data"`
	CreateTime   string             `bson:"createTime,omitempty" json:"createTime,omitempty"` //创建时间
	UpdateTime   string             `bson:"updateTime,omitempty" json:"updateTime,omitempty"` //更新时间
}
type BoxData struct {
	SensorId    string          `bson:"sensorId" json:"sensorId"`                           //plc对应变量的地址位
	SensorName  string          `bson:"sensorName" json:"sensorName"`                       //Box设备名称
	StartTime   string          `bson:"startTime,omitempty" json:"startTime,omitempty"`     //石墨化送电时刻 其他工艺不用
	StoveNumber string          `bson:"stoveNumber,omitempty" json:"stoveNumber,omitempty"` //石墨化炉号 其他工艺不用
	Detail      []BoxDataDetail `bson:"detail" json:"detail"`                               //详细数据信息
}
type BoxDataDetail struct {
	Key   string `bson:"key" json:"key"`     //Box检测的key值
	Value string `bson:"value" json:"value"` //value值
	Unit  string `bson:"unit" json:"unit"`   //单位
}
type GraphitingDisplacementExcel struct {
	Name      string      `bson:"Name" json:"Name"`           //石墨化位移对照表
	ExcelData []ExcelData `bson:"ExcelData" json:"ExcelData"` //位移对照表数据
}
type ExcelData struct {
	Key   string `bson:"key" json:"key"`     //大位移值
	Value string `bson:"value" json:"value"` //对应小位移值
}
