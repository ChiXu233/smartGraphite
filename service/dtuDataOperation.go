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

type DtuType struct {
	SensorIdAndName string
	DtuDetails      []model.DTUDataDetail
}

//Dtu数据定时存储
func DataOperation(interval time.Duration) {
	endTime := time.Now().Format("2006-01-02 15:04:05")
	startTime := time.Now().Add(-interval).Format("2006-01-02 15:04:05")
	//据此获取设备code
	var DTU model.DTU
	err := global.DTUDataColl.FindOne(context.TODO(), bson.M{}).Decode(&DTU)
	if err != nil {
		fmt.Println("DataOperation", err.Error())
	}
	var datas []model.DTU
	filter := bson.M{
		"createTime": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
		"DTUId": DTU.DTUId,
	}
	if err := utils.Find(global.DTUHisDataColl, &datas, filter); err != nil {
		fmt.Println("DataOperation", err.Error())
	}
	if datas == nil {
		fmt.Println("Dtu此时间段无数据")
		return
	}
	//规定好顺序,避免乱序问题
	var dtuDatas []DtuType
	var ku []SortKV
	for _, a := range datas[0].DTUData {
		dtuDatas = append(dtuDatas, DtuType{
			SensorIdAndName: a.SensorId + "|" + a.SensorName,
			DtuDetails:      nil,
		})
		for _, b := range a.DTUDataDetail {
			ku = append(ku, SortKV{
				KeyAndUnit: b.Key + "|" + b.Unit,
				value:      nil,
			})
		}
	}
	//分组传感器
	for _, data := range datas {
		for _, datum := range data.DTUData {
			for i, c := range dtuDatas {
				if c.SensorIdAndName == datum.SensorId+"|"+datum.SensorName {
					dtuDatas[i].DtuDetails = append(dtuDatas[i].DtuDetails, datum.DTUDataDetail...)
				}
			}
		}
	}
	var comDTUDatas []model.DTUData
	//分组传感器中的检测值
	for _, dtuData := range dtuDatas {
		for _, datum := range dtuData.DtuDetails {
			for i, b := range ku {
				if b.KeyAndUnit == datum.Key+"|"+datum.Unit {
					ku[i].value = append(ku[i].value, datum.Value)
				}
			}
		}
		//求最大值平均值最小值
		var comDtuDataDetail []model.DTUDataDetail
		for _, c := range ku {
			keyUnit := strings.Split(c.KeyAndUnit, "|")
			comDtuDataDetail = append(comDtuDataDetail, []model.DTUDataDetail{
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
		idName := strings.Split(dtuData.SensorIdAndName, "|")
		comDTUDatas = append(comDTUDatas, model.DTUData{
			SensorId:      idName[0],
			SensorName:    idName[1],
			DTUDataDetail: comDtuDataDetail,
		})
	}
	DTU.DTUData = comDTUDatas
	DTU.Id, DTU.Payload, DTU.UpdateTime, DTU.CreateTime = primitive.ObjectID{}, "", "", endTime
	storeData(DTU, interval)
}
func max(vals ...string) string {
	var maxValue float64
	var nullNum int
	for _, val := range vals {
		if val != "" {
			if v, err := strconv.ParseFloat(val, 32); err == nil {
				if v >= maxValue || maxValue == 0 {
					maxValue = v
				}
			}
		} else {
			nullNum++
		}
	}
	if nullNum == len(vals) {
		return ""
	}
	return fmt.Sprintf("%.3f", maxValue)
}
func min(vals ...string) string {
	var minValue float64
	var nullNum int
	for _, val := range vals {
		if val != "" {
			if v, err := strconv.ParseFloat(val, 32); err == nil {
				if v <= minValue || minValue == 0 {
					minValue = v
				}
			}
		} else {
			nullNum++
		}
	}
	if nullNum == len(vals) {
		return ""
	}
	return fmt.Sprintf("%.3f", minValue)
}
func avg(vals ...string) string {
	var sum float64
	var length float64
	var nullNum int
	for _, val := range vals {
		if val != "" {
			if v, err := strconv.ParseFloat(val, 32); err == nil {
				length += 1
				sum += v
			}
		} else {
			nullNum++
		}
	}
	if nullNum == len(vals) {
		return ""
	}
	return fmt.Sprintf("%.3f", sum/length)
}
func storeData(DTU model.DTU, interval time.Duration) {
	var collectionHis *mongo.Collection
	var s string
	switch interval {
	case time.Minute * 10:
		collectionHis = global.DTUTenMinDataColl
		s = "10分钟数据存储成功"
	case time.Minute * 30:
		collectionHis = global.DTUThirtyMinDataColl
		s = "30分钟数据存储成功"
	case time.Hour:
		collectionHis = global.DTUHourDataColl
		s = "60分钟数据存储成功"
	default:
		return
	}
	if _, err := collectionHis.InsertOne(context.TODO(), DTU); err != nil {
		fmt.Println("数据存储出错", err.Error())
		return
	}
	fmt.Println("Dtu" + s)
}
