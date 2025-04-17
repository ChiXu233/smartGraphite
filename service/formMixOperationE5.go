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

// E5设备分钟值生成
func E5RoastDataFilter() {
	endTime := time.Now().Format("2006-01-02 15:04") + ":00"
	startTime := time.Now().Add(-1*time.Minute).Format("2006-01-02 15:04") + ":00"

	filter := bson.M{
		"createTime": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}

	var limit int64 = 11
	opts := options.FindOptions{
		Limit: &limit,
	}

	res, err := global.CollMap["e51"][1].Find(context.TODO(), filter, &opts)
	if err != nil {
		fmt.Println("成型混捏&晾料生产工艺，按照指定规则生成分钟值", err)
		return
	}

	var dataDB []model.SensorData
	if err = res.All(context.TODO(), &dataDB); err != nil {
		fmt.Println("成型混捏&晾料生产工艺，按照指定规则生成分钟值", err)
		return
	}

	if dataDB == nil {
		fmt.Println("成型混捏&晾料生产工艺，此时间段没有数据")
		return
	}

	//数据处理
	//使用枚举法获取最大值

	//找到第一个枚举值
	var data model.SensorData
	i := 0
	var flag bool
	for j := range dataDB {
		flag = true
		for val := range dataDB[j].DTUDataDetail {
			//if stringCompare(val.Value, "500") { //如果有大于1000的值则直接舍弃这条数据
			//	flag = false
			//	break
			//}
			if stringCompare(dataDB[j].DTUDataDetail[val].Value, "1000") { //如果有大于1000的值则将值赋为-999
				fmt.Println(dataDB[j].DTUDataDetail[val].Value)
				dataDB[j].DTUDataDetail[val].Value = "-999.000"
				flag = true
			}
		}

		//找到第一个枚举值
		if flag {
			i = j
			data = dataDB[j]
			break
		}
	}

	if !flag { //所有数据都超标
		return
	}

	//初始赋值
	data = dataDB[i]
	i++

	//拿到最大值
	for ; i < len(dataDB); i++ {
		flag := true
		for k := range dataDB[i].DTUDataDetail {
			if stringCompare(dataDB[i].DTUDataDetail[k].Value, "500") {
				flag = false
				break
			}
		}

		if flag { //有效值
			for k := range dataDB[i].DTUDataDetail {
				data.DTUDataDetail[k].Value = stringMax(data.DTUDataDetail[k].Value, dataDB[i].DTUDataDetail[k].Value)
			}
		}
	}

	//生成总混捏
	data.DTUDataDetail = append(dataDB[0].DTUDataDetail,
		model.DTUDataDetail{
			Key:   "tzh",
			Value: stringMax(dataDB[0].DTUDataDetail[2].Value, dataDB[0].DTUDataDetail[3].Value),
			Unit:  "℃",
		},
	)

	//数据存储
	createTime := dataDB[len(dataDB)-1].CreateTime[:]
	data.CreateTime = createTime[:len(createTime)-2] + "00"
	E5roastTempReportNew(data)
	//err = global.CollMap["e51"][101].FindOneAndUpdate(context.TODO(), bson.M{"createTime": data.CreateTime}, bson.M{"$set": data}).Decode(&bson.M{})
	//if err != nil {
	//	_, err = global.CollMap["e51"][101].InsertOne(context.TODO(), data)
	//	if err != nil {
	//		fmt.Println("成型混捏&晾料生产工艺，按照指定规则生成分钟值,数据存储更新失败", err)
	//	}
	//}

}

func E5roastTempReportNew(data model.SensorData) {
	var RendTime, RstartTime, rstartTime string //报表结束时间取值,开始时间取值；对应晾料后最大值时间，混捏结束最大值时间,若开始时间没找到，数据错误，则将晾料锅最大值记录为开始时间
	CreateTime := data.CreateTime
	stime, _ := time.Parse("2006-01-02 15:04:05", CreateTime)
	//直接取13分钟数据，取晾料皮带t2的值
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
	res, err := global.E5RoastTempMin.Find(context.TODO(), filter, &opt)
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
	fmt.Println(dateDB[0].DTUDataDetail[0])
	var t2max string //修改为混捏结束时间的晾料锅温度,取两条中的最大值
	for i := 0; i < len(dateDB)-4; i++ {
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
	result, err := global.E5RoastTempMin.Find(context.TODO(), filters, &opts)
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
	maxMap := make(map[string]model.DTUDataDetail)
	for j := 0; j < len(dataDB)-4; j++ {
		if j == 0 {
			maxMap["晾料后"] = model.DTUDataDetail{
				Key:   "t2",
				Value: t2max,
				Unit:  "℃",
			}
			maxMap["混捏结束温度"] = dataDB[j].DTUDataDetail[16]
			maxMap["晾料中"] = dataDB[j].DTUDataDetail[0]
		}
		if stringCompare(dataDB[j].DTUDataDetail[16].Value, maxMap["混捏结束温度"].Value) {
			maxMap["混捏结束温度"] = dataDB[j].DTUDataDetail[16]
		}
		if stringCompare(dataDB[j].DTUDataDetail[0].Value, maxMap["晾料中"].Value) {
			maxMap["晾料中"] = dataDB[j].DTUDataDetail[0]
			rstartTime = dataDB[j].CreateTime
		}
		if stringCompare(dataDB[j].DTUDataDetail[16].Value, "70") && stringCompare("70", dataDB[j+1].DTUDataDetail[16].Value) && stringCompare("70", dataDB[j+2].DTUDataDetail[16].Value) && stringCompare("70", dataDB[j+3].DTUDataDetail[16].Value) && stringCompare("70", dataDB[j+4].DTUDataDetail[16].Value) {
			RstartTime = dataDB[j].CreateTime
			break
		}
	}

	//数据存储
	var report model.RoastTempTMReportNew
	report.Code = time.Now().Format("20060102150405")
	report.Name = "小成型温度报表"
	report.StartTime = RstartTime
	if report.StartTime == "" {
		report.StartTime = rstartTime
	}
	report.EndTime = RendTime
	report.CreateTime = utils.TimeFormat(time.Now())
	report.MaxMap = maxMap
	//err = global.E5RoastTempReport.FindOneAndUpdate(context.TODO(), bson.M{"endTime": report.EndTime}, bson.M{"$set": report}).Decode(&bson.M{})
	//if err != nil {
	//	_, err = global.E5RoastTempReport.InsertOne(context.TODO(), report)
	//	if err != nil {
	//		fmt.Println("E2RoastTempTMReport，数据存储失败", err)
	//	}
	//}
}
