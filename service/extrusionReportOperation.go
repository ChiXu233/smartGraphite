package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"SmartGraphite-server/utils"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math"
	"strconv"
	"time"
)

// 成型温度 挤压报表
func E2extrusionTime(DTUdata []model.SensorData) {
	//获取当前数据的前一条数据
	var beforOnedata model.SensorData
	var startTime, endTime time.Time
	var totalTime time.Duration

	dataTime := DTUdata[0].CreateTime
	//"2023-09-01 09:20:06"
	filter := bson.M{
		"createTime": bson.M{
			"$lt": dataTime,
		},
	}
	opts := options.FindOneOptions{
		Sort: bson.M{"_id": -1},
	}

	if err := global.CollMap["e21"][1].FindOne(context.TODO(), filter, &opts).Decode(&beforOnedata); err != nil {
		fmt.Println("E2挤压报表查找前一条数据出错", err.Error())
		return
	}
	//k1压力与前一条数据进行比较，判断是否为记录
	//if stringCompare(DTUdata[0].DTUDataDetail[8].Value, "6") && stringCompare("6", befordata.DTUDataDetail[8].Value) {
	//	//记录当前时间为报表开始时间
	//	startT, err := time.Parse("2006-01-02 15:04:05", dataTime)
	//	if err != nil {
	//		fmt.Println("E2挤压报表时间转化错误", err.Error())
	//	}
	//	startTime = startT
	//}
	if stringCompare("4.59", DTUdata[0].DTUDataDetail[8].Value) && stringCompare(beforOnedata.DTUDataDetail[8].Value, "4.59") {
		//记录当前时间为报表结束时间
		endT, err := time.Parse("2006-01-02 15:04:05", dataTime)
		if err != nil {
			fmt.Println("E2挤压报表时间转化错误", err.Error())
		}
		endTime = endT
	}

	ntime, _ := time.Parse("2006-01-02 15:04:05", dataTime)
	sTime := ntime.Add(-8 * time.Minute).Format("2006-01-02 15:04:05")
	opt := options.FindOptions{
		Sort: bson.M{"_id": -1},
	}
	filters := bson.M{
		"createTime": bson.M{
			"$lte": dataTime,
			"$gte": sTime,
		},
	}
	res, err := global.CollMap["e21"][1].Find(context.TODO(), filters, &opt)
	if err != nil {
		fmt.Println("查找startTime失败", err.Error())
		return
	}
	var dataDB []model.SensorData
	if err := res.All(context.TODO(), &dataDB); err != nil {
		fmt.Println("查找startTime失败", err)
		return
	}

	if len(dataDB) == 0 {
		fmt.Println("暂无数据")
		return
	}
	for j := 0; j < len(dataDB)-1; j++ {
		//从当前时间往前查找，获取startTime

		if j+1 < len(dataDB) {
			if stringCompare(dataDB[j].DTUDataDetail[8].Value, "4.59") && stringCompare("4.59", dataDB[j+1].DTUDataDetail[8].Value) {
				//记录当前时间为报表开始时间
				startT, err := time.Parse("2006-01-02 15:04:05", dataDB[j].CreateTime)
				if err != nil {
					fmt.Println("E2挤压报表时间转化错误", err.Error())
				}
				startTime = startT
				break
			}
		}
	}
	if startTime.Format("2006-01-02 15:04:05") == "0001-01-01 00:00:00" {
		fmt.Println("此时间为预压时间")
		return
	}

	totalTime = time.Duration(endTime.Sub(startTime).Seconds())
	E2extrusionGetData(startTime, endTime, totalTime)
}

func E2extrusionGetData(starttime time.Time, endtime time.Time, totalTime time.Duration) {
	//计算报表所需要的数据
	startTime := starttime.Format("2006-01-02 15:04:05")
	endTime := endtime.Format("2006-01-02 15:04:05")

	filter := bson.M{
		"createTime": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}

	opts := options.FindOptions{
		Sort: bson.M{"_id": -1},
	}
	res, err := global.CollMap["e21"][1].Find(context.TODO(), filter, &opts)
	if err != nil {
		fmt.Println("E2挤压报表查找实时数据出错", err)
		return
	}
	var dataDB []model.SensorData
	if err := res.All(context.TODO(), &dataDB); err != nil {
		fmt.Println("E2挤压报表出错", err)
		return
	}
	if len(dataDB) == 0 {
		fmt.Println("当前时间暂未挤压/预压")
		return
	}

	//获取k1,k2最大值，最小值，平均值，垂头l总位移，位移速度
	keyMap := make(map[string]int)
	for i := 0; i < 1; i++ {
		for key, value := range dataDB[i].DTUDataDetail {
			keyMap[value.Key] = key
		}
	}

	maxMap := make(map[string]model.DTUDataDetail)
	minMap := make(map[string]model.DTUDataDetail)
	avgMap := make(map[string]model.DTUDataDetail)

	//压力总值，抽真空总值,初位移,末位移
	//压力平均值，抽真空平均值
	var k1sum, k2sum, l0, l1 float64
	var k1avg, k2avg string

	for i := range dataDB {

		//初始化
		if i == 0 {
			maxMap["压力最大值"] = dataDB[i].DTUDataDetail[8]
			minMap["压力最小值"] = dataDB[i].DTUDataDetail[8]
			avgMap["压力平均值"] = dataDB[i].DTUDataDetail[8]
			maxMap["抽真空最大值"] = dataDB[i].DTUDataDetail[12]
			minMap["抽真空最小值"] = dataDB[i].DTUDataDetail[12]
			avgMap["抽真空平均值"] = dataDB[i].DTUDataDetail[12]
			maxMap["料室温度最大值"] = dataDB[i].DTUDataDetail[9]
			maxMap["直线区温度最大值"] = dataDB[i].DTUDataDetail[7]
			maxMap["变形区温度最大值"] = dataDB[i].DTUDataDetail[6]
			l0, err = strconv.ParseFloat(dataDB[i].DTUDataDetail[10].Value, 64)
			if err != nil {
				fmt.Println("数据转化出错", err.Error())
			}
		}
		if i == len(dataDB)-1 {
			l1, err = strconv.ParseFloat(dataDB[i].DTUDataDetail[10].Value, 64)
			if err != nil {
				fmt.Println("数据转化出错", err.Error())
			}
		}
		if stringCompare(dataDB[i].DTUDataDetail[keyMap["k1"]].Value, maxMap["压力最大值"].Value) {
			maxMap["压力最大值"] = dataDB[i].DTUDataDetail[keyMap["k1"]]
		} else {
			minMap["压力最小值"] = dataDB[i].DTUDataDetail[keyMap["k1"]]
		}
		if stringCompare(dataDB[i].DTUDataDetail[keyMap["k2"]].Value, maxMap["抽真空最大值"].Value) {
			maxMap["抽真空最大值"] = dataDB[i].DTUDataDetail[keyMap["k2"]]
		} else {
			minMap["抽真空最小值"] = dataDB[i].DTUDataDetail[keyMap["k2"]]
		}

		//取变形区,直线区,料室温度最大值
		if stringCompare(dataDB[i].DTUDataDetail[keyMap["z1"]].Value, maxMap["变形区温度最大值"].Value) {
			maxMap["变形区温度最大值"] = dataDB[i].DTUDataDetail[keyMap["z1"]]
		}
		if stringCompare(dataDB[i].DTUDataDetail[keyMap["z2"]].Value, maxMap["直线区温度最大值"].Value) {
			maxMap["直线区温度最大值"] = dataDB[i].DTUDataDetail[keyMap["z2"]]
		}
		if stringCompare(dataDB[i].DTUDataDetail[keyMap["c1"]].Value, maxMap["料室温度最大值"].Value) {
			maxMap["料室温度最大值"] = dataDB[i].DTUDataDetail[keyMap["c1"]]
		}

		//累加求平均值
		//累加和sum
		value, err := strconv.ParseFloat(dataDB[i].DTUDataDetail[8].Value, 64)
		if err != nil {
			fmt.Println("类型转化出错", err.Error())
			return
		}
		value2, err := strconv.ParseFloat(dataDB[i].DTUDataDetail[12].Value, 64)
		if err != nil {
			fmt.Println("类型转化出错", err.Error())
			return
		}
		k1sum += value
		k2sum += value2
	}

	//计算平均值
	k1avg = fmt.Sprintf("%f", float32(k1sum/float64(len(dataDB))))
	k2avg = fmt.Sprintf("%f", float32(k2sum/float64(len(dataDB))))
	avgMap["压力平均值"] = model.DTUDataDetail{
		Key:   "k1",
		Value: k1avg,
		Unit:  "",
	}
	avgMap["抽真空度平均值"] = model.DTUDataDetail{
		Key:   "k2",
		Value: k2avg,
		Unit:  "",
	}
	//计算垂头速度,l2为垂头位移绝对值
	l2 := math.Abs(l1 - l0)
	v := l2 / float64(totalTime)
	E2extrusionReport(startTime, endTime, maxMap, minMap, avgMap, l2, v)
}
func E2extrusionReport(startTime string, endTime string, maxMap map[string]model.DTUDataDetail, minMap map[string]model.DTUDataDetail, avgMap map[string]model.DTUDataDetail, l2 float64, v float64) {
	var report model.ExtrusionReport
	if stringCompare(maxMap["压力最大值"].Value, "18") {
		//压力最大值>18为预压，否则为挤压
		report.Status = "预压"
	} else {
		report.Status = "挤压"
	}
	report.Name = "挤压报表"
	report.StartTime = startTime
	report.EndTime = endTime
	report.CreateTime = utils.TimeFormat(time.Now())
	report.MaxMap = maxMap
	report.MinMap = minMap
	report.AvgMap = avgMap
	report.Speed = fmt.Sprintf("%f", v)
	report.Displace = fmt.Sprintf("%f", l2)
	report.Code = time.Now().Format("20060102150405")
	err := global.E2extrusionReport.FindOneAndUpdate(context.TODO(), bson.M{"createTime": report.StartTime}, bson.M{"$set": report}).Decode(&bson.M{})
	if err != nil {
		_, err = global.E2extrusionReport.InsertOne(context.TODO(), report)
		if err != nil {
			fmt.Println("E2extrusionReport，数据存储失败", err)
		}
	}
}

func E5extrusionTime(DTUdata []model.SensorData) {
	var beforOnedata model.SensorData
	var startTime, endTime time.Time
	var totalTime time.Duration
	dataTime := DTUdata[0].CreateTime
	filter := bson.M{
		"createTime": bson.M{
			"$lt": dataTime,
		},
	}
	opts := options.FindOneOptions{Sort: bson.M{"_id": -1}}
	if err := global.CollMap["e51"][1].FindOne(context.TODO(), filter, &opts).Decode(&beforOnedata); err != nil {
		fmt.Println("E5挤压报表查找前一条数据出错", err.Error())
		return
	}
	if stringCompare("4.59", DTUdata[0].DTUDataDetail[4].Value) && stringCompare(beforOnedata.DTUDataDetail[4].Value, "4.59") {
		//记录为报表结束时间
		endT, err := time.Parse("2006-01-02 15:04:05", dataTime)
		if err != nil {
			fmt.Println("E2挤压报表时间转化错误", err.Error())
		}
		endTime = endT
	}
	ntime, _ := time.Parse("2006-01-02 15:04:05", dataTime)
	sTime := ntime.Add(-8 * time.Minute).Format("2006-01-02 15:04:05")
	opt := options.FindOptions{
		Sort: bson.M{"_id": -1},
	}
	filters := bson.M{
		"createTime": bson.M{
			"$lte": dataTime,
			"$gte": sTime,
		},
	}
	res, err := global.CollMap["e51"][1].Find(context.TODO(), filters, &opt)
	if err != nil {
		fmt.Println("查找startTime失败", err.Error())
		return
	}
	var dataDB []model.SensorData
	if err := res.All(context.TODO(), &dataDB); err != nil {
		fmt.Println("查找startTime失败", err)
		return
	}
	if len(dataDB) == 0 {
		fmt.Println("暂无数据")
		return
	}
	for j := 0; j < len(dataDB)-1; j++ {
		if j+1 < len(dataDB) {
			if stringCompare(dataDB[j].DTUDataDetail[4].Value, "4.59") && stringCompare("4.59", dataDB[j+1].DTUDataDetail[4].Value) {
				//记录当前时间为报表开始时间
				startT, err := time.Parse("2006-01-02 15:04:05", dataDB[j].CreateTime)
				if err != nil {
					fmt.Println("E2挤压报表时间转化错误", err.Error())
				}
				startTime = startT
				break
			}
		}
	}
	if startTime.Format("2006-01-02 15:04:05") == "0001-01-01 00:00:00" {
		fmt.Println("此时间为预压时间")
		return
	}
	totalTime = time.Duration(endTime.Sub(startTime).Seconds())
	E5extrusionGetData(startTime, endTime, totalTime)
}

func E5extrusionGetData(starttime time.Time, endtime time.Time, totalTime time.Duration) {
	startTime := starttime.Format("2006-01-02 15:04:05")
	endTime := endtime.Format("2006-01-02 15:04:05")
	filter := bson.M{
		"createTime": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}
	opts := options.FindOptions{
		Sort: bson.M{"_id": -1},
	}
	res, err := global.CollMap["e51"][1].Find(context.TODO(), filter, &opts)
	if err != nil {
		fmt.Println("E5挤压报表查找实时数据出错", err)
	}
	var dataDB []model.SensorData
	if err := res.All(context.TODO(), &dataDB); err != nil {
		fmt.Println("E5挤压报表出错", err)
		return
	}
	if len(dataDB) == 0 {
		fmt.Println("当前时间暂未挤压/预压")
		return
	}
	//获取k1,k2最大值，最小值，平均值，垂头l总位移，位移速度
	keyMap := make(map[string]int)
	for i := 0; i < 1; i++ {
		for key, value := range dataDB[i].DTUDataDetail {
			keyMap[value.Key] = key
		}
	}

	maxMap := make(map[string]model.DTUDataDetail)
	minMap := make(map[string]model.DTUDataDetail)
	avgMap := make(map[string]model.DTUDataDetail)

	//压力总值，抽真空总值,初位移,末位移
	//压力平均值，抽真空平均值
	var k1sum, k2sum float64
	var k1avg, k2avg string
	for i := range dataDB {
		if i == 0 {
			maxMap["压力最大值"] = dataDB[i].DTUDataDetail[4]
			minMap["压力最小值"] = dataDB[i].DTUDataDetail[4]
			avgMap["压力平均值"] = dataDB[i].DTUDataDetail[4]
			maxMap["抽真空最大值"] = dataDB[i].DTUDataDetail[5]
			minMap["抽真空最小值"] = dataDB[i].DTUDataDetail[5]
			avgMap["抽真空平均值"] = dataDB[i].DTUDataDetail[5]
			maxMap["料室温度最大值"] = dataDB[i].DTUDataDetail[10]
			maxMap["变形区温度最大值"] = dataDB[i].DTUDataDetail[11]
		}
		if stringCompare(dataDB[i].DTUDataDetail[keyMap["k"]].Value, maxMap["压力最大值"].Value) {
			maxMap["压力最大值"] = dataDB[i].DTUDataDetail[keyMap["k"]]
		} else {
			minMap["压力最小值"] = dataDB[i].DTUDataDetail[keyMap["k"]]
		}
		if stringCompare(dataDB[i].DTUDataDetail[keyMap["c"]].Value, maxMap["抽真空最大值"].Value) {
			maxMap["抽真空最大值"] = dataDB[i].DTUDataDetail[keyMap["c"]]
		} else {
			minMap["抽真空最小值"] = dataDB[i].DTUDataDetail[keyMap["c"]]
		}
		if stringCompare(dataDB[i].DTUDataDetail[keyMap["l2"]].Value, maxMap["变形区温度最大值"].Value) {
			maxMap["变形区温度最大值"] = dataDB[i].DTUDataDetail[keyMap["l2"]]
		}
		if stringCompare(dataDB[i].DTUDataDetail[keyMap["l1"]].Value, maxMap["料室温度最大值"].Value) {
			maxMap["料室温度最大值"] = dataDB[i].DTUDataDetail[keyMap["l1"]]
		}

		//累加求平均值
		//累加和sum
		value, err := strconv.ParseFloat(dataDB[i].DTUDataDetail[4].Value, 64)
		if err != nil {
			fmt.Println("类型转化出错", err.Error())
			return
		}
		value2, err := strconv.ParseFloat(dataDB[i].DTUDataDetail[5].Value, 64)
		if err != nil {
			fmt.Println("类型转化出错", err.Error())
			return
		}
		k1sum += value
		k2sum += value2
	}

	//计算平均值
	k1avg = fmt.Sprintf("%f", float32(k1sum/float64(len(dataDB))))
	k2avg = fmt.Sprintf("%f", float32(k2sum/float64(len(dataDB))))
	avgMap["压力平均值"] = model.DTUDataDetail{
		Key:   "k1",
		Value: k1avg,
		Unit:  "",
	}
	avgMap["抽真空度平均值"] = model.DTUDataDetail{
		Key:   "k2",
		Value: k2avg,
		Unit:  "",
	}
	E5extrusionReport(startTime, endTime, maxMap, minMap, avgMap)
}

func E5extrusionReport(startTime string, endTime string, maxMap map[string]model.DTUDataDetail, minMap map[string]model.DTUDataDetail, avgMap map[string]model.DTUDataDetail) {
	var report model.ExtrusionReport
	if stringCompare(maxMap["压力最大值"].Value, "18") {
		//压力最大值>18为预压，否则为挤压
		report.Status = "预压"
	} else {
		report.Status = "挤压"
	}
	report.Name = "挤压报表"
	report.StartTime = startTime
	report.EndTime = endTime
	report.CreateTime = utils.TimeFormat(time.Now())
	report.MaxMap = maxMap
	report.MinMap = minMap
	report.AvgMap = avgMap
	report.Code = time.Now().Format("20060102150405")
	err := global.E5extrusionReport.FindOneAndUpdate(context.TODO(), bson.M{"createTime": report.StartTime}, bson.M{"$set": report}).Decode(&bson.M{})
	if err != nil {
		_, err = global.E5extrusionReport.InsertOne(context.TODO(), report)
		if err != nil {
			fmt.Println("E2extrusionReport，数据存储失败", err)
		}
	}
}
