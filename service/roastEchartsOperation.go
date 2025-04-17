package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"SmartGraphite-server/utils"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
	"time"
)

func RoastEchartsTrend() {
	//type InfoDataDetail struct {
	//	Data []string `json:"data" bson:"data"`
	//	Time []string `json:"time" bson:"time"`
	//	Unit string   `json:"unit" bson:"unit"`
	//}
	//
	//type DataDetail struct { //声明model.BigScreenLook中Data的类型
	//	Name string                    `json:"name" bson:"name"`
	//	Code string                    `json:"code" bson:"code"`
	//	Info map[string]InfoDataDetail `json:"info" bson:"info"`
	//}

	RoastId, _ := primitive.ObjectIDFromHex("62874bb27cc89967383a5b80")
	deviceTypeIds := []primitive.ObjectID{
		RoastId,
	}

	updateTime := utils.TimeFormat(time.Now())
	for _, typeId := range deviceTypeIds {
		var bigScreenLook model.BigScreenLook
		var deviceDB []model.Device
		var data []model.DataDetail
		_ = utils.Find(global.DeviceColl, &deviceDB, bson.M{"deviceTypeId": typeId, "isValid": true})

		for _, device := range deviceDB {

			var hourDataDB []model.SensorData
			filter := bson.M{
				"sensorId": device.Code,
				"createTime": bson.M{
					"$gte": utils.TimeFormat(time.Now().Add(-time.Hour * 25)),
					"$lte": utils.TimeFormat(time.Now()),
				},
			}

			_ = utils.Find(global.CRoastHourDataColl, &hourDataDB, filter)

			//if device.Code == "b2" {
			//	fmt.Println(len(hourDataDB))
			//}
			//数据处理
			nameMap := make(map[string]model.InfoDataDetail)
			for j, hourData := range hourDataDB { //24条数据
				if j == 0 { //数据格式化
					for i := range hourData.DTUDataDetail {
						var infoDataDetail model.InfoDataDetail
						if strings.Contains(hourData.DTUDataDetail[i].Key, "平均值") { //详细数据遍历

							infoDataDetail.Data = append(infoDataDetail.Data, hourData.DTUDataDetail[i].Value)
							infoDataDetail.Time = append(infoDataDetail.Time, hourData.CreateTime)
							infoDataDetail.Unit = hourData.DTUDataDetail[i].Unit

							nameMap[strings.TrimRight(hourData.DTUDataDetail[i].Key, "平均值")] = infoDataDetail
						}
					}

				} else {
					for _, detail := range hourData.DTUDataDetail { //详细数据遍历
						for key := range nameMap {
							if detail.Key == key+"平均值" {
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

		err := global.RoastEcahrtsColl.FindOneAndUpdate(context.TODO(), bson.M{"code": bigScreenLook.Code}, bson.D{{"$set", bigScreenLook}}).Decode(&bson.M{})
		if err != nil {
			_, _ = global.RoastEcahrtsColl.InsertOne(context.TODO(), bigScreenLook)
		}
	}

}
