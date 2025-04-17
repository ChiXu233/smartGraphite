package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"SmartGraphite-server/utils"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
)

//近24小时机加工除尘电流数据
func ThreeElectricityECharsOperation() {
	ThreeElectricityTypeId, _ := primitive.ObjectIDFromHex("632af40c1957a532016a1ce8")
	deviceTypeIds := []primitive.ObjectID{
		ThreeElectricityTypeId,
	}

	updateTime := utils.TimeFormat(time.Now())
	for _, typeId := range deviceTypeIds {
		var bigScreenLook model.BigScreenLook
		var deviceDB []model.Device
		var data []model.DataDetail
		_ = utils.Find(global.DeviceColl, &deviceDB, bson.M{"deviceTypeId": typeId})

		for _, device := range deviceDB {
			var hourDataDB []model.DTU
			filter := bson.M{
				"DTUId": device.Code,
				"createTime": bson.M{
					"$gte": utils.TimeFormat(time.Now().Add(-time.Hour * 25)),
					"$lte": utils.TimeFormat(time.Now()),
				},
			}

			opts := options.Find().SetLimit(24)

			_ = utils.Find(global.ThreeElectricityHourDataColl, &hourDataDB, filter, opts)
			//_ = utils.Find(global.ThreeElectricityHourDataColl, &hourDataDB, bson.M{}, opts)
			//数据处理
			nameMap := make(map[string]model.InfoDataDetail)
			for j, hourData := range hourDataDB { //24条数据
				if j == 0 { //数据格式化
					for _, sensor := range hourData.DTUData {
						for _, detail := range sensor.DTUDataDetail {
							if strings.Contains(detail.Key, "平均值") && strings.Contains(detail.Key, "总有") || strings.Contains(detail.Key, "平均值") && strings.Contains(detail.Key, "总无") {
								var infoDataDetail model.InfoDataDetail
								infoDataDetail.Data = append(infoDataDetail.Data, detail.Value)
								infoDataDetail.Time = append(infoDataDetail.Time, hourData.CreateTime)
								infoDataDetail.Unit = detail.Unit

								//if strings.Contains(strings.TrimRight(detail.Key, "平均值")+sensor.SensorId, "b8") {
								//	fmt.Println(strings.TrimRight(detail.Key, "平均值")+sensor.SensorId, "test")
								//}
								nameMap[strings.TrimRight(detail.Key, "平均值")] = infoDataDetail
								//fmt.Println(strings.TrimRight(detail.Key, "平均值")+sensor.SensorId,nameMap[strings.TrimRight(detail.Key, "平均值")+sensor.SensorId])
							}
						}
					}

				} else {
					for _, sensor := range hourData.DTUData {
						for _, detail := range sensor.DTUDataDetail {
							for key := range nameMap {
								if strings.TrimRight(detail.Key, "平均值") == key {
									//map不能在地址上被直接修改值
									infoDataDetail := nameMap[key]
									infoDataDetail.Data = append(infoDataDetail.Data, detail.Value)
									infoDataDetail.Time = append(infoDataDetail.Time, hourData.CreateTime)

									nameMap[key] = infoDataDetail
								}
							}
						}
					}

				}
			}

			data = append(data, model.DataDetail{
				Name: device.Sensors[0].Name,
				Code: device.Code,
				Info: nameMap,
			})

		}
		bigScreenLook = model.BigScreenLook{
			Name:       "数据大屏/DTU",
			Code:       typeId.Hex(),
			Data:       data,
			Method:     "GET",
			Url:        "",
			UpdateTime: updateTime,
		}

		err := global.ThreeElectricityECharsColl.FindOneAndUpdate(context.TODO(), bson.M{"code": bigScreenLook.Code}, bson.D{{"$set", bigScreenLook}}).Decode(&bson.M{})
		if err != nil {
			_, _ = global.ThreeElectricityECharsColl.InsertOne(context.TODO(), bigScreenLook)
		}
	}
}
