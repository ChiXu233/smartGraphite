package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"SmartGraphite-server/utils"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
	"time"
)

//成型PLC 选取指定数值计算
func FormPlcTrend() {
	//type InfoDataDetail struct {
	//	Data []string `json:"data" bson:"data"`
	//	Time []string `json:"time" bson:"time"`
	//	Unit string   `json:"unit" bson:"unit"`
	//}
	//
	//type DataDetail struct { //声明model.BigScreenLook中Data的类型
	//	Name string                    `json:"name" bson:"name"`
	//	Info map[string]InfoDataDetail `json:"info" bson:"info"`
	//}

	FormPlcId, _ := primitive.ObjectIDFromHex("62a888b0d8a1cbe11b55b494")
	deviceTypeIds := []primitive.ObjectID{
		FormPlcId,
	}

	updateTime := utils.TimeFormat(time.Now())
	for _, typeId := range deviceTypeIds {
		var bigScreenLook model.BigScreenLook
		var deviceDB []model.Device
		var data []model.DataDetail
		_ = utils.Find(global.DeviceColl, &deviceDB, bson.M{"deviceTypeId": typeId})
		for _, device := range deviceDB {
			var RealDataColl, HourDataColl *mongo.Collection
			var realData model.Box
			var hourDataDB []model.Box

			switch device.Code {
			case "65d27a491d744a0e91b4d8e6db628887": //成型PLC
				RealDataColl = global.FormPlcDataColl
				HourDataColl = global.FormPlcHourDataColl
			default:
				return
			}
			filter := bson.M{
				"boxId": device.Code,
				"createTime": bson.M{
					"$gte": utils.TimeFormat(time.Now().Add(-time.Hour * 25)),
					"$lte": utils.TimeFormat(time.Now()),
				},
			}
			_ = RealDataColl.FindOne(context.TODO(), bson.M{"boxId": device.Code}).Decode(&realData)
			_ = utils.Find(HourDataColl, &hourDataDB, filter)

			//数据处理
			nameMap := make(map[string]model.InfoDataDetail)
			for j, hourData := range hourDataDB { //24条数据
				if j == 0 { //数据格式化
					for i := range hourData.Data[0].Detail {
						var infoDataDetail model.InfoDataDetail
						if strings.Contains(hourData.Data[0].Detail[i].Key, "D地沟皮带电流平均值") || strings.Contains(hourData.Data[0].Detail[i].Key, "C煅烧风机电流平均值") { //详细数据遍历
							infoDataDetail.Data = append(infoDataDetail.Data, hourData.Data[0].Detail[i].Value)
							infoDataDetail.Time = append(infoDataDetail.Time, hourData.CreateTime)
							infoDataDetail.Unit = hourData.Data[0].Detail[i].Unit
							nameMap[strings.TrimRight(hourData.Data[0].Detail[i].Key, "平均值")] = infoDataDetail
						}
					}

				} else {
					for _, detail := range hourData.Data[0].Detail { //详细数据遍历
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
				Info: nameMap,
			})
		}
		bigScreenLook = model.BigScreenLook{
			Name:       "数据大屏/PLC",
			Code:       typeId.Hex(),
			Data:       data,
			Method:     "GET",
			Url:        "/getBigScreenLook",
			UpdateTime: updateTime,
		}
		err := global.FormPlcEchartsColl.FindOneAndUpdate(context.TODO(), bson.M{"code": bigScreenLook.Code}, bson.D{{"$set", bigScreenLook}}).Decode(&bson.M{})
		if err != nil {
			_, _ = global.FormPlcEchartsColl.InsertOne(context.TODO(), bigScreenLook)
		}
	}
}
