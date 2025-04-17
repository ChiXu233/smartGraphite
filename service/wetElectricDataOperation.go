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
	"strings"
	"time"
)

//焙烧湿电数据定时存储
func RoastingDataOperation(interval time.Duration) {
	endTime := time.Now().Format("2006-01-02 15:04:05")
	startTime := time.Now().Add(-interval).Format("2006-01-02 15:04:05")
	var box model.Box
	err := global.RoastWetElectricDataColl.FindOne(context.TODO(), bson.M{}).Decode(&box)
	if err != nil {
		fmt.Println("RoastingDataOperation", err.Error())
	}
	var datas []model.Box
	filter := bson.M{
		"createTime": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
		"boxId": box.BoxId,
	}
	if err := utils.Find(global.RoastWetElectricHisDataColl, &datas, filter); err != nil {
		fmt.Println("RoastingDataOperation", err.Error())
	}
	if datas == nil {
		fmt.Println("焙烧湿电此时间段无数据")
		return
	}
	//规定好顺序,避免乱序问题
	var wetDatas []HisDataType
	var ku []SortKV
	for _, a := range datas[0].Data {
		wetDatas = append(wetDatas, HisDataType{
			SensorIdAndName: a.SensorId + "|" + a.SensorName,
			Details:         nil,
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
			for i, c := range wetDatas {
				if c.SensorIdAndName == datum.SensorId+"|"+datum.SensorName {
					wetDatas[i].Details = append(wetDatas[i].Details, datum.Detail...)
				}
			}
		}
	}
	var comWetDatas []model.BoxData
	var count int
	//分组传感器中的检测值
	for _, graData := range wetDatas {
		for _, datum := range graData.Details {
			for i, b := range ku {
				if b.KeyAndUnit == datum.Key+"|"+datum.Unit {
					ku[i].value = append(ku[i].value, datum.Value)
				}
			}
		}
		//求最大值平均值最小值
		var comWetDataDetail []model.BoxDataDetail
		for i, c := range ku {
			keyUnit := strings.Split(c.KeyAndUnit, "|")
			if i <= 8 {
				if keyUnit[0] == "运行" {
					for _, r := range c.value {
						if r == "1" {
							count++
						}
					}
				}
				comWetDataDetail = append(comWetDataDetail, []model.BoxDataDetail{
					{
						Key:   keyUnit[0],
						Value: c.value[len(c.value)-1],
						Unit:  keyUnit[1],
					},
				}...)
			} else {
				comWetDataDetail = append(comWetDataDetail, []model.BoxDataDetail{
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
		}
		idName := strings.Split(graData.SensorIdAndName, "|")
		comWetDatas = append(comWetDatas, model.BoxData{
			SensorId:   idName[0],
			SensorName: idName[1],
			Detail:     comWetDataDetail,
		})
	}
	if len(comWetDatas) == 1 {
		comWetDatas[0].Detail = append(comWetDatas[0].Detail, model.BoxDataDetail{
			Key:   "运行时间",
			Value: strconv.Itoa(count),
			Unit:  "分钟",
		})
	}
	box.Data = comWetDatas
	box.Id, box.UpdateTime, box.CreateTime = primitive.ObjectID{}, "", endTime
	storeRoastWetData(box, interval)
}

//隧道窑湿电数据定时存储
func TunWetDataOperation(interval time.Duration) {
	endTime := time.Now().Format("2006-01-02 15:04:05")
	startTime := time.Now().Add(-interval).Format("2006-01-02 15:04:05")
	var box model.Box
	err := global.TunWetElectricDataColl.FindOne(context.TODO(), bson.M{}).Decode(&box)
	if err != nil {
		fmt.Println("TunnelWetDataOperation", err.Error())
	}
	var datas []model.Box
	filter := bson.M{
		"createTime": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
		"boxId": box.BoxId,
	}
	if err := utils.Find(global.TunWetElectricHisDataColl, &datas, filter); err != nil {
		fmt.Println("TunnelWetDataOperation", err.Error())
	}
	if datas == nil {
		fmt.Println("隧道窑湿电此时间段无数据")
		return
	}
	//规定好顺序,避免乱序问题
	var wetDatas []HisDataType
	var ku []SortKV
	for _, a := range datas[0].Data {
		wetDatas = append(wetDatas, HisDataType{
			SensorIdAndName: a.SensorId + "|" + a.SensorName,
			Details:         nil,
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
			for i, c := range wetDatas {
				if c.SensorIdAndName == datum.SensorId+"|"+datum.SensorName {
					wetDatas[i].Details = append(wetDatas[i].Details, datum.Detail...)
				}
			}
		}
	}
	var comWetDatas []model.BoxData
	var count int
	//分组传感器中的检测值
	for _, graData := range wetDatas {
		for _, datum := range graData.Details {
			for i, b := range ku {
				if b.KeyAndUnit == datum.Key+"|"+datum.Unit {
					ku[i].value = append(ku[i].value, datum.Value)
				}
			}
		}
		//求最大值平均值最小值
		var comWetDataDetail []model.BoxDataDetail
		for i, c := range ku {
			keyUnit := strings.Split(c.KeyAndUnit, "|")
			if i <= 8 {
				if keyUnit[0] == "运行" {
					for _, r := range c.value {
						if r == "1" {
							count++
						}
					}
				}
				comWetDataDetail = append(comWetDataDetail, []model.BoxDataDetail{
					{
						Key:   keyUnit[0],
						Value: c.value[len(c.value)-1],
						Unit:  keyUnit[1],
					},
				}...)
			} else {
				comWetDataDetail = append(comWetDataDetail, []model.BoxDataDetail{
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
		}
		idName := strings.Split(graData.SensorIdAndName, "|")
		comWetDatas = append(comWetDatas, model.BoxData{
			SensorId:   idName[0],
			SensorName: idName[1],
			Detail:     comWetDataDetail,
		})
	}
	if len(comWetDatas) == 1 {
		comWetDatas[0].Detail = append(comWetDatas[0].Detail, model.BoxDataDetail{
			Key:   "运行时间",
			Value: strconv.Itoa(count),
			Unit:  "分钟",
		})
	}
	box.Data = comWetDatas
	box.Id, box.UpdateTime, box.CreateTime = primitive.ObjectID{}, "", endTime
	storeTunnelWetData(box, interval)
}

//石墨化湿电数据定时存储
func GraWetDataOperation(interval time.Duration) {
	endTime := time.Now().Format("2006-01-02 15:04:05")
	startTime := time.Now().Add(-interval).Format("2006-01-02 15:04:05")
	var box model.Box
	err := global.GraWetElectricDataColl.FindOne(context.TODO(), bson.M{}).Decode(&box)
	if err != nil {
		fmt.Println("GraphiteWetDataOperation", err.Error())
	}
	var datas []model.Box
	filter := bson.M{
		"createTime": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
		"boxId": box.BoxId,
	}
	if err := utils.Find(global.GraWetElectricHisDataColl, &datas, filter); err != nil {
		fmt.Println("GraphiteWetDataOperation", err.Error())
	}
	if datas == nil {
		fmt.Println("石墨化湿电此时间段无数据")
		return
	}
	//规定好顺序,避免乱序问题
	var wetDatas []HisDataType
	var ku []SortKV
	for _, a := range datas[0].Data {
		wetDatas = append(wetDatas, HisDataType{
			SensorIdAndName: a.SensorId + "|" + a.SensorName,
			Details:         nil,
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
			for i, c := range wetDatas {
				if c.SensorIdAndName == datum.SensorId+"|"+datum.SensorName {
					wetDatas[i].Details = append(wetDatas[i].Details, datum.Detail...)
				}
			}
		}
	}
	var comWetDatas []model.BoxData
	var count int
	//分组传感器中的检测值
	for _, graData := range wetDatas {
		for _, datum := range graData.Details {
			for i, b := range ku {
				if b.KeyAndUnit == datum.Key+"|"+datum.Unit {
					ku[i].value = append(ku[i].value, datum.Value)
				}
			}
		}
		//求最大值平均值最小值
		var comWetDataDetail []model.BoxDataDetail
		for i, c := range ku {
			keyUnit := strings.Split(c.KeyAndUnit, "|")
			if i <= 8 {
				if keyUnit[0] == "运行" {
					for _, r := range c.value {
						if r == "1" {
							count++
						}
					}
				}
				comWetDataDetail = append(comWetDataDetail, []model.BoxDataDetail{
					{
						Key:   keyUnit[0],
						Value: c.value[len(c.value)-1],
						Unit:  keyUnit[1],
					},
				}...)
			} else {
				comWetDataDetail = append(comWetDataDetail, []model.BoxDataDetail{
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
		}
		idName := strings.Split(graData.SensorIdAndName, "|")
		comWetDatas = append(comWetDatas, model.BoxData{
			SensorId:   idName[0],
			SensorName: idName[1],
			Detail:     comWetDataDetail,
		})
	}
	if len(comWetDatas) == 1 {
		comWetDatas[0].Detail = append(comWetDatas[0].Detail, model.BoxDataDetail{
			Key:   "运行时间",
			Value: strconv.Itoa(count),
			Unit:  "分钟",
		})
	}
	box.Data = comWetDatas
	box.Id, box.UpdateTime, box.CreateTime = primitive.ObjectID{}, "", endTime
	storeGraphiteWetData(box, interval)
}

//存储焙烧湿电数据
func storeRoastWetData(box model.Box, interval time.Duration) {
	var collectionHis *mongo.Collection
	var s string
	switch interval {
	case time.Minute * 10:
		collectionHis = global.RoastWetElectricTenDataColl
		s = "10分钟数据存储成功"
	case time.Minute * 30:
		collectionHis = global.RoastWetElectricThirtyDataColl
		s = "30分钟数据存储成功"
	case time.Hour:
		collectionHis = global.RoastWetElectricHourDataColl
		s = "60分钟数据存储成功"
	default:
		return
	}
	if _, err := collectionHis.InsertOne(context.TODO(), box); err != nil {
		fmt.Println("数据存储出错", err.Error())
		return
	}
	fmt.Println("焙烧湿电" + s)
}

//存储隧道窑湿电数据
func storeTunnelWetData(box model.Box, interval time.Duration) {
	var collectionHis *mongo.Collection
	var s string
	switch interval {
	case time.Minute * 10:
		collectionHis = global.TunWetElectricTenDataColl
		s = "10分钟数据存储成功"
	case time.Minute * 30:
		collectionHis = global.TunWetElectricThirtyDataColl
		s = "30分钟数据存储成功"
	case time.Hour:
		collectionHis = global.TunWetElectricHourDataColl
		s = "60分钟数据存储成功"
	default:
		return
	}
	if _, err := collectionHis.InsertOne(context.TODO(), box); err != nil {
		fmt.Println("数据存储出错", err.Error())
		return
	}
	fmt.Println("隧道窑湿电" + s)
}

//存储石墨化湿电数据
func storeGraphiteWetData(box model.Box, interval time.Duration) {
	var collectionHis *mongo.Collection
	var s string
	switch interval {
	case time.Minute * 10:
		collectionHis = global.GraWetElectricTenDataColl
		s = "10分钟数据存储成功"
	case time.Minute * 30:
		collectionHis = global.GraWetElectricThirtyDataColl
		s = "30分钟数据存储成功"
	case time.Hour:
		collectionHis = global.GraWetElectricHourDataColl
		s = "60分钟数据存储成功"
	default:
		return
	}
	if _, err := collectionHis.InsertOne(context.TODO(), box); err != nil {
		fmt.Println("数据存储出错", err.Error())
		return
	}
	fmt.Println("石墨化湿电" + s)
}
