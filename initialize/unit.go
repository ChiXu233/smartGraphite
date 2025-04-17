package initialize

import (
	"SmartGraphite-server/global"
	"go.mongodb.org/mongo-driver/mongo"
)

// 动态协议接收的设备用
func UnitInit() {
	//编号补0会造成很多而外的计算开销,直接拼接即可
	//string key的生成方式为设备编号+传感器编号:deviceCode+sensorId
	if global.CollMap == nil {
		global.CollMap = make(map[string]map[int]*mongo.Collection)

		global.CollMap["c11"] = make(map[int]*mongo.Collection)
		global.CollMap["c11"][1] = global.CRoastHisDataColl
		global.CollMap["c11"][10] = global.CRoastTenMinDataColl
		global.CollMap["c11"][30] = global.CRoastThirtyDataColl
		global.CollMap["c11"][60] = global.CRoastHourDataColl

		global.CollMap["ba1"] = make(map[int]*mongo.Collection)
		global.CollMap["ba1"][1] = global.BATransducer
		global.CollMap["ba1"][10] = global.BATransducerTenData
		global.CollMap["ba1"][30] = global.BATransducerThirtyData
		global.CollMap["ba1"][60] = global.BATransducerHourData

		global.CollMap["bb1"] = make(map[int]*mongo.Collection)
		global.CollMap["bb1"][1] = global.BBTransducer
		global.CollMap["bb1"][10] = global.BBTransducerTenData
		global.CollMap["bb1"][30] = global.BBTransducerThirtyData
		global.CollMap["bb1"][60] = global.BBTransducerHourData

		global.CollMap["bc1"] = make(map[int]*mongo.Collection)
		global.CollMap["bc1"][1] = global.BCTransducer
		global.CollMap["bc1"][10] = global.BCTransducerTenData
		global.CollMap["bc1"][30] = global.BCTransducerThirtyData
		global.CollMap["bc1"][60] = global.BCTransducerHourData

		global.CollMap["d11"] = make(map[int]*mongo.Collection)
		global.CollMap["d11"][1] = global.D1Transducer
		global.CollMap["d11"][10] = global.D1TransducerTenData
		global.CollMap["d11"][30] = global.D1TransducerThirtyData
		global.CollMap["d11"][60] = global.D1TransducerHourData

		global.CollMap["d21"] = make(map[int]*mongo.Collection)
		global.CollMap["d21"][1] = global.D2Transducer
		global.CollMap["d21"][10] = global.D2TransducerTenData
		global.CollMap["d21"][30] = global.D2TransducerThirtyData
		global.CollMap["d21"][60] = global.D2TransducerHourData

		global.CollMap["d31"] = make(map[int]*mongo.Collection)
		global.CollMap["d31"][1] = global.D3Transducer
		global.CollMap["d31"][10] = global.D3TransducerTenData
		global.CollMap["d31"][30] = global.D3TransducerThirtyData
		global.CollMap["d31"][60] = global.D3TransducerHourData

		global.CollMap["ee1"] = make(map[int]*mongo.Collection)
		global.CollMap["ee1"][1] = global.RoastTemp

		global.CollMap["e21"] = make(map[int]*mongo.Collection)
		global.CollMap["e21"][1] = global.E2RoastTemp
		global.CollMap["e21"][101] = global.E2RoastTempMin

		global.CollMap["e51"] = make(map[int]*mongo.Collection)
		global.CollMap["e51"][1] = global.E5RoastTemp
		global.CollMap["e51"][101] = global.E5RoastTempMin

		global.CollMap["e41"] = make(map[int]*mongo.Collection)
		global.CollMap["e41"][1] = global.E4RoastTemp
		global.CollMap["e41"][101] = global.E4RoastTempMin

	}
}
