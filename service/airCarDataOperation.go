package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"SmartGraphite-server/utils"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
	"time"
)

type RunTime struct {
	Id           primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	DeviceTypeId primitive.ObjectID `bson:"deviceTypeId" json:"deviceTypeId"`                 //设备类型id 对应设备类型表
	BoxId        string             `bson:"boxId" json:"boxId"`                               //Box上唯一表示
	RunTime      string             `bson:"runTime" json:"runTime"`                           //运行时长
	Desc         string             `bson:"desc" json:"desc"`                                 //备注
	CreateTime   string             `bson:"createTime,omitempty" json:"createTime,omitempty"` //创建时间
}

//吸料天车运行数据
func AirCarDataOperation(interval time.Duration) {
	endTime := time.Now().Format("2006-01-02 15:04:05")
	startTime := time.Now().Add(-interval).Format("2006-01-02 15:04:05")
	objId, err := primitive.ObjectIDFromHex("61e3e52c6e10a0ee645ae3c4")
	if err != nil {
		fmt.Println(err)
	}
	cur, err := global.DeviceColl.Find(context.TODO(), bson.M{"deviceTypeId": objId})
	if err != nil {
		fmt.Println(err)
	}
	var devices []model.Device
	if err = cur.All(context.TODO(), &devices); err != nil {
		fmt.Println(err)
	}
	collections := make(map[string]*mongo.Collection)
	collections["f73fe0d8688046e088bb073849aa0c3f"] = global.EastAirCarHisDataColl
	collections["b46a0faf11cc4000a4c290eba5cc949a"] = global.WestAirCarHisDataColl
	for _, a := range devices {
		var datas []model.Box
		filter := bson.M{
			"createTime": bson.M{
				"$gte": startTime,
				"$lte": endTime,
			},
			"boxId": a.Code,
		}
		if err := utils.Find(collections[a.Code], &datas, filter); err != nil {
			fmt.Println("AirCarRunTime存储"+a.Name+"err:", err.Error())
		}
		if datas == nil {
			fmt.Println(a.Name + "在此时间段数据为空")
			continue
		}
		storeAirCarRunTimeData(endTime, a.Name, datas)
	}
	return
}

//存储吸料天车运行时间数据
func storeAirCarRunTimeData(endTime, desc string, datas []model.Box) {
	var airCar RunTime
	airCar.CreateTime = endTime
	var count int
	for i, a := range datas {
		if i == 0 {
			airCar.BoxId = a.BoxId
			airCar.DeviceTypeId = a.DeviceTypeId
		}
		for _, b := range a.Data {
			for _, c := range b.Detail {
				if c.Value == "1" {
					count++
				} else {
					continue
				}
			}
		}
	}
	airCar.RunTime = strconv.Itoa(count) + "分钟"
	airCar.Desc = desc + "数据"
	_, err := global.AirCarRunTimeColl.InsertOne(context.TODO(), airCar)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}
