package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"strconv"
	"strings"
)

//脚本
func Back() {
	//遍历4月30天
	for i := 1; i <= 30; i++ {
		var day string
		if i < 10 {
			day = "0" + strconv.Itoa(i)
		} else {
			day = strconv.Itoa(i)
		}
		date := "2022-04-" + day + " 00:10:00"
		//fmt.Println(date)
		var one model.Box
		err := global.GraphitingHisDataColl.FindOne(context.TODO(), bson.M{"createTime": date}).Decode(&one)
		if err != nil {
			fmt.Println(err)
			continue
		}
		var active float64
		var ele1 float64
		var ele2 float64
		var ele3 float64
		var ele4 float64
		var totalPrice float64
		for _, a := range one.Data[0].Detail {
			flo, _ := strconv.ParseFloat(a.Value, 64)
			if a.Key == "有功电量" {
				active = flo
			} else if a.Key == "低谷电度" {
				ele1 = flo
			} else if a.Key == "平段电度" {
				ele2 = flo
			} else if a.Key == "高峰电度" {
				ele3 = flo
			} else if a.Key == "尖峰电度" {
				ele4 = flo
			} else if strings.Contains(a.Key, "总价") {
				totalPrice += flo
			} else {
				continue
			}
		}
		totalElectric := ele1 + ele2 + ele3 + ele4
		if active == totalElectric {
			continue
		} else {
			//fmt.Println(date)
			date2 := "2022-04-" + day + " 00:00:00"
			date3 := "2022-04-" + day + " 01:00:00"
			if date=="2022-04-20 00:10:00"{
				date2="2022-04-20 00:00:01"
			}
			//0:00数据
			var box1 model.Box
			err := global.GraphitingHisDataColl.FindOne(context.TODO(), bson.M{"createTime": date2}).Decode(&box1)
			if err != nil {
				continue
				//fmt.Println("0:00--", err)
			}
			//1:00数据
			var box2 model.Box
			err = global.GraphitingHisDataColl.FindOne(context.TODO(), bson.M{"createTime": date3}).Decode(&box2)
			if err != nil {
				continue
				//fmt.Println("1:00--", err)
			}
			var low1 float64
			var price1 float64
			var low2 float64
			var price2 float64
			for _, b1 := range box1.Data[0].Detail {
				if b1.Key == "低谷电度" {
					flo, _ := strconv.ParseFloat(b1.Value, 64)
					low1 = flo
				} else if b1.Key == "低谷总价" {
					flo, _ := strconv.ParseFloat(b1.Value, 64)
					price1 = flo
				} else {
					continue
				}
			}
			for _, b2 := range box2.Data[0].Detail {
				if b2.Key == "低谷电度" {
					flo, _ := strconv.ParseFloat(b2.Value, 64)
					low2 = flo
				} else if b2.Key == "低谷总价" {
					flo, _ := strconv.ParseFloat(b2.Value, 64)
					price2 = flo
				} else {
					continue
				}
			}
			var realLow float64
			var realPrice float64
			realLow = (low2 - low1) / 2
			realPrice = (price2 - price1) / 2
			fmt.Println("日期:", date, "有功电量:", active, "合计耗电量:", totalElectric, "低谷电度差值:", realLow,
				"低谷电价差值:", realPrice, "总价:", totalPrice, "平均单价:", fmt.Sprintf("%f", totalPrice/totalElectric))
		}
	}
}

func GraSearchTest() {
	//var lastElectric []model.GraElectric
	//findOptions := new(options.FindOptions)
	//findOptions = &options.FindOptions{}
	//findOptions.SetSort(bson.D{{"createTime", -1}})
	//findOptions.SetLimit(1) //每页数据数量
	//cur, err := global.GraElectricTimeColl.Find(context.TODO(), bson.M{}, findOptions)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//if err = cur.All(context.TODO(), &lastElectric); err != nil {
	//	fmt.Println(err)
	//}
	cur, err := global.GraPowerTimeColl.Find(context.TODO(), bson.M{"createTime": bson.M{"$gte": "2022-04-04 07:05:00", "$lte": "2022-04-05 08:07:00"}})
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
	fmt.Println(powerTime[len(powerTime)-1])
	fmt.Println(powerTime[0])
	var report model.GraReportForm
	report.EndTime = powerTime[len(powerTime)-1].CreateTime //结束时间
	report.StartTime = powerTime[0].StartTime               //送电时刻
	var data []model.Box
	cur, err = global.GraphitinFifteenDataColl.Find(context.TODO(), bson.M{"createTime": bson.M{"$gte": report.StartTime, "$lte": report.EndTime}})
	if err != nil {
		log.Println(err)
	}
	if err = cur.All(context.TODO(), &data); err != nil {
		log.Println(err)
	}
	var box model.Box
	if data == nil {
		fmt.Println("石墨化工艺报表在此时间段无15分钟数据")
		return
	} else if len(data) >= 1 {
		box = data[len(data)-1]
	}
	fmt.Println(box)
	fmt.Println(data[0])
}
func Test2() {
	var gra model.Box
	err := global.GraphitingHisDataColl.FindOne(context.TODO(), bson.M{"createTime": "2022-04-24 11:04:00"}).Decode(&gra)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(gra)
	//StoreGraReportByElectricity2("2022-04-23 15:25:00", "2022-04-24 11:06:00")
}

//石墨化历史数据处理创建报表
func Test() {
	cur, err := global.GraphitingHisDataColl.Find(context.TODO(), bson.M{"createTime": bson.M{"$gte": "2022-03-07 10:40:00"}})
	if err != nil {
		fmt.Println(err)
	}
	var data []model.Box
	if err = cur.All(context.TODO(), &data); err != nil {
		fmt.Println(err)
	}
	var power1 string
	var power2 string
	for i := 0; i < len(data)-1; i++ {
		if i == len(data)-2 {
			break
		}
		power1 = data[i].Data[0].Detail[3].Value
		power2 = data[i+1].Data[0].Detail[3].Value
		intPower1, _ := strconv.Atoi(power1)
		intPower2, _ := strconv.Atoi(power2)
		if intPower1 > 1 && intPower2 < 1 {
			res := RetTimeTest(data[i+1].Data[0].Detail)
			if res.StartTime != "0-0-0 00:00:00" {
				storeGraTest(res, data[i+1])
			}
		} else {
			continue
		}
	}
}

type RetTest struct {
	RunTime     string
	StartTime   string
	HeadTitle   string
	StoveNumber string
}

func RetTimeTest(boxDetail []model.BoxDataDetail) RetTest {
	var data RetTest
	var runMin string  //送电时长分钟
	var runHour string //送电时长小时
	var day1 string    //时刻天
	var month1 string  //时刻月
	var year string    //时刻年
	var min string
	var hour string
	var stoveNumber2 string
	for _, a := range boxDetail {
		if a.Key == "分" {
			runMin = a.Value
		} else if a.Key == "时" {
			runHour = a.Value
		} else if a.Key == "日" {
			intValue, _ := strconv.Atoi(a.Value)
			if intValue < 10 && intValue != 0 {
				day1 = "0" + a.Value
			} else {
				day1 = a.Value
			}
		} else if a.Key == "月" {
			intValue, _ := strconv.Atoi(a.Value)
			if intValue < 10 && intValue != 0 {
				month1 = "0" + a.Value
			} else {
				month1 = a.Value
			}
		} else if a.Key == "年" {
			year = a.Value
		} else if a.Key == "分钟" {
			intValue, _ := strconv.Atoi(a.Value)
			if intValue < 10 {
				min = "0" + a.Value
			} else {
				min = a.Value
			}
		} else if a.Key == "小时" {
			intValue, _ := strconv.Atoi(a.Value)
			if intValue < 10 {
				hour = "0" + a.Value
			} else {
				hour = a.Value
			}
		} else if a.Key == "炉号" {
			stoveNumber2 = a.Value
		} else {
			continue
		}
	}
	data.StartTime = year + "-" + month1 + "-" + day1 + " " + hour + ":" + min + ":00"
	data.RunTime = runHour + "小时" + runMin + "分"
	data.HeadTitle = "SMHPLC" + year + month1 + day1 + hour + min //标题
	data.StoveNumber = stoveNumber2
	return data
}
func storeGraTest(res RetTest, box model.Box) {
	var report model.GraReportForm
	report.CreateTime = box.CreateTime   //报表创建时间
	report.EndTime = box.CreateTime      //结束时间
	report.StoveNumber = res.StoveNumber //炉号
	report.StartTime = res.StartTime     //送电时刻
	report.RunTime = res.RunTime         //送电运行时长
	report.HeadTitle = res.HeadTitle     //标题
	var one interface{}
	err := global.GraphitePlcReportColl.FindOne(context.TODO(), bson.M{"startTime": report.StartTime, "stoveNumber": report.StoveNumber}).Decode(&one)
	if err == nil {
		return
	}
	//然后处理数据，从石墨化15分钟表里获取该炉数据
	var data []model.Box
	cur, err := global.GraphitinFifteenDataColl.Find(context.TODO(), bson.M{"createTime": bson.M{"$gte": report.StartTime, "$lte": report.EndTime}})
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
		fmt.Println("石墨化报表err:", err)
	}
	fmt.Println("石墨化报表数据创建成功:", insertOneResult.InsertedID)
}

//存储石墨化报表数据(根据有功电量判断)
//func StoreGraReportByElectricity2(lastElectricTime, nowElectricTime string) {
//	//now := time.Now().Format("2006-01-02 15:04:05")
//	cur, err := global.GraPowerTimeColl.Find(context.TODO(), bson.M{"createTime": bson.M{"$gt": lastElectricTime, "$lt": nowElectricTime}})
//	if err != nil {
//		fmt.Println(err)
//	}
//	var powerTime []model.GraPower
//	if err = cur.All(context.TODO(), &powerTime); err != nil {
//		fmt.Println(err)
//	}
//	if powerTime == nil {
//		fmt.Println("功率时刻值为空")
//		return
//	}
//	var report model.GraReportForm
//	//report.CreateTime = now                                 //报表创建时间
//	report.EndTime = powerTime[len(powerTime)-1].CreateTime //结束时间
//	report.StartTime = powerTime[0].StartTime               //送电时刻
//	//从石墨化15分钟表里获取该炉数据
//	var data1 []model.Box
//	cur, err = global.GraphitinFifteenDataColl.Find(context.TODO(), bson.M{"createTime": bson.M{"$gte": report.StartTime, "$lte": report.EndTime}})
//	if err != nil {
//		log.Println(err)
//	}
//	if err = cur.All(context.TODO(), &data1); err != nil {
//		log.Println(err)
//	}
//	if data1 == nil {
//		fmt.Println("石墨化工艺报表在此时间段无15分钟数据")
//		return
//	}
//	//获取清零之前最后一个box 获取运行时间和电量情况
//	var box model.Box
//	err = global.GraphitingHisDataColl.FindOne(context.TODO(), bson.M{"createTime": report.EndTime}).Decode(&box)
//	if err != nil {
//		fmt.Println(err)
//	}
//	//报表结束时间为,有功功率小于1的前一个有功功率大于1的创建时间
//	for i := len(data1) - 1; i > 0; i-- {
//		power1, _ := strconv.Atoi(data1[i].Data[0].Detail[3].Value)
//		power2, _ := strconv.Atoi(data1[i-1].Data[0].Detail[3].Value)
//		if power1 < 1 && power2 > 1 {
//			report.EndTime = data1[i-1].CreateTime
//			break
//		}
//	}
//	var data []model.Box
//	cur, err = global.GraphitinFifteenDataColl.Find(context.TODO(), bson.M{"createTime": bson.M{"$gte": report.StartTime, "$lte": report.EndTime}})
//	if err != nil {
//		log.Println(err)
//	}
//	if err = cur.All(context.TODO(), &data); err != nil {
//		log.Println(err)
//	}
//	year := report.StartTime[0:4]
//	month := report.StartTime[5:7]
//	day := report.StartTime[8:10]
//	hour := report.StartTime[11:13]
//	report.HeadTitle = "SMH" + year + month + day + hour //标题
//	report.StoveNumber = box.Data[0].StoveNumber         //炉号
//	rt2 := RetRunTimeAndHead2(box.Data[0].Detail)
//	report.RunTime = rt2.RunTime //送电运行时长
//	//如果该炉报表已创建则结束(否则报表记录会重复创建)
//	//var one interface{}
//	//err = global.GraphitePlcReportColl.FindOne(context.TODO(), bson.M{"startTime": report.StartTime, "stoveNumber": report.StoveNumber}).Decode(&one)
//	//if err == nil {
//	//	fmt.Println("报表重复")
//	//	return
//	//}
//	//从设备表里获取key unit等信息，先筛选要存储的信息
//	var device model.Device
//	err = global.DeviceColl.FindOne(context.TODO(), bson.M{"code": box.BoxId}).Decode(&device)
//	if err != nil {
//		fmt.Println(err)
//	}
//	//这些变量要存整个生产中每15分钟一次的数据
//	var variants []string
//	variants = append(variants, "直流电压", "直流电流", "档位显示", "有功功率", "AB线电压(U1)", "BC线电压(U2)", "AC线电压(U3)",
//		"IA(A1)", "IB(A2)", "IC(A3)", "变压器温度", "A柜水温", "B柜水温", "位移", "功率因数", "直流功率", "炉阻", "功率设定")
//	var repDetail1 []model.ReportDetail
//	//先添加key和unit
//	for _, a := range device.Sensors[0].DetectionValue {
//		for _, b := range variants {
//			if a.Key == b {
//				repDetail1 = append(repDetail1, model.ReportDetail{
//					Key:  a.Key,
//					Unit: a.Unit,
//					VT:   nil,
//				})
//				break
//			} else {
//				continue
//			}
//		}
//	}
//	//添加每个变量的15分钟值
//	for _, a := range data {
//		for _, b := range a.Data[0].Detail {
//			for i, c := range repDetail1 {
//				if b.Key == c.Key {
//					var vt model.ValueAndTime
//					vt.CreateTime = a.CreateTime
//					vt.Value = b.Value
//					repDetail1[i].VT = append(repDetail1[i].VT, vt)
//					break
//				} else {
//					continue
//				}
//			}
//		}
//	}
//	//计算总电量 总价等时刻值
//	var totalPrice float64
//	var totalKWH float64
//	for i, a := range box.Data[0].Detail {
//		if strings.Contains(a.Key, "电度") == true {
//			flo, _ := strconv.ParseFloat(a.Value, 64)
//			totalKWH += flo
//		}
//		if strings.Contains(a.Key, "总价") == true {
//			flo, _ := strconv.ParseFloat(a.Value, 64)
//			totalPrice += flo
//		}
//		if i == 4 || i >= 33 && i <= 44 {
//			var vt []model.ValueAndTime
//			vt = append(vt, model.ValueAndTime{
//				CreateTime: box.CreateTime,
//				Value:      a.Value,
//			})
//			repDetail1 = append(repDetail1, model.ReportDetail{
//				Key:  a.Key,
//				Unit: a.Unit,
//				VT:   vt,
//			})
//		}
//	}
//	//总电度
//	var vt1 []model.ValueAndTime
//	vt1 = append(vt1, model.ValueAndTime{
//		CreateTime: box.CreateTime,
//		Value:      fmt.Sprintf("%.2f", totalKWH),
//	})
//	//总价
//	var vt2 []model.ValueAndTime
//	vt2 = append(vt2, model.ValueAndTime{
//		CreateTime: box.CreateTime,
//		Value:      fmt.Sprintf("%.2f", totalPrice),
//	})
//	//平均电度价格
//	var vt3 []model.ValueAndTime
//	vt3 = append(vt3, model.ValueAndTime{
//		CreateTime: box.CreateTime,
//		Value:      fmt.Sprintf("%f", totalPrice/totalKWH),
//	})
//	repDetail1 = append(repDetail1, model.ReportDetail{
//		Key:  "总电度",
//		VT:   vt1,
//		Unit: "kWh",
//	}, model.ReportDetail{
//		Key:  "总价",
//		VT:   vt2,
//		Unit: "元",
//	}, model.ReportDetail{
//		Key:  "电度平均价格",
//		VT:   vt3,
//		Unit: "",
//	})
//	report.Data = append(report.Data, repDetail1)
//	update := bson.M{"$set": report}
//	objId, _ := primitive.ObjectIDFromHex("6264be98cd4b960338b8a63c")
//	Result, err := global.GraphitePlcReportColl.UpdateOne(context.TODO(), bson.M{"_id": objId}, update)
//	if err != nil {
//		fmt.Println("石墨化生产工艺报表(有功电量)err:", err)
//	}
//	fmt.Println("石墨化生产工艺报表数据修改成功(有功电量):", Result)
//}
