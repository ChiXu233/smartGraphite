package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// DeviceDataOperation 传入设备Id，选择时间间隔后开始计算，适用于使用动态协议接收数据的设备
func DeviceDataOperation(idStr string, inter int) {

	interval := time.Duration(inter)
	endTime := time.Now().Format("2006-01-02 15:04:05")
	startTime := time.Now().Add(-interval * time.Minute).Format("2006-01-02 15:04:05")

	var collHis *mongo.Collection
	switch idStr {
	case "642401d201972e9942398321":
		//ba
		collHis = global.BATransducer
	case "6424049f01972e9942398339":
		//bb
		collHis = global.BBTransducer
	case "6423f6ba01972e994239829f":
		//bc
		collHis = global.BCTransducer
	case "6426410bbda900f9bafd1f50":
		//d1
		collHis = global.D1Transducer
	case "64264129bda900f9bafd1f54":
		//d2
		collHis = global.D2Transducer
	case "642640c9bda900f9bafd1f49":
		//d3
		collHis = global.D3Transducer

	default:
		fmt.Println("没有添加设备 " + idStr + " 的原始数据库")
		return
	}

	filter := bson.M{
		"createTime": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	//查找设备
	var deviceDB model.Device
	err = global.DeviceColl.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&deviceDB)
	if err != nil {
		fmt.Println(err)
		return
	}

	//找到对应设备的Unit
	var unitDB model.Unit
	err = global.ElectricityUnitColl.FindOne(context.TODO(), bson.M{"DTUId": deviceDB.Code}).Decode(&unitDB)
	if err != nil {
		fmt.Println(err)
		return
	}

	if unitDB.Sensor == nil {
		fmt.Println("传感器单位为空")
		return
	}

	var sensorDB []model.UnitSensor
	sensorDB = append(sensorDB, unitDB.Sensor...)
	unitMap := make(map[string]map[string]bool)
	for i := range sensorDB {
		unitMap[sensorDB[i].SensorId] = make(map[string]bool)
		unitMap[sensorDB[i].SensorId] = sensorDB[i].IsCal

	}

	for _, sensors := range deviceDB.Sensors {

		//在设备表中找相应传感器的数据
		res, err := collHis.Find(context.TODO(), filter)
		if err != nil {
			fmt.Println(err)
			continue
		}
		var sensorHis []model.SensorData
		if err = res.All(context.TODO(), &sensorHis); err != nil {
			fmt.Println(err)
			continue
		}

		//没有找到数据
		if sensorHis == nil {
			fmt.Println("没找到数据")
			continue
		}

		//待插入的数据
		var sensorData model.SensorData
		sensorData.Code = sensors.Code
		sensorData.Name = sensors.Name
		sensorData.CreateTime = endTime[:len(endTime)-2] + "00" //格式化秒

		//开始处理数据
		for i := range sensorHis {
			if i == 0 { //初始化数据
				for _, detail := range sensorHis[i].DTUDataDetail {
					sensorData.DTUDataDetail = append(sensorData.DTUDataDetail, []model.DTUDataDetail{
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
			} else { //数据在遍历的同时进行比较和累加，最后求平均值
				detailLen := len(sensorHis[i].DTUDataDetail)
				for k := 0; k < detailLen; k++ {
					temp := k * 3
					if sensorHis[i].DTUDataDetail[k].Key+"最大值" == sensorData.DTUDataDetail[temp].Key && unitMap[sensors.Code][sensorHis[i].DTUDataDetail[k].Key] { //找到对应的最大值位置
						//最大值
						sensorData.DTUDataDetail[temp].Value = stringMax(sensorData.DTUDataDetail[temp].Value, sensorHis[i].DTUDataDetail[k].Value)
						//最小值
						sensorData.DTUDataDetail[temp+1].Value = stringMin(sensorData.DTUDataDetail[temp+1].Value, sensorHis[i].DTUDataDetail[k].Value)
						//平均值
						sensorData.DTUDataDetail[temp+2].Value = stringAdd(sensorData.DTUDataDetail[temp+2].Value, sensorHis[i].DTUDataDetail[k].Value)
					}
				}

			}
		}

		//数据累加结束，求平均值
		sensorDataLen := len(sensorHis)
		detailLen := len(sensorData.DTUDataDetail)
		for i := 2; i < detailLen; i += 3 {
			sensorData.DTUDataDetail[i].Value = DivisionTen(sensorData.DTUDataDetail[i].Value, sensorDataLen)
		}

		//数据存储
		StoreDeviceData(deviceDB.Code, sensors.Code, inter, sensorData)

	}

}

// StoreDeviceData 可把写在代码中的
func StoreDeviceData(deviceCode, sensorCode string, interval int, data model.SensorData) {

	if deviceCode == "" || sensorCode == "" {
		return
	}

	var coll *mongo.Collection
	//获取到指定数据库
	coll = global.CollMap[deviceCode+sensorCode][interval]
	if coll == nil {
		fmt.Println(deviceCode, interval, "没有做表映射")
		return
	}
	err := coll.FindOneAndUpdate(context.TODO(), bson.M{"createTime": data.CreateTime}, bson.M{"$set": data}).Decode(&bson.M{})
	if err != nil {
		_, err = coll.InsertOne(context.TODO(), data)
		if err != nil {
			fmt.Println("设备编号为：" + deviceCode + " 数据存储更新失败")
		}
	}

	fmt.Println("设备编号为："+deviceCode, "的设备", interval, "分钟存储成功")

}
