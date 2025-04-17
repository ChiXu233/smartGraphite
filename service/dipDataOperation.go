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
	"strings"
	"time"
)

type HisDataType struct {
	SensorIdAndName string
	Details         []model.BoxDataDetail
}
type SortKV struct {
	KeyAndUnit string
	value      []string
}

//浸渍数据定时存储
func DipDataOperation(interval time.Duration) {
	endTime := time.Now().Format("2006-01-02 15:04:05")
	startTime := time.Now().Add(-interval).Format("2006-01-02 15:04:05")
	var box model.Box
	err := global.DipDataColl.FindOne(context.TODO(), bson.M{}).Decode(&box)
	if err != nil {
		fmt.Println("DipDataOperation", err.Error())
	}
	var datas []model.Box
	filter := bson.M{
		"createTime": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
		"boxId": box.BoxId,
	}
	if err := utils.Find(global.DipDataHisColl, &datas, filter); err != nil {
		fmt.Println("DipDataOperation", err.Error())
	}
	if datas == nil {
		fmt.Println("浸渍此时间段无数据")
		return
	}
	//规定好顺序,避免乱序问题
	var dipDatas []HisDataType
	var ku []SortKV
	for _, a := range datas[0].Data {
		dipDatas = append(dipDatas, HisDataType{
			SensorIdAndName: a.SensorId + "|" + a.SensorName,
			Details:      nil,
		})
		for _, b := range a.Detail {
			ku = append(ku, SortKV{
				KeyAndUnit: b.Key + "|" + b.Unit,
				value:      nil,
			})
		}
	}
	//分组传感器
	for _, data := range datas {
		for _, datum := range data.Data {
			for i, c := range dipDatas {
				if c.SensorIdAndName == datum.SensorId+"|"+datum.SensorName {
					dipDatas[i].Details = append(dipDatas[i].Details, datum.Detail...)
				}
			}
		}
	}
	var comDipDatas []model.BoxData
	//分组传感器中的检测值
	for _, dipData := range dipDatas {
		for _, datum := range dipData.Details {
			for i, b := range ku {
				if b.KeyAndUnit == datum.Key+"|"+datum.Unit {
					ku[i].value = append(ku[i].value, datum.Value)
				}
			}
		}
		//求最大值平均值最小值
		var comDipDataDetail []model.BoxDataDetail
		for _, c := range ku {
			keyUnit := strings.Split(c.KeyAndUnit, "|")
			comDipDataDetail = append(comDipDataDetail, []model.BoxDataDetail{
				{
					Key:   keyUnit[0] + "平均值",
					Value: avg(c.value...),
					Unit:  keyUnit[1],
				},
				{
					Key:   keyUnit[0] + "最小值",
					Value: min(c.value...),
					Unit:  keyUnit[1],
				},
				{
					Key:   keyUnit[0] + "最大值",
					Value: max(c.value...),
					Unit:  keyUnit[1],
				},
			}...)
		}
		idName := strings.Split(dipData.SensorIdAndName, "|")
		comDipDatas = append(comDipDatas, model.BoxData{
			SensorId:   idName[0],
			SensorName: idName[1],
			Detail:     comDipDataDetail,
		})
	}
	box.Data = comDipDatas
	box.Id, box.UpdateTime, box.CreateTime = primitive.ObjectID{}, "", endTime
	storeDipData(box, interval)
}

func storeDipData(box model.Box, interval time.Duration) {
	var collectionHis *mongo.Collection
	var s string
	switch interval {
	case time.Minute * 10:
		collectionHis = global.DipTenMinDataColl
		s = "10分钟数据存储成功"
	case time.Minute * 30:
		collectionHis = global.DipThirtyMinDataColl
		s = "30分钟数据存储成功"
	case time.Hour:
		collectionHis = global.DipHourDataColl
		s = "60分钟数据存储成功"
	default:
		return
	}
	if _, err := collectionHis.InsertOne(context.TODO(), box); err != nil {
		fmt.Println("数据存储出错", err.Error())
		return
	}
	fmt.Println("浸渍" + s)
}
