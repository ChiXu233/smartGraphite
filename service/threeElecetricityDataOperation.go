package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

func ThreeElectricityDataOperation(interval time.Duration) {
	endTime := time.Now().Format("2006-01-02 15:04:05")
	startTime := time.Now().Add(-interval).Format("2006-01-02 15:04:05")
	//根据氨气逃逸类型查询设备
	deviceTypeId, _ := primitive.ObjectIDFromHex("632af40c1957a532016a1ce8")

	var devices []model.Device
	curr, err := global.DeviceColl.Find(context.TODO(), bson.M{"deviceTypeId": deviceTypeId, "isValid": true, "status": "正常"})
	if err != nil {
		log.Println("三相电表监测设备", err)
		return
	}
	if err := curr.All(context.TODO(), &devices); err != nil {
		log.Println("三相电表监测设备", err)
		return
	}

	//使用下标索引
	for devI := range devices { //单个设备依次计算
		//待插入数据库中的字段
		var info model.DTU
		//切片空间增加
		info.DTUData = append(info.DTUData, model.DTUData{})

		//根据设备绑定的dtu编码在历史表中查询数据
		var dtuDB []model.DTU
		filter := bson.M{
			"createTime": bson.M{
				"$gte": startTime,
				"$lte": endTime,
			},
			"DTUId": devices[devI].Code,
		}
		//根据filter条件查询历史表
		curr, err := global.ThreeElectricityHisDataColl.Find(context.TODO(), filter)
		if err != nil {
			log.Println("三相电表监测历史数据", err)
			return
		}
		if err = curr.All(context.TODO(), &dtuDB); err != nil {
			log.Println("三相电表监测历史数据", err)
			return
		}
		if dtuDB == nil {
			log.Println(devices[devI].Name, "此设备在此时间段没有数据")
			continue
		}

		//数据计算
		for i := range dtuDB { //单个设备数据计算
			//多个传感器数据计算
			if i == 0 { //第一条数据
				info.CreateTime = endTime
				info.DTUId = dtuDB[i].DTUId
				var DTUDatas []model.DTUData
				//传感器遍历
				for _, sensor := range dtuDB[i].DTUData {
					var DTUData model.DTUData
					//详细数据遍历
					DTUData.SensorId = sensor.SensorId
					DTUData.SensorName = sensor.SensorName
					for _, detail := range sensor.DTUDataDetail {
						DTUData.DTUDataDetail = append(DTUData.DTUDataDetail, []model.DTUDataDetail{
							{
								Key:   detail.Key + "最大值",
								Value: detail.Value,
								Unit:  detail.Unit,
							},
							{
								Key:   detail.Key + "最小值",
								Value: detail.Value,
								Unit:  detail.Unit,
							},
							{
								Key:   detail.Key + "平均值",
								Value: detail.Value,
								Unit:  detail.Unit,
							},
						}...)
					}
					DTUDatas = append(DTUDatas, DTUData) //传感器数组增加
					info.DTUData = DTUDatas
				}

			} else { //第二条记录开始,数据在遍历的同时进行比较和累加，最后再求平均值
				for sen := range dtuDB[i].DTUData {
					for key, detail := range dtuDB[i].DTUData[sen].DTUDataDetail {
						if detail.Key+"最大值" == info.DTUData[sen].DTUDataDetail[key].Key {
							//最大值
							info.DTUData[sen].DTUDataDetail[key].Value = stringMax(info.DTUData[sen].DTUDataDetail[key].Value, detail.Value)
							//最小值
							info.DTUData[sen].DTUDataDetail[key+1].Value = stringMax(info.DTUData[sen].DTUDataDetail[key].Value, detail.Value)
							//平均值
							info.DTUData[sen].DTUDataDetail[key+2].Value = stringAdd(info.DTUData[sen].DTUDataDetail[key].Value, detail.Value)
						}
					}
				}

			}

		}

		//数据累加完毕，求平均
		dtuDBLen := len(dtuDB)
		detailLen := len(info.DTUData[0].DTUDataDetail)
		for i := range info.DTUData {
			for key := 2; key < detailLen; key += 3 {
				info.DTUData[i].DTUDataDetail[key].Value = DivisionTen(info.DTUData[i].DTUDataDetail[key].Value, dtuDBLen)
			}
		}

		//数据累加完毕，求平均
		//dtuDBLen := len(dtuDB)
		//detailLen := len(info.DTUData[0].DTUDataDetail)
		//for i := 2; i < detailLen; i += 3 {
		//	info.DTUData[0].DTUDataDetail[i].Value = DivisionTen(info.DTUData[0].DTUDataDetail[i].Value, dtuDBLen)
		//}

		//数据存储
		ThreeElectricityData(info, interval)
	}

}

func ThreeElectricityData(DTU model.DTU, interval time.Duration) {
	var collectionHis *mongo.Collection
	var s string
	switch interval {
	case time.Minute * 10:
		collectionHis = global.ThreeElectricityTenMinDataColl
		s = "10分钟数据存储成功"
	case time.Minute * 30:
		collectionHis = global.ThreeElectricityThirtyMinDataColl
		s = "30分钟数据存储成功"
	case time.Hour:
		collectionHis = global.ThreeElectricityHourDataColl
		s = "60分钟数据存储成功"
	default:
		return
	}

	if _, err := collectionHis.InsertOne(context.TODO(), DTU); err != nil {
		fmt.Println("数据存储出错", err.Error())
		return
	}
	fmt.Println("三相电表监测" + s)
}
