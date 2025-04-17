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
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strconv"
	"strings"
	"time"
)

// 石墨化数据定时存储
func GraphitingDataOperation(interval time.Duration) {
	endTime := time.Now().Format("2006-01-02 15:04:05")
	startTime := time.Now().Add(-interval).Format("2006-01-02 15:04:05")
	var box model.Box
	err := global.GraphitingDataColl.FindOne(context.TODO(), bson.M{}).Decode(&box)
	if err != nil {
		fmt.Println("GraphitingDataOperation", err.Error())
	}
	if interval == time.Minute*15 {
		if box.Data == nil {
			fmt.Println("石墨化15分钟最新数据为空")
			return
		}
		//石墨化plc添加吨/电量
		if box.BoxId == "be67c2b8216e49e8981a95663413f115" {
			for _, v := range box.Data[0].Detail {
				if v.Key == "有功电量" {
					var List model.GraphiteOriginList
					value, _ := strconv.ParseFloat(v.Value, 64)
					if err := global.GraphiteOriginList.FindOne(context.TODO(), bson.M{}).Decode(&List); err != nil {
						fmt.Println("查询原料表吨数失败", err.Error())
						return
					}
					box.Data[0].Detail = append(box.Data[0].Detail, model.BoxDataDetail{Key: "吨/电量", Value: fmt.Sprintf("%f", value/List.RealWeight), Unit: "kwh/t"})
					box.Data[0].Detail = append(box.Data[0].Detail, model.BoxDataDetail{Key: "规格", Value: List.Name, Unit: ""})
					box.Data[0].Detail = append(box.Data[0].Detail, model.BoxDataDetail{Key: "本体重量", Value: fmt.Sprintf("%f", List.RealWeight), Unit: "t"})
					box.Data[0].Detail = append(box.Data[0].Detail, model.BoxDataDetail{Key: "附属品重量", Value: fmt.Sprintf("%f", List.AccessWeight), Unit: "t"})
				}
			}
		}
		box.CreateTime = endTime
		box.Id = primitive.ObjectID{}
		box.UpdateTime = ""
		storeGraData(box, interval)
		return
	}
	var datas []model.Box
	filter := bson.M{
		"createTime": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
		"boxId": box.BoxId,
	}
	if err := utils.Find(global.GraphitingHisDataColl, &datas, filter); err != nil {
		fmt.Println("GraphitingDataOperation", err.Error())
	}
	if datas == nil {
		fmt.Println("石墨化此时间段无数据")
		return
	}
	//规定好顺序,避免乱序问题
	var graDatas []HisDataType
	var ku []SortKV
	for _, a := range datas[0].Data {
		graDatas = append(graDatas, HisDataType{
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
			for i, c := range graDatas {
				if c.SensorIdAndName == datum.SensorId+"|"+datum.SensorName {
					graDatas[i].Details = append(graDatas[i].Details, datum.Detail...)
				}
			}
		}
	}
	var comGraDatas []model.BoxData
	//分组传感器中的检测值
	for _, graData := range graDatas {
		for _, datum := range graData.Details {
			for i, b := range ku {
				if b.KeyAndUnit == datum.Key+"|"+datum.Unit {
					ku[i].value = append(ku[i].value, datum.Value)
				}
			}
		}
		//求最大值平均值最小值
		var comGraDataDetail []model.BoxDataDetail
		for i, c := range ku {
			keyUnit := strings.Split(c.KeyAndUnit, "|")
			if i == 4 || i >= 18 || i == 2 || i == 16 {
				comGraDataDetail = append(comGraDataDetail, []model.BoxDataDetail{
					{
						Key:   keyUnit[0],
						Value: c.value[len(c.value)-1],
						Unit:  keyUnit[1],
					},
				}...)
			} else {
				comGraDataDetail = append(comGraDataDetail, []model.BoxDataDetail{
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
		comGraDatas = append(comGraDatas, model.BoxData{
			SensorId:   idName[0],
			SensorName: idName[1],
			Detail:     comGraDataDetail,
		})
	}
	box.Data = comGraDatas

	//石墨化PLC添加电量/吨
	if box.BoxId == "be67c2b8216e49e8981a95663413f115" {
		for _, v := range box.Data[0].Detail {
			if v.Key == "有功电量" {
				var List model.GraphiteOriginList
				value, _ := strconv.ParseFloat(v.Value, 64)
				if err := global.GraphiteOriginList.FindOne(context.TODO(), bson.M{}).Decode(&List); err != nil {
					fmt.Println("查询原料表吨数失败", err.Error())
					return
				}
				box.Data[0].Detail = append(box.Data[0].Detail, model.BoxDataDetail{Key: "吨/电量", Value: fmt.Sprintf("%f", value/List.RealWeight), Unit: "kwh/t"})
				box.Data[0].Detail = append(box.Data[0].Detail, model.BoxDataDetail{Key: "规格", Value: List.Name, Unit: ""})
				box.Data[0].Detail = append(box.Data[0].Detail, model.BoxDataDetail{Key: "本体重量", Value: fmt.Sprintf("%f", List.RealWeight), Unit: "t"})
				box.Data[0].Detail = append(box.Data[0].Detail, model.BoxDataDetail{Key: "附属品重量", Value: fmt.Sprintf("%f", List.AccessWeight), Unit: "t"})
			}
		}
	}
	box.Id, box.UpdateTime, box.CreateTime = primitive.ObjectID{}, "", endTime

	storeGraData(box, interval)
}

// 存储石墨化10 30 小时数据
func storeGraData(box model.Box, interval time.Duration) {
	var collectionHis *mongo.Collection
	var s string
	switch interval {
	case time.Minute * 10:
		collectionHis = global.GraphitinTenDataColl
		s = "10分钟数据存储成功"
	case time.Minute * 15:
		collectionHis = global.GraphitinFifteenDataColl
		s = "15分钟数据存储成功"
	case time.Minute * 30:
		collectionHis = global.GraphitingThirtyDataColl
		s = "30分钟数据存储成功"
	case time.Hour:
		collectionHis = global.GraphitingHourDataColl
		s = "60分钟数据存储成功"
	default:
		return
	}
	if _, err := collectionHis.InsertOne(context.TODO(), box); err != nil {
		fmt.Println("数据存储出错", err.Error())
		return
	}
	fmt.Println("石墨化" + s)
}

// 存储石墨化报表数据(根据有功功率判断)
func storeGraReportData(box model.Box) {
	now := time.Now().Format("2006-01-02 15:04:05")
	var report model.GraReportForm
	report.CreateTime = now                      //报表创建时间
	report.EndTime = box.UpdateTime              //结束时间
	report.StoveNumber = box.Data[0].StoveNumber //炉号
	report.StartTime = box.Data[0].StartTime     //送电时刻
	rt := RetRunTimeAndHead(box.Data[0].Detail)
	report.RunTime = rt.RunTime     //送电运行时长
	report.HeadTitle = rt.HeadTitle //标题
	//如果该炉报表已创建则结束(否则报表记录会重复创建)
	var one interface{}
	err := global.GraReportFormDataColl.FindOne(context.TODO(), bson.M{"startTime": report.StartTime, "stoveNumber": report.StoveNumber}).Decode(&one)
	if err == nil {
		fmt.Println(report.HeadTitle + "报表记录已创建")
		return
	}
	//然后处理数据，从石墨化15分钟表里获取该炉数据
	findOptions := new(options.FindOptions)
	findOptions = &options.FindOptions{}
	findOptions.SetSort(bson.D{{"createTime", -1}})
	var data []model.Box
	cur, err := global.GraphitinFifteenDataColl.Find(context.TODO(), bson.M{"createTime": bson.M{"$gte": report.StartTime, "$lte": report.EndTime}}, findOptions)
	if err != nil {
		log.Println(err)
	}
	if err = cur.All(context.TODO(), &data); err != nil {
		log.Println(err)
	}
	//从设备表里获取key unit等信息，先筛选要存储的信息
	var device model.Device
	err = global.DeviceColl.FindOne(context.TODO(), bson.M{"code": box.BoxId}).Decode(&device)
	if err != nil {
		fmt.Println(err)
	}
	//这些变量要存整个生产中每15分钟一次的数据
	var variants []string
	variants = append(variants, "直流电压", "直流电流", "档位显示", "有功功率", "AB线电压(U1)", "BC线电压(U2)", "AC线电压(U3)",
		"IA(A1)", "IB(A2)", "IC(A3)", "变压器温度", "A柜水温", "B柜水温", "位移", "功率因数", "直流功率", "炉阻", "功率设定")
	var repDetail1 []model.ReportDetail
	//先添加key和unit
	for _, a := range device.Sensors[0].DetectionValue {
		for _, b := range variants {
			if a.Key == b {
				repDetail1 = append(repDetail1, model.ReportDetail{
					Key:  a.Key,
					Unit: a.Unit,
					VT:   nil,
				})
				break
			} else {
				continue
			}
		}
	}
	//添加每个变量的15分钟值
	for _, a := range data {
		for _, b := range a.Data[0].Detail {
			for i, c := range repDetail1 {
				if b.Key == c.Key {
					var vt model.ValueAndTime
					vt.CreateTime = a.CreateTime
					vt.Value = b.Value
					repDetail1[i].VT = append(repDetail1[i].VT, vt)
					break
				} else {
					continue
				}
			}
		}
	}
	//要添加开始和结束两个时间的时刻数据
	startMin := report.StartTime[len(report.StartTime)-5 : len(report.StartTime)-3] //开始时间的分钟
	endMin := report.EndTime[len(report.EndTime)-5 : len(report.EndTime)-3]         //结束时间的分钟
	//开始时刻值
	if startMin != "00" && startMin != "15" && startMin != "30" && startMin != "45" {
		var startBox model.Box
		err = global.GraphitingHisDataColl.FindOne(context.TODO(), bson.M{"createTime": report.StartTime}).Decode(&startBox)
		if err == nil {
			for _, a := range startBox.Data[0].Detail {
				for i, c := range repDetail1 {
					if a.Key == c.Key {
						var vt1 []model.ValueAndTime
						vt1 = append(vt1, model.ValueAndTime{
							CreateTime: startBox.CreateTime,
							Value:      a.Value,
						})
						repDetail1[i].VT = append(vt1, repDetail1[i].VT...)
						break
					} else {
						continue
					}
				}
			}
		}
	}
	//结束时刻值
	if endMin != "00" && endMin != "15" && endMin != "30" && endMin != "45" {
		var endBox model.Box
		err = global.GraphitingHisDataColl.FindOne(context.TODO(), bson.M{"createTime": report.EndTime}).Decode(&endBox)
		if err == nil {
			for _, a := range endBox.Data[0].Detail {
				for i, c := range repDetail1 {
					if a.Key == c.Key {
						var vt1 []model.ValueAndTime
						vt1 = append(vt1, model.ValueAndTime{
							CreateTime: endBox.CreateTime,
							Value:      a.Value,
						})
						repDetail1[i].VT = append(repDetail1[i].VT, vt1...)
						break
					} else {
						continue
					}
				}
			}
		}
	}
	//计算总电量 总价等时刻值
	var totalPrice float64
	var totalKWH float64
	for i, a := range box.Data[0].Detail {
		if strings.Contains(a.Key, "电度") == true {
			flo, _ := strconv.ParseFloat(a.Value, 64)
			totalKWH += flo
		}
		if strings.Contains(a.Key, "总价") == true {
			flo, _ := strconv.ParseFloat(a.Value, 64)
			totalPrice += flo
		}
		if i == 4 || i >= 33 && i <= 44 {
			var vt []model.ValueAndTime
			vt = append(vt, model.ValueAndTime{
				CreateTime: box.UpdateTime,
				Value:      a.Value,
			})
			repDetail1 = append(repDetail1, model.ReportDetail{
				Key:  a.Key,
				Unit: a.Unit,
				VT:   vt,
			})
		}
	}
	//总电度
	var vt1 []model.ValueAndTime
	vt1 = append(vt1, model.ValueAndTime{
		CreateTime: box.UpdateTime,
		Value:      fmt.Sprintf("%.2f", totalKWH),
	})
	//总价
	var vt2 []model.ValueAndTime
	vt2 = append(vt2, model.ValueAndTime{
		CreateTime: box.UpdateTime,
		Value:      fmt.Sprintf("%.2f", totalPrice),
	})
	//平均电度价格
	var vt3 []model.ValueAndTime
	vt3 = append(vt3, model.ValueAndTime{
		CreateTime: box.UpdateTime,
		Value:      fmt.Sprintf("%f", totalPrice/totalKWH),
	})
	repDetail1 = append(repDetail1, model.ReportDetail{
		Key:  "总电度",
		VT:   vt1,
		Unit: "kWh",
	}, model.ReportDetail{
		Key:  "总价",
		VT:   vt2,
		Unit: "元",
	}, model.ReportDetail{
		Key:  "电度平均价格",
		VT:   vt3,
		Unit: "",
	})
	report.Data = append(report.Data, repDetail1)
	insertOneResult, err := global.GraReportFormDataColl.InsertOne(context.TODO(), report)
	if err != nil {
		fmt.Println("石墨化报表err:", err)
	}
	fmt.Println("石墨化报表数据创建成功:", insertOneResult.InsertedID)
}

// 存储石墨化报表数据(根据有功电量判断)
func storeGraReportByElectricity(lastElectricTime, nowElectricTime string) {
	now := time.Now().Format("2006-01-02 15:04:05")
	cur, err := global.GraPowerTimeColl.Find(context.TODO(), bson.M{"createTime": bson.M{"$gt": lastElectricTime, "$lt": nowElectricTime}})
	if err != nil {
		fmt.Println(err)
	}
	var powerTime []model.GraPower
	if err = cur.All(context.TODO(), &powerTime); err != nil {
		fmt.Println(err)
	}
	if powerTime == nil {
		fmt.Println("功率时刻值为空")
		return
	}
	var report model.GraReportForm
	report.CreateTime = now                                 //报表创建时间
	report.EndTime = powerTime[len(powerTime)-1].CreateTime //结束时间
	report.StartTime = powerTime[0].StartTime               //送电时刻
	//从石墨化15分钟表里获取该炉数据
	var data1 []model.Box
	cur, err = global.GraphitinFifteenDataColl.Find(context.TODO(), bson.M{"createTime": bson.M{"$gte": report.StartTime, "$lte": report.EndTime}})
	if err != nil {
		log.Println(err)
	}
	if err = cur.All(context.TODO(), &data1); err != nil {
		log.Println(err)
	}
	if data1 == nil {
		fmt.Println("石墨化工艺报表在此时间段无15分钟数据")
		return
	}
	//获取清零之前最后一个box 获取运行时间和电量情况
	var box model.Box
	err = global.GraphitingHisDataColl.FindOne(context.TODO(), bson.M{"createTime": report.EndTime}).Decode(&box)
	if err != nil {
		fmt.Println(err)
	}
	//报表结束时间为,有功功率小于1的前一个有功功率大于1的创建时间
	for i := len(data1) - 1; i > 0; i-- {
		power1, _ := strconv.Atoi(data1[i].Data[0].Detail[3].Value)
		power2, _ := strconv.Atoi(data1[i-1].Data[0].Detail[3].Value)
		if power1 < 1 && power2 > 1 {
			report.EndTime = data1[i-1].CreateTime
			break
		}
	}
	var data []model.Box
	cur, err = global.GraphitinFifteenDataColl.Find(context.TODO(), bson.M{"createTime": bson.M{"$gte": report.StartTime, "$lte": report.EndTime}})
	if err != nil {
		log.Println(err)
	}
	if err = cur.All(context.TODO(), &data); err != nil {
		log.Println(err)
	}
	year := report.StartTime[0:4]
	month := report.StartTime[5:7]
	day := report.StartTime[8:10]
	hour := report.StartTime[11:13]
	report.HeadTitle = "SMH" + year + month + day + hour //标题
	report.StoveNumber = box.Data[0].StoveNumber         //炉号
	rt2 := RetRunTimeAndHead2(box.Data[0].Detail)
	report.RunTime = rt2.RunTime //送电运行时长
	//如果该炉报表已创建则结束(否则报表记录会重复创建)
	var one interface{}
	err = global.GraphitePlcReportColl.FindOne(context.TODO(), bson.M{"startTime": report.StartTime, "stoveNumber": report.StoveNumber}).Decode(&one)
	if err == nil {
		fmt.Println("报表重复")
		return
	}
	//从设备表里获取key unit等信息，先筛选要存储的信息
	var device model.Device
	err = global.DeviceColl.FindOne(context.TODO(), bson.M{"code": box.BoxId}).Decode(&device)
	if err != nil {
		fmt.Println(err)
	}
	//这些变量要存整个生产中每15分钟一次的数据
	var variants []string
	variants = append(variants, "直流电压", "直流电流", "档位显示", "有功功率", "AB线电压(U1)", "BC线电压(U2)", "AC线电压(U3)",
		"IA(A1)", "IB(A2)", "IC(A3)", "变压器温度", "A柜水温", "B柜水温", "位移", "功率因数", "直流功率", "炉阻", "功率设定")
	var repDetail1 []model.ReportDetail
	//先添加key和unit
	for _, a := range device.Sensors[0].DetectionValue {
		for _, b := range variants {
			if a.Key == b {
				repDetail1 = append(repDetail1, model.ReportDetail{
					Key:  a.Key,
					Unit: a.Unit,
					VT:   nil,
				})
				break
			} else {
				continue
			}
		}
	}
	//添加每个变量的15分钟值
	for _, a := range data {
		for _, b := range a.Data[0].Detail {
			for i, c := range repDetail1 {
				if b.Key == c.Key {
					var vt model.ValueAndTime
					vt.CreateTime = a.CreateTime
					vt.Value = b.Value
					repDetail1[i].VT = append(repDetail1[i].VT, vt)
					break
				} else {
					continue
				}
			}
		}
	}
	//计算总电量 总价等时刻值
	var totalPrice float64
	var totalKWH float64
	for i, a := range box.Data[0].Detail {
		if strings.Contains(a.Key, "电度") == true {
			flo, _ := strconv.ParseFloat(a.Value, 64)
			totalKWH += flo
		}
		if strings.Contains(a.Key, "总价") == true {
			flo, _ := strconv.ParseFloat(a.Value, 64)
			totalPrice += flo
		}
		if i == 4 || i >= 33 && i <= 44 {
			var vt []model.ValueAndTime
			vt = append(vt, model.ValueAndTime{
				CreateTime: box.CreateTime,
				Value:      a.Value,
			})
			repDetail1 = append(repDetail1, model.ReportDetail{
				Key:  a.Key,
				Unit: a.Unit,
				VT:   vt,
			})
		}
	}
	//总电度
	var vt1 []model.ValueAndTime
	vt1 = append(vt1, model.ValueAndTime{
		CreateTime: box.CreateTime,
		Value:      fmt.Sprintf("%.2f", totalKWH),
	})
	//总价
	var vt2 []model.ValueAndTime
	vt2 = append(vt2, model.ValueAndTime{
		CreateTime: box.CreateTime,
		Value:      fmt.Sprintf("%.2f", totalPrice),
	})
	//平均电度价格
	var vt3 []model.ValueAndTime
	vt3 = append(vt3, model.ValueAndTime{
		CreateTime: box.CreateTime,
		Value:      fmt.Sprintf("%f", totalPrice/totalKWH),
	})
	repDetail1 = append(repDetail1, model.ReportDetail{
		Key:  "总电度",
		VT:   vt1,
		Unit: "kWh",
	}, model.ReportDetail{
		Key:  "总价",
		VT:   vt2,
		Unit: "元",
	}, model.ReportDetail{
		Key:  "电度平均价格",
		VT:   vt3,
		Unit: "",
	})
	report.Data = append(report.Data, repDetail1)
	insertOneResult, err := global.GraphitePlcReportColl.InsertOne(context.TODO(), report)
	if err != nil {
		fmt.Println("石墨化生产工艺报表(有功电量)err:", err)
	}
	fmt.Println("石墨化生产工艺报表数据创建成功(有功电量):", insertOneResult.InsertedID)
}

// 运行时间和标题类型 用于拼好后返回
type RTAndHead struct {
	RunTime   string
	HeadTitle string
}

// 拼好送电时长和标题
func RetRunTimeAndHead(boxDetail []model.BoxDataDetail) RTAndHead {
	var data RTAndHead
	var runMin string  //送电时长分钟
	var runHour string //送电时长小时
	var day string     //时刻天
	var month string   //时刻月
	var year string    //时刻年
	var hour string    //小时
	var Min string     //分钟
	for _, a := range boxDetail {
		if a.Key == "分" {
			runMin = a.Value
		} else if a.Key == "时" {
			runHour = a.Value
		} else if a.Key == "日" {
			intValue, _ := strconv.Atoi(a.Value)
			if intValue < 10 {
				day = "0" + a.Value
			} else {
				day = a.Value
			}
		} else if a.Key == "分钟" {
			intValue, _ := strconv.Atoi(a.Value)
			if intValue < 10 {
				Min = "0" + a.Value
			} else {
				Min = a.Value
			}
		} else if a.Key == "月" {
			intValue, _ := strconv.Atoi(a.Value)
			if intValue < 10 {
				month = "0" + a.Value
			} else {
				month = a.Value
			}
		} else if a.Key == "年" {
			year = a.Value
		} else if a.Key == "小时" {
			intValue, _ := strconv.Atoi(a.Value)
			if intValue < 10 {
				hour = "0" + a.Value
			} else {
				hour = a.Value
			}
		} else {
			continue
		}
	}
	data.RunTime = runHour + "小时" + runMin + "分"
	data.HeadTitle = "SMHPLC" + year + month + day + hour + Min //标题
	return data
}

// 拼好送电时长和标题
func RetRunTimeAndHead2(boxDetail []model.BoxDataDetail) RTAndHead {
	var data RTAndHead
	var runMin string  //送电时长分钟
	var runHour string //送电时长小时
	var day string     //时刻天
	var month string   //时刻月
	var year string    //时刻年
	var hour string    //小时
	for _, a := range boxDetail {
		if a.Key == "分" {
			runMin = a.Value
		} else if a.Key == "时" {
			runHour = a.Value
		} else if a.Key == "日" {
			intValue, _ := strconv.Atoi(a.Value)
			if intValue < 10 {
				day = "0" + a.Value
			} else {
				day = a.Value
			}
		} else if a.Key == "月" {
			intValue, _ := strconv.Atoi(a.Value)
			if intValue < 10 {
				month = "0" + a.Value
			} else {
				month = a.Value
			}
		} else if a.Key == "年" {
			year = a.Value
		} else if a.Key == "小时" {
			intValue, _ := strconv.Atoi(a.Value)
			if intValue < 10 {
				hour = "0" + a.Value
			} else {
				hour = a.Value
			}
		} else {
			continue
		}
	}
	data.RunTime = runHour + "小时" + runMin + "分"
	data.HeadTitle = "SMH" + year + month + day + hour //标题
	return data
}
