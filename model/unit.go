package model

import "go.mongodb.org/mongo-driver/bson/primitive"

//在数据包没有传变比或单位时可以使用动态转换单位
type Unit struct {
	Id          primitive.ObjectID `json:"id,omitempty" bson:"id,omitempty"`                 //Id
	TypeId      primitive.ObjectID `json:"typeId" bson:"typeId"`                             //设备类型id
	DTUId       string             `json:"DTUId" bson:"DTUId"`                               //设备编号
	IsValid     bool               `json:"isValid" bson:"isValid"`                           //是否启用
	SensorMerge bool               `json:"sensorMerge" bson:"sensorMerge"`                   //传感器是否合并
	Sensor      []UnitSensor       `json:"sensor" bson:"sensor"`                             //传感器信息
	CreateTime  string             `bson:"createTime,omitempty" json:"createTime,omitempty"` //创建时间
	UpdateTime  string             `bson:"updateTime,omitempty" json:"updateTime,omitempty"` //更新时间
}

//传感器
type UnitSensor struct {
	SensorId  string                   `json:"sensorId" bson:"sensorId"`                   //传感器编码
	Var       float64                  `json:"var" bson:"var"`                             //数据包变比,没有变比就为1
	DataMap   map[string]float64       `json:"dataMap,omitempty" bson:"dataMap,omitempty"` //单位值
	IsCal     map[string]bool          `json:"isCal" bson:"isCal"`                         //在计算10，15，30分钟等时间间隔时是否需要计算指定值
	IsEcharts map[string]bool          `json:"isEcharts" bson:"isEcharts"`                 //选择指定值计算echarts
	UnitMap   map[string]UnitMapDetail `json:"unitMap,omitempty" bson:"unitMap,omitempty"` //同DataMap作用
}

//指定解析单位详细信息
type UnitMapDetail struct {
	Value float64 `json:"value" bson:"value"` //单位值
	Flag  bool    `json:"flag" bson:"flag"`   //是否乘变比，在UnitMap中不用除法，在计算机中除法处理比乘法处理慢
}
