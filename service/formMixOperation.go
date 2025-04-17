package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"SmartGraphite-server/utils"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"time"
)

//成型混捏&晾料生产工艺

// E2RoastDataFilter 按照指定规则生成分钟值
func E2RoastDataFilter() {
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

	res, err := global.CollMap["e21"][1].Find(context.TODO(), filter, &opts)
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

	//生成T5
	data.DTUDataDetail = append(dataDB[0].DTUDataDetail,
		model.DTUDataDetail{
			Key:   "t5",
			Value: stringMax(dataDB[0].DTUDataDetail[2].Value, dataDB[0].DTUDataDetail[3].Value),
			Unit:  "℃",
		},
	)

	//数据存储
	createTime := dataDB[len(dataDB)-1].CreateTime[:]
	data.CreateTime = createTime[:len(createTime)-2] + "00"
	E2roastTempReportNew(data)

	var E4data model.SensorData
	//添加沥青温度
	err = global.CollMap["e41"][101].FindOne(context.TODO(), bson.M{"createTime": data.CreateTime}).Decode(&E4data)
	if E4data.DTUDataDetail == nil {
		data.DTUDataDetail = append(data.DTUDataDetail, model.DTUDataDetail{
			Key:   "t1lq",
			Value: "0",
			Unit:  "℃",
		})
		data.DTUDataDetail = append(data.DTUDataDetail, model.DTUDataDetail{
			Key:   "t2lq",
			Value: "0",
			Unit:  "℃",
		})
	} else {
		for i := 0; i < 2; i++ {
			data.DTUDataDetail = append(data.DTUDataDetail, model.DTUDataDetail{
				Key:   E4data.DTUDataDetail[i].Key + "lq",
				Value: E4data.DTUDataDetail[i].Value,
				Unit:  E4data.DTUDataDetail[i].Unit,
			})
		}
	}

	err = global.CollMap["e21"][101].FindOneAndUpdate(context.TODO(), bson.M{"createTime": data.CreateTime}, bson.M{"$set": data}).Decode(&bson.M{})
	if err != nil {
		_, err = global.CollMap["e21"][101].InsertOne(context.TODO(), data)
		if err != nil {
			fmt.Println("成型混捏&晾料生产工艺，按照指定规则生成分钟值,数据存储更新失败", err)
		}
	}
}

// E4分钟值生成
// E4RoastDataFilter 按照指定规则生成分钟值
func E4RoastDataFilter() {
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

	res, err := global.CollMap["e41"][1].Find(context.TODO(), filter, &opts)
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

	//生成T5
	//data.DTUDataDetail = append(dataDB[0].DTUDataDetail,
	//	model.DTUDataDetail{
	//		Key:   "t5",
	//		Value: stringMax(dataDB[0].DTUDataDetail[2].Value, dataDB[0].DTUDataDetail[3].Value),
	//		Unit:  "℃",
	//	},
	//)

	//数据存储
	createTime := dataDB[len(dataDB)-1].CreateTime[:]
	data.CreateTime = createTime[:len(createTime)-2] + "00"
	//E2roastTempReportNew(data)
	err = global.CollMap["e41"][101].FindOneAndUpdate(context.TODO(), bson.M{"createTime": data.CreateTime}, bson.M{"$set": data}).Decode(&bson.M{})
	if err != nil {
		_, err = global.CollMap["e41"][101].InsertOne(context.TODO(), data)
		if err != nil {
			fmt.Println("成型混捏&晾料生产工艺，按照指定规则生成分钟值,数据存储更新失败", err)
		}
	}

}

// E2RoastGetTM 从分钟表，计算得到TM1-TM4
func E2RoastGetTM() {
	endTime := time.Now().Format("2006-01-02 15:04") + ":00"
	startTime := ""

	//从TM表中找到最新的一条记录
	var tmDB model.SensorData
	err := global.E2RoastTempTM.FindOne(context.TODO(), bson.M{}, &options.FindOneOptions{Sort: bson.M{"createTime": -1}}).Decode(&tmDB)
	if err != nil { //没有数据,则去分钟表中拿到最开始的一条记录
		var roastDB model.SensorData
		err = global.E2RoastTempMin.FindOne(context.TODO(), bson.M{}).Decode(&roastDB)
		if err != nil { //分钟表没有数据
			return
		}

		startTime = roastDB.CreateTime
	} else { //有数据

		startTime = tmDB.CreateTime
	}
	filter := bson.M{
		"createTime": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}

	//获取limit
	limit, err := utils.GetLimit(startTime, endTime, "1", "2006-01-02 15:04:05")
	if err != nil {
		fmt.Println("E2RoastGetTM 从分钟表计算得到TM1-TM4", err)
		return
	}
	opts := options.FindOptions{
		Limit: &limit,
	}

	res, err := global.CollMap["e21"][101].Find(context.TODO(), filter, &opts)
	if err != nil {
		fmt.Println("E2RoastGetTM 从分钟表计算得到TM1-TM4", err)
	}

	var dataDB []model.SensorData
	if err = res.All(context.TODO(), &dataDB); err != nil {
		fmt.Println("E2RoastGetTM 从分钟表计算得到TM1-TM4", err)
		return
	}

	if len(dataDB) < 5 {
		fmt.Println("E2RoastGetTM 从分钟表计算得到TM1-TM4,此时间段没有TM值")
		return
	}

	//拿到key字典
	keyMap := make(map[string]int)
	for i := 0; i < 1; i++ {
		for key, value := range dataDB[i].DTUDataDetail {
			keyMap[value.Key] = key
		}
	}
	data := make(map[string][]model.SensorData)
	flag := true
	//数据处理
	for i := 5; i < len(dataDB); i++ {
		flag = true
		//TM1
		if stringCompare(dataDB[i].DTUDataDetail[keyMap["t5"]].Value, "100") {
			for j := i - 1; j >= i-5; j-- {
				if stringCompare(dataDB[j].DTUDataDetail[keyMap["t5"]].Value, dataDB[i].DTUDataDetail[keyMap["t5"]].Value) || stringCompare(dataDB[j].DTUDataDetail[keyMap["t5"]].Value, "100") {
					flag = false
				}
			}
			if flag {
				data["TM1"] = append(data["TM1"], model.SensorData{
					Code:          "TM1",
					DTUDataDetail: dataDB[i].DTUDataDetail,
					CreateTime:    dataDB[i].CreateTime,
				})
			}
		}

	}

	section := 0
	//数据存储
	for key := range data {
		if section == 0 && key == "TM1" {
			section = len(data[key])
		}
		for _, val := range data[key] {
			err = global.E2RoastTempTM.FindOneAndUpdate(context.TODO(), bson.M{"createTime": val.CreateTime}, bson.M{"$set": val}).Decode(&bson.M{})
			if err != nil {
				_, err = global.E2RoastTempTM.InsertOne(context.TODO(), val)
				if err != nil {
					fmt.Println("E2RoastGetTM," + val.Code + "数据存储更新失败")
				}
			}
		}
	}

	//计算报表
	E2RoastTempTMReportLimit()

}

// E2RoastTempTMReportLimit 根据TM最新记录时间和报表最新的时间，计算出最多需要生成多少个报表
func E2RoastTempTMReportLimit() {

	var reportDB model.RoastTempTMReport
	tm1Filter := bson.M{}
	err := global.E2RoastTempReport.FindOne(context.TODO(), bson.M{}, &options.FindOneOptions{Sort: bson.M{"_id": -1}}).Decode(&reportDB)
	if err != nil {
		//第一次初始化值
		fmt.Println("E2RoastTempTMReportLimit", "第一次初始化值")
	} else {
		tm1Filter["createTime"] = bson.M{"$gt": reportDB.StartTime}
	}

	//获取待计算的TM条数
	limit, err := global.E2RoastTempTM.CountDocuments(context.TODO(), tm1Filter)
	if err != nil {
		fmt.Println("E2RoastTempTMReportLimit", err)
		return
	}

	lim := int(limit)
	for i := 0; i < lim; i++ {
		E2RoastTempTMReport()
	}

}

// E2RoastTempTMReport 计算TM报表
func E2RoastTempTMReport() {

	var reportDB model.RoastTempTMReport
	tm1Filter := bson.M{}
	err := global.E2RoastTempReport.FindOne(context.TODO(), bson.M{}, &options.FindOneOptions{Sort: bson.M{"_id": -1}}).Decode(&reportDB)
	if err != nil {
		//第一次初始化值
		fmt.Println("E2RoastTempTMReport", "第一次初始化值")
	} else {
		tm1Filter["createTime"] = bson.M{"$gt": reportDB.StartTime}
	}
	tm1Filter["sensorId"] = "TM1"

	//查询出最新一条TM1作为报表开始时间
	var tm1DB model.SensorData
	err = global.E2RoastTempTM.FindOne(context.TODO(), tm1Filter).Decode(&tm1DB)
	if err != nil {
		fmt.Println("tm1DB", err)
		return
	}

	//eTime, err := time.Parse("2006-01-02 15:04:05", tm1DB.CreateTime)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//endTime := eTime.Add(time.Minute * 70).Format("2006-01-02 15:04:05")
	//filter := bson.M{
	//	"createTime": bson.M{
	//		"$gte": tm1DB.CreateTime,
	//		"$lte": endTime,
	//	},
	//}
	//
	//res, err := global.E2RoastTempMin.Find(context.TODO(), filter)
	//if err != nil {
	//	fmt.Println("E2RoastTempTMReport", err)
	//	return
	//}
	//
	//var dataDB []model.SensorData
	//if err = res.All(context.TODO(), &dataDB); err != nil {
	//	fmt.Println("E2RoastTempTMReport", err)
	//	return
	//}
	//
	//if dataDB == nil {
	//	fmt.Println("E2RoastTempTMReport，此时间段没有数据")
	//}
	//
	//if len(dataDB) < 70 { //还没有近一个小时的数据
	//	fmt.Println("E2RoastTempTMReport,此时间段数据量过少")
	//	return
	//}
	//
	////拿到key字典
	//keyMap := make(map[string]int)
	//for i := 0; i < 1; i++ {
	//	for key, value := range dataDB[i].DTUDataDetail {
	//		keyMap[value.Key] = key
	//	}
	//}
	//
	//var report model.RoastTempTMReport
	////数据处理
	//maxMap := make(map[string]model.SensorData)
	////maxMap["混捏结束温度"]//20
	////maxMap["晾料开始温度"]//20
	////maxMap["晾料结束温度"]//60
	//for i := range dataDB {
	//
	//	//chixu，前20分钟，混捏温度最大值，晾料温度最大值
	//	if i < 20 {
	//		if i == 0 { //初始化
	//			if stringCompare(dataDB[i].DTUDataDetail[keyMap["t5"]].Value, "500") || stringCompare(dataDB[i].DTUDataDetail[keyMap["t1"]].Value, "500") || stringCompare(dataDB[i].DTUDataDetail[keyMap["t2"]].Value, "500") {
	//				maxMap["混捏结束温度"] = dataDB[i+1]
	//				maxMap["晾料开始温度"] = dataDB[i+1]
	//				maxMap["晾料结束温度"] = dataDB[i+1]
	//			} else {
	//				maxMap["混捏结束温度"] = dataDB[i]
	//				maxMap["晾料开始温度"] = dataDB[i]
	//				maxMap["晾料结束温度"] = dataDB[i]
	//			}
	//		} else {
	//
	//			if stringCompare(dataDB[i].DTUDataDetail[keyMap["t5"]].Value, maxMap["混捏结束温度"].DTUDataDetail[keyMap["t5"]].Value) && stringCompare("500", dataDB[i].DTUDataDetail[keyMap["t5"]].Value) {
	//				maxMap["混捏结束温度"] = dataDB[i]
	//			}
	//			if stringCompare(dataDB[i].DTUDataDetail[keyMap["t1"]].Value, maxMap["晾料开始温度"].DTUDataDetail[keyMap["t1"]].Value) && stringCompare("500", dataDB[i].DTUDataDetail[keyMap["t1"]].Value) {
	//				maxMap["晾料开始温度"] = dataDB[i]
	//			}
	//		}
	//	}
	//	//60分钟晾料结束温度
	//	if stringCompare(dataDB[i].DTUDataDetail[keyMap["t2"]].Value, maxMap["晾料结束温度"].DTUDataDetail[keyMap["t2"]].Value) && stringCompare("500", dataDB[i].DTUDataDetail[keyMap["t2"]].Value) {
	//		maxMap["晾料结束温度"] = dataDB[i]
	//	}
	//
	//	//前数据处理
	//	if i == 0 {
	//
	//		report.Code = time.Now().Format("20060102150405")
	//		report.Name = "成型温度报表"
	//		report.StartTime = tm1DB.CreateTime
	//		report.EndTime = dataDB[len(dataDB)-1].CreateTime
	//		report.CreateTime = utils.TimeFormat(time.Now())
	//
	//		var data model.DataDetail
	//		data.Info = make(map[string]model.InfoDataDetail)
	//		for key, val := range dataDB[i].DTUDataDetail {
	//			data.Info[val.Key] = model.InfoDataDetail{
	//				Data: []string{dataDB[i].DTUDataDetail[key].Value},
	//				Time: []string{dataDB[i].CreateTime},
	//				Unit: "℃",
	//			}
	//		}
	//
	//		report.Data = append(report.Data, data)
	//	} else {
	//		for _, val := range dataDB[i].DTUDataDetail {
	//
	//			//使用值修改的方式修改map
	//			infoDetail := report.Data[0].Info[val.Key]
	//			infoDetail.Data = append(infoDetail.Data, val.Value)
	//			infoDetail.Time = append(infoDetail.Time, dataDB[i].CreateTime)
	//			report.Data[0].Info[val.Key] = infoDetail
	//
	//		}
	//	}
	//}
	//
	//report.MaxMap = maxMap
	//var TM1 model.SensorData
	//if err = global.E2RoastTempTM.FindOne(context.TODO(), bson.M{}, &options.FindOneOptions{Sort: bson.M{"createTime": -1}}).Decode(&TM1); err != nil {
	//	return
	//}
	//TMtime := TM1.CreateTime

	var limit int64 = 46
	opts := options.FindOptions{
		Limit: &limit,
	}
	filters := bson.M{
		"createTime": bson.M{
			//time要为string类型
			"$gte": tm1DB.CreateTime,
		},
	}
	res, err := global.E2RoastTempMin.Find(context.TODO(), filters, &opts)
	if err != nil {
		fmt.Println("获取前70条数据失败", err.Error())
	}
	var data []model.SensorData
	err = res.All(context.TODO(), &data)
	if err != nil {
		fmt.Println("获取前70条数据失败", err.Error())
	}
	if data == nil {
		fmt.Println("E2RoastTempTMReport，此时间段没有数据")
	}
	if len(data) < 45 {
		fmt.Println("E2RoastTempTMReport,此时间段数据量过少")
		return
	}

	//拿到key字典
	keyMap := make(map[string]int)
	for i := 0; i < 1; i++ {
		for key, value := range data[i].DTUDataDetail {
			keyMap[value.Key] = key
		}
	}

	var report model.RoastTempTMReport
	//数据处理
	maxMap := make(map[string]model.SensorData)
	for i := range data {
		//chixu，前20分钟，混捏温度最大值，晾料温度最大值
		if i < 20 {
			if i == 0 { //初始化
				if stringCompare(data[i].DTUDataDetail[keyMap["t5"]].Value, "500") || stringCompare(data[i].DTUDataDetail[keyMap["t1"]].Value, "500") || stringCompare(data[i].DTUDataDetail[keyMap["t2"]].Value, "500") {
					maxMap["混捏结束温度"] = data[i+1]
					maxMap["晾料开始温度"] = data[i+1]
					maxMap["晾料结束温度"] = data[i+1]
				} else {
					maxMap["混捏结束温度"] = data[i]
					maxMap["晾料开始温度"] = data[i]
					maxMap["晾料结束温度"] = data[i]
				}
			} else {

				if stringCompare(data[i].DTUDataDetail[keyMap["t5"]].Value, maxMap["混捏结束温度"].DTUDataDetail[keyMap["t5"]].Value) && stringCompare("500", data[i].DTUDataDetail[keyMap["t5"]].Value) {
					maxMap["混捏结束温度"] = data[i]
				}
				if stringCompare(data[i].DTUDataDetail[keyMap["t1"]].Value, maxMap["晾料开始温度"].DTUDataDetail[keyMap["t1"]].Value) && stringCompare("500", data[i].DTUDataDetail[keyMap["t1"]].Value) {
					maxMap["晾料开始温度"] = data[i]
				}
			}
		}
		//60分钟晾料结束温度
		if stringCompare(data[i].DTUDataDetail[keyMap["t2"]].Value, maxMap["晾料结束温度"].DTUDataDetail[keyMap["t2"]].Value) && stringCompare("500", data[i].DTUDataDetail[keyMap["t2"]].Value) {
			maxMap["晾料结束温度"] = data[i]
		}

		if i == 0 {
			//初始化
			report.Code = time.Now().Format("20060102150405")
			report.Name = "成型温度报表"
			report.StartTime = tm1DB.CreateTime
			report.EndTime = data[len(data)-1].CreateTime
			report.CreateTime = utils.TimeFormat(time.Now())
		}
		//	var date model.DataDetail
		//	date.Info = make(map[string]model.InfoDataDetail)
		//	for key, val := range data[i].DTUDataDetail {
		//		date.Info[val.Key] = model.InfoDataDetail{
		//			Data: []string{data[i].DTUDataDetail[key].Value},
		//			Time: []string{data[i].CreateTime},
		//			Unit: "℃",
		//		}
		//	}
		//	report.Data = append(report.Data, date)
		//} else {
		//for _, val := range data[i].DTUDataDetail {

		//使用值修改的方式修改map
		//infoDetail := report.Data[0].Info[val.Key]
		//infoDetail.Data = append(infoDetail.Data, val.Value)
		//infoDetail.Time = append(infoDetail.Time, data[i].CreateTime)
		//report.Data[0].Info[val.Key] = infoDetail

		//}

	}
	report.MaxMap = maxMap
	err = global.E2RoastTempReport.FindOneAndUpdate(context.TODO(), bson.M{"createTime": report.StartTime}, bson.M{"$set": report}).Decode(&bson.M{})
	if err != nil {
		_, err = global.E2RoastTempReport.InsertOne(context.TODO(), report)
		if err != nil {
			fmt.Println("E2RoastTempTMReport，数据存储失败", err)
		}
	}

}

// E2RoastTempDelete 只保留一个小时的原始数据
func E2RoastTempDelete() {

	var dataDB model.SensorData
	if err := global.E2RoastTemp.FindOne(context.TODO(), bson.M{}, &options.FindOneOptions{Sort: bson.M{"_id": -1}}).Decode(&dataDB); err != nil {
		fmt.Println("E2RoastTemp数据删除", err)
		return
	}

	STime, err := time.Parse("2006-01-02 15:04:05", dataDB.CreateTime)
	if err != nil {
		fmt.Println("E2RoastTemp数据删除", err)
		return
	}

	startTime := STime.Add(-24*time.Hour).Format("2006-01-02 15:04") + ":00"
	filter := bson.M{
		"createTime": bson.M{
			"$lt": startTime,
		},
	}

	_, err = global.E2RoastTemp.DeleteMany(context.TODO(), filter)
	if err != nil {
		fmt.Println(err)
		return
	}

}

// 分钟数据保留一周
func E2RoastTempMinDelete() {

	var dataDB model.SensorData
	if err := global.E2RoastTempMin.FindOne(context.TODO(), bson.M{}, &options.FindOneOptions{Sort: bson.M{"_id": -1}}).Decode(&dataDB); err != nil {
		fmt.Println("E2RoastTemp数据删除", err)
		return
	}

	STime, err := time.Parse("2006-01-02 15:04:05", dataDB.CreateTime)
	if err != nil {
		fmt.Println("E2RoastTempMin数据删除", err)
		return
	}

	startTime := STime.Add(-168*time.Hour).Format("2006-01-02 15:04") + ":00"
	filter := bson.M{
		"createTime": bson.M{
			"$lt": startTime,
		},
	}

	_, err = global.E2RoastTempMin.DeleteMany(context.TODO(), filter)
	if err != nil {
		fmt.Println(err)
		return
	}

}

// 转成10进制后比较
func stringCompare(x, y string) bool {

	xf, _ := strconv.ParseFloat(x, 10)
	yf, _ := strconv.ParseFloat(y, 10)

	if xf >= yf {
		return true
	}

	return false
}
