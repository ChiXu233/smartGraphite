package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"SmartGraphite-server/utils"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func E2roastTempReportNew(data model.SensorData) {
	var RendTime, RstartTime, rstartTime string //报表结束时间取值,开始时间取值；对应晾料后最大值时间，混捏结束最大值时间,若开始时间没找到，数据错误，则将晾料锅最大值记录为开始时间
	CreateTime := data.CreateTime
	stime, _ := time.Parse("2006-01-02 15:04:05", CreateTime)
	//直接取13分钟数据，包含t2后5条数据，同时取t2最大值，t2结束温度
	sTime := stime.Add(-8 * time.Minute).Format("2006-01-02 15:04:05")
	filter := bson.M{
		"createTime": bson.M{
			"$gte": sTime,
			"$lte": CreateTime,
		},
	}
	var limit int64 = 9
	opt := options.FindOptions{
		Limit: &limit,
		Sort:  bson.M{"_id": -1},
	}
	res, err := global.E2RoastTempMin.Find(context.TODO(), filter, &opt)
	if err != nil {
		fmt.Println("E2晾料报表", err.Error())
		return
	}
	var dateDB []model.SensorData
	err = res.All(context.TODO(), &dateDB)
	if err != nil {
		fmt.Println("E2晾料报表", err.Error())
		return
	}
	var t2max string //修改为混捏结束时间的晾料锅温度
	for i := 0; i < len(dateDB)-4; i++ {
		//if stringCompare(dateDB[i].DTUDataDetail[1].Value, t2max) {
		//	t2max = dateDB[i].DTUDataDetail[1].Value
		//}
		if stringCompare(dateDB[i].DTUDataDetail[1].Value, "70") && stringCompare("70", dateDB[i+1].DTUDataDetail[1].Value) && stringCompare("70", dateDB[i+2].DTUDataDetail[1].Value) && stringCompare("70", dateDB[i+3].DTUDataDetail[1].Value) && stringCompare("70", dateDB[i+4].DTUDataDetail[1].Value) {
			RendTime = dateDB[i].CreateTime
			if stringCompare(dateDB[i].DTUDataDetail[0].Value, dateDB[i+1].DTUDataDetail[0].Value) {
				t2max = dateDB[i].DTUDataDetail[0].Value
			} else {
				t2max = dateDB[i+1].DTUDataDetail[0].Value
			}
			break
		}
	}
	if RendTime == "" {
		fmt.Println("当前暂无报表结束时间")
		return
	}
	startTime, _ := time.Parse("2006-01-02 15:04:05", RendTime)
	stTime := startTime.Add(-70 * time.Minute).Format("2006-01-02 15:04:05")
	filters := bson.M{
		"createTime": bson.M{
			"$gte": stTime,
			"$lte": RendTime,
		},
	}
	var limits int64 = 70
	opts := options.FindOptions{
		Sort:  bson.M{"_id": -1},
		Limit: &limits,
	}
	//去分钟表拿到从报表结束时间往前60分钟的数据，找到开始时间
	result, err := global.E2RoastTempMin.Find(context.TODO(), filters, &opts)
	if err != nil {
		fmt.Println("E2晾料报表查找出错", err.Error())
		return
	}
	var dataDB []model.SensorData
	err = result.All(context.TODO(), &dataDB)
	if err != nil {
		fmt.Println("E2晾料报表", err.Error())
		return
	}
	//keyMap := make(map[string]int)
	//for i := 0; i < 1; i++ {
	//	for key, value := range dataDB[i].DTUDataDetail {
	//		keyMap[value.Key] = key
	//	}
	//}
	maxMap := make(map[string]model.DTUDataDetail)
	for j := 0; j < len(dataDB)-4; j++ {
		if j == 0 {
			maxMap["晾料后"] = model.DTUDataDetail{
				Key:   "t2",
				Value: t2max,
				Unit:  "℃",
			}
			maxMap["混捏结束温度"] = dataDB[j].DTUDataDetail[17]
			maxMap["晾料中"] = dataDB[j].DTUDataDetail[0]
		}
		if stringCompare(dataDB[j].DTUDataDetail[17].Value, maxMap["混捏结束温度"].Value) {
			maxMap["混捏结束温度"] = dataDB[j].DTUDataDetail[17]
		}
		if stringCompare(dataDB[j].DTUDataDetail[0].Value, maxMap["晾料中"].Value) {
			maxMap["晾料中"] = dataDB[j].DTUDataDetail[0]
			rstartTime = dataDB[j].CreateTime
		}
		if stringCompare(dataDB[j].DTUDataDetail[17].Value, "70") && stringCompare("70", dataDB[j+1].DTUDataDetail[17].Value) && stringCompare("70", dataDB[j+2].DTUDataDetail[17].Value) && stringCompare("70", dataDB[j+3].DTUDataDetail[17].Value) && stringCompare("70", dataDB[j+4].DTUDataDetail[17].Value) {
			RstartTime = dataDB[j].CreateTime
			break
		}
	}

	//数据存储
	var report model.RoastTempTMReportNew
	report.Code = time.Now().Format("20060102150405")
	report.Name = "成型温度报表"
	report.StartTime = RstartTime
	if report.StartTime == "" {
		report.StartTime = rstartTime
	}
	report.EndTime = RendTime
	report.CreateTime = utils.TimeFormat(time.Now())
	report.MaxMap = maxMap
	err = global.E2RoastTempReportNew.FindOneAndUpdate(context.TODO(), bson.M{"endTime": report.EndTime}, bson.M{"$set": report}).Decode(&bson.M{})
	if err != nil {
		_, err = global.E2RoastTempReportNew.InsertOne(context.TODO(), report)
		if err != nil {
			fmt.Println("E2RoastTempTMReport，数据存储失败", err)
		}
	}
}

// 计算报表生成后十分钟内t2的最大值
func E2RoastTempGetT2Max() {
	nowTime := time.Now().Format("2006-01-02 15:04") + ":00"
	filter := bson.M{
		"createTime": bson.M{
			"$lte": nowTime,
		},
	}
	opt := options.FindOneOptions{
		Sort: bson.M{"_id": -1},
	}
	var report model.RoastTempTMReportNew
	if err := global.E2RoastTempReport.FindOne(context.TODO(), filter, &opt).Decode(&report); err != nil {
		fmt.Println("E2RoastTempReport查找最新一条报表时间失败", err.Error())
		return
	}
	sTime, _ := time.Parse("2006-01-02 15:04:05", report.EndTime)
	nTime, _ := time.Parse("2006-01-02 15:04:05", nowTime)
	SubTIme := fmt.Sprintf("%.0f", nTime.Sub(sTime).Minutes())
	if SubTIme == "5" {
		var dataDB []model.SensorData
		filters := bson.M{
			"createTime": bson.M{
				"$lte": nowTime,
				"$gte": report.EndTime,
			},
		}
		res, err := global.E2RoastTempMin.Find(context.TODO(), filters)
		if err != nil {
			fmt.Println("计算t2最大值出错", err.Error())
		}
		err = res.All(context.TODO(), &dataDB)
		if err != nil {
			fmt.Println("计算t2最大值出错", err.Error())
		}

		//定义最大值接收t2
		var max string
		for i := range dataDB {
			if i == 0 {
				//初始化
				max = dataDB[i].DTUDataDetail[1].Value
			} else {
				if stringCompare(dataDB[i].DTUDataDetail[1].Value, max) {
					max = dataDB[i].DTUDataDetail[1].Value
				}
			}
		}
		report.MaxMap["晾料后"] = model.DTUDataDetail{
			Key:   "t2",
			Value: max,
			Unit:  "℃",
		}
		err = global.E2RoastTempReportNew.FindOneAndUpdate(context.TODO(), bson.M{"createTime": report.CreateTime}, bson.M{"$set": report}).Decode(&bson.M{})
		if err != nil {
			_, err = global.E2RoastTempReportNew.InsertOne(context.TODO(), report)
			if err != nil {
				fmt.Println("成型混捏&晾料生产工艺，按照指定规则生成分钟值,数据存储更新失败", err)
			}
		}
	} else {
		fmt.Println("现在时间暂无t2最大值")
		return
	}
}
