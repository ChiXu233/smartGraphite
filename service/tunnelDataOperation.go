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

type TunnelType struct {
	SensorIdAndName string
	TunnelDetails   []model.BoxDataDetail
}
type TunnelSortKV struct {
	SensorIdAndName string
	KUV             []SortKV
}

//隧道窑Box数据定时存储
func TunnelDataOperation(interval time.Duration) {
	endTime := time.Now().Format("2006-01-02 15:04:05")
	startTime := time.Now().Add(-interval).Format("2006-01-02 15:04:05")
	var box model.Box
	err := global.TunnelDataColl.FindOne(context.TODO(), bson.M{}).Decode(&box)
	if err != nil {
		fmt.Println("TunnelDataOperation", err.Error())
	}
	var datas []model.Box
	filter := bson.M{
		"createTime": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
		"boxId": box.BoxId,
	}
	if err := utils.Find(global.TunnelHisDataColl, &datas, filter); err != nil {
		fmt.Println("TunnelDataOperation", err.Error())
	}
	if datas == nil {
		fmt.Println("隧道窑Box此时间段无数据")
		return
	}
	//规定好顺序,避免乱序问题
	var tunnelDatas []TunnelType
	var tunnelKV []TunnelSortKV
	for _, a := range datas[0].Data {
		tunnelDatas = append(tunnelDatas, TunnelType{
			SensorIdAndName: a.SensorId + "|" + a.SensorName,
			TunnelDetails:   nil,
		})
		var ku []SortKV
		for _, b := range a.Detail {
			ku = append(ku, SortKV{
				KeyAndUnit: b.Key + "|" + b.Unit,
				value:      nil,
			})
		}
		tunnelKV = append(tunnelKV, TunnelSortKV{
			SensorIdAndName: a.SensorId + "|" + a.SensorName,
			KUV:             ku,
		})
	}
	//分组传感器
	for _, data := range datas {
		for _, datum := range data.Data {
			for i, c := range tunnelDatas {
				if c.SensorIdAndName == datum.SensorId+"|"+datum.SensorName {
					tunnelDatas[i].TunnelDetails = append(tunnelDatas[i].TunnelDetails, datum.Detail...)
				}
			}
		}
	}
	var comTunnelDatas []model.BoxData
	//分组传感器中的检测值
	for _, tunnelData := range tunnelDatas {
		for _, datum := range tunnelData.TunnelDetails {
			for i, b := range tunnelKV {
				if b.SensorIdAndName == tunnelData.SensorIdAndName {
					for j, kv := range b.KUV {
						if kv.KeyAndUnit == datum.Key+"|"+datum.Unit {
							tunnelKV[i].KUV[j].value = append(tunnelKV[i].KUV[j].value, datum.Value)
						}
					}
				}
			}
		}
	}
	//求最大值平均值最小值
	for _, c := range tunnelKV {
		var comTunnelDataDetail []model.BoxDataDetail
		for _, d := range c.KUV {
			keyUnit := strings.Split(d.KeyAndUnit, "|")
			comTunnelDataDetail = append(comTunnelDataDetail, []model.BoxDataDetail{
				{
					Key:   keyUnit[0] + "平均值",
					Value: avg(d.value...),
					Unit:  keyUnit[1],
				},
				{
					Key:   keyUnit[0] + "最小值",
					Value: min(d.value...),
					Unit:  keyUnit[1],
				},
				{
					Key:   keyUnit[0] + "最大值",
					Value: max(d.value...),
					Unit:  keyUnit[1],
				},
			}...)
		}
		idName := strings.Split(c.SensorIdAndName, "|")
		comTunnelDatas = append(comTunnelDatas, model.BoxData{
			SensorId:   idName[0],
			SensorName: idName[1],
			Detail:     comTunnelDataDetail,
		})
	}
	box.Data = comTunnelDatas
	box.Id, box.UpdateTime, box.CreateTime = primitive.ObjectID{}, "", endTime
	storeTunnelBoxData(box, interval)
}

func storeTunnelBoxData(box model.Box, interval time.Duration) {
	var collectionHis *mongo.Collection
	var s string
	switch interval {
	case time.Minute * 10:
		collectionHis = global.TunnelTenMinDataColl
		s = "10分钟数据存储成功"
	case time.Minute * 30:
		collectionHis = global.TunnelThirtyMinDataColl
		s = "30分钟数据存储成功"
	case time.Hour:
		collectionHis = global.TunnelHourDataColl
		s = "60分钟数据存储成功"
	default:
		return
	}
	if _, err := collectionHis.InsertOne(context.TODO(), box); err != nil {
		fmt.Println("数据存储出错", err.Error())
		return
	}
	fmt.Println("隧道窑Box" + s)
}
