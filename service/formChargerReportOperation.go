package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
)

//压型配料报表定时计算 存储   小时
func FormChargerReportOperation(interval time.Duration) {
	endTime := time.Now().Format("2006-01-02 15:04:05")
	startTime := time.Now().Add(-interval).Format("2006-01-02 15:04:05")

	filter := bson.M{
		"createTime": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}

	var boxDB []model.Box
	res, err := global.FormPlcHisDataColl.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("成型PLC数据查询失败", err)
		return
	}

	if err = res.All(context.TODO(), &boxDB); err != nil && res == nil {
		fmt.Println(err, "data:", res)
	}

	if boxDB == nil {
		fmt.Println("成型PLC数据目前时间段无历史数据")
		return
	}

	var lastBox model.Box //仅当i=0时使用,i=0的上一条数据
	var data []model.FormChargerReport

	//boxLen := len(boxDB)
	for i, box := range boxDB {
		if len(box.Data[0].Detail) < 156 {
			//数据长度不符合要求
			continue
		}

		for k, detail := range box.Data[0].Detail {
			if strings.Contains(detail.Key, "Y下西锅") && strings.Contains(detail.Value, "1") { //西锅，此次信号值为一，下次信号值为0
				if i > 0 {
					//westNext=boxDB[i+1].Data[0].Detail[k].Value
					if boxDB[i-1].Data[0].Detail[k].Value == "0" {
						var info model.FormChargerReport
						info.CreateTime = box.CreateTime
						var sum string
						for _, dataDetail := range box.Data[0].Detail {
							if strings.Contains(dataDetail.Key, "Y称重实际值") {
								info.Data = append(info.Data, model.FormChargerReportDetail{
									OriginKey: dataDetail.Key,
									Value:     dataDetail.Value,
									Unit:      dataDetail.Unit,
								})
								sum = stringAdd(sum, dataDetail.Value)
							}
						}
						info.Crucible = "西锅"
						formChargerReportFormat(&info) //格式处理
						info.Data = append(info.Data, model.FormChargerReportDetail{
							Key:   "总重",
							Value: sum,
							Unit:  "kg",
						})
						//if sum != "0" {
						data = append(data, info)
						//}
					}
				} else if i == 0 { //本次计算的第一条数据
					//查询在100条数据的上一条数据
					sortOpt := options.FindOneOptions{Sort: bson.M{"_id": -1}}
					err = global.FormPlcHisDataColl.FindOne(context.TODO(), bson.M{"createTime": bson.M{"$lt": startTime}}, &sortOpt).Decode(&lastBox)
					if err != nil {
						continue
					}
					if lastBox.Data[0].Detail[k].Value == "0" { //在100条数据的上一条数据
						var info model.FormChargerReport
						info.CreateTime = box.CreateTime
						var sum string
						for _, dataDetail := range box.Data[0].Detail {

							if strings.Contains(dataDetail.Key, "Y称重实际值") {
								info.Data = append(info.Data, model.FormChargerReportDetail{
									OriginKey: dataDetail.Key,
									Value:     dataDetail.Value,
									Unit:      dataDetail.Unit,
								})
								sum = stringAdd(sum, dataDetail.Value)
							}
						}
						info.Crucible = "西锅"
						formChargerReportFormat(&info) //格式处理
						info.Data = append(info.Data, model.FormChargerReportDetail{
							Key:   "总重",
							Value: sum,
							Unit:  "kg",
						})

						//if sum != "0" {
						data = append(data, info)
						//}

					}
				}

			} else if strings.Contains(detail.Key, "Y下东锅") && strings.Contains(detail.Value, "1") { ////东锅，此次信号值为一，下次信号值为0
				if i > 0 {
					if boxDB[i-1].Data[0].Detail[k].Value == "0" {
						var info model.FormChargerReport
						info.CreateTime = box.CreateTime
						var sum string
						for _, dataDetail := range box.Data[0].Detail {
							if strings.Contains(dataDetail.Key, "Y称重实际值") {
								info.Data = append(info.Data, model.FormChargerReportDetail{
									OriginKey: dataDetail.Key,
									Value:     dataDetail.Value,
									Unit:      dataDetail.Unit,
								})
								sum = stringAdd(sum, dataDetail.Value)
							}
						}
						info.Crucible = "东锅"
						formChargerReportFormat(&info) //格式处理
						info.Data = append(info.Data, model.FormChargerReportDetail{
							Key:   "总重",
							Value: sum,
							Unit:  "kg",
						})
						//if sum != "0" {
						data = append(data, info)
						//}
					}
				} else if i == 0 { //本次第一条数据计算
					//查询在100条数据的上一条数据
					sortOpt := options.FindOneOptions{Sort: bson.M{"_id": -1}}
					err = global.FormPlcHisDataColl.FindOne(context.TODO(), bson.M{"createTime": bson.M{"$lt": startTime}}, &sortOpt).Decode(&lastBox)
					if err != nil {
						continue
					}
					if lastBox.Data[0].Detail[k].Value == "0" { //在100条数据的上一条数据
						var info model.FormChargerReport
						info.CreateTime = box.CreateTime
						var sum string
						for _, dataDetail := range box.Data[0].Detail {

							if strings.Contains(dataDetail.Key, "Y称重实际值") {
								info.Data = append(info.Data, model.FormChargerReportDetail{
									OriginKey: dataDetail.Key,
									Value:     dataDetail.Value,
									Unit:      dataDetail.Unit,
								})
								sum = stringAdd(sum, dataDetail.Value)
							}
						}
						info.Crucible = "东锅"
						formChargerReportFormat(&info) //格式处理
						info.Data = append(info.Data, model.FormChargerReportDetail{
							Key:   "总重",
							Value: sum,
							Unit:  "kg",
						})

						//if sum != "0" {
						data = append(data, info)
						//}

					}
				}
			}
		}

	}

	//数据库插入
	for i := range data {
		opts := options.Update().SetUpsert(true)
		result, err := global.FormChargerReportColl.UpdateOne(context.TODO(), bson.M{"createTime": data[i].CreateTime}, bson.D{{"$set", data[i]}}, opts)
		if err != nil {
			fmt.Println("压型配料报表数据库存储失败", err)
		}
		if result.UpsertedCount == 0 && result.MatchedCount == 0 {
			fmt.Println("压型配料报表数据存储失败，数据重复")
			continue
		}
		fmt.Println("压型配料报表数据存储成功")
	}
}

func formChargerReportFormat(info *model.FormChargerReport) {
	for i := range info.Data {
		if strings.Compare(info.Data[i].OriginKey, "Y称重实际值1") == 0 { //左边
			info.Data[i].Key = "2-1l"
		} else if strings.Compare(info.Data[i].OriginKey, "Y称重实际值2") == 0 {
			info.Data[i].Key = "8-4l"
		} else if strings.Compare(info.Data[i].OriginKey, "Y称重实际值3") == 0 {
			info.Data[i].Key = "1-0l"
		} else if strings.Compare(info.Data[i].OriginKey, "Y称重实际值4") == 0 {
			info.Data[i].Key = "4-2l"
		} else if strings.Compare(info.Data[i].OriginKey, "Y称重实际值5") == 0 {
			info.Data[i].Key = "雷蒙粉l"
		} else if strings.Compare(info.Data[i].OriginKey, "Y称重实际值6") == 0 {
			info.Data[i].Key = "生碎l"
		} else if strings.Compare(info.Data[i].OriginKey, "Y称重实际值7") == 0 {
			info.Data[i].Key = "石墨粉l"
		} else if strings.Compare(info.Data[i].OriginKey, "Y称重实际值8") == 0 {
			info.Data[i].Key = "除尘粉l"
		} else if strings.Compare(info.Data[i].OriginKey, "Y称重实际值9") == 0 { //右边
			info.Data[i].Key = "2-1r"
		} else if strings.Compare(info.Data[i].OriginKey, "Y称重实际值10") == 0 {
			info.Data[i].Key = "8-4r"
		} else if strings.Compare(info.Data[i].OriginKey, "Y称重实际值11") == 0 {
			info.Data[i].Key = "1-0r"
		} else if strings.Compare(info.Data[i].OriginKey, "Y称重实际值12") == 0 {
			info.Data[i].Key = "4-2r"
		} else if strings.Compare(info.Data[i].OriginKey, "Y称重实际值13") == 0 {
			info.Data[i].Key = "雷蒙粉r"
		} else if strings.Compare(info.Data[i].OriginKey, "Y称重实际值14") == 0 {
			info.Data[i].Key = "生碎r"
		} else if strings.Compare(info.Data[i].OriginKey, "Y称重实际值15") == 0 {
			info.Data[i].Key = "石墨粉r"
		} else if strings.Compare(info.Data[i].OriginKey, "Y称重实际值16") == 0 {
			info.Data[i].Key = "除尘粉r"
		}
	}
}

//压型配料报表 向前计算至2022-06-14 21:12:11
//从FormChargerEndTime:=“2022-07-04 23:29:00”开始算
//每十分钟向前计算200分钟的数据
//{"createTime":{"$gte":"2022-06-19 17:45:01","$lte":"2022-7-04 21:12:11"}}
func FormChargerReportOperationForward() {

	FormChargerEndTime := "2022-07-23 23:01:00"

	var formChargerEndTime model.FormChargerReportEndTime
	if err := global.FormChargerReportEndTime.FindOne(context.TODO(), bson.M{}).Decode(&formChargerEndTime); err != nil {
		fmt.Println(err)
	}

	startTime := formChargerEndTime.UpdateTime
	if strings.Compare(formChargerEndTime.UpdateTime, formChargerEndTime.EndTime) == 0 {
		fmt.Println("压型配料报表向前计算结束，及时停止该向前计算操作")
		return
	}
	filter := bson.M{
		"createTime": bson.M{
			"$gte": startTime,
			"$lte": FormChargerEndTime,
		},
	}
	opts := options.Find().SetLimit(100)
	var boxDB []model.Box
	res, err := global.FormPlcHisDataColl.Find(context.TODO(), filter, opts)
	if err != nil {
		fmt.Println("成型PLC数据查询失败", err)
		return
	}

	if err = res.All(context.TODO(), &boxDB); err != nil && res == nil {
		fmt.Println(err, "data:", res)
	}

	if boxDB == nil {
		fmt.Println("成型PLC数据目前时间段无历史数据")
		return
	}

	var lastBox model.Box //仅当i=0时使用,i=0的上一条数据
	var data []model.FormChargerReport

	//boxLen := len(boxDB)
	for i, box := range boxDB {
		if len(box.Data[0].Detail) < 156 {
			//数据长度不符合要求
			continue
		}

		for k, detail := range box.Data[0].Detail {
			if strings.Contains(detail.Key, "Y下西锅") && strings.Contains(detail.Value, "1") { //西锅，此次信号值为一，下次信号值为0
				if i > 0 {
					//westNext=boxDB[i+1].Data[0].Detail[k].Value
					if boxDB[i-1].Data[0].Detail[k].Value == "0" {
						var info model.FormChargerReport
						info.CreateTime = box.CreateTime
						var sum string
						for _, dataDetail := range box.Data[0].Detail {
							if strings.Contains(dataDetail.Key, "Y称重实际值") {
								info.Data = append(info.Data, model.FormChargerReportDetail{
									OriginKey: dataDetail.Key,
									Value:     dataDetail.Value,
									Unit:      dataDetail.Unit,
								})
								sum = stringAdd(sum, dataDetail.Value)
							}
						}
						info.Crucible = "西锅"
						formChargerReportFormat(&info) //格式处理
						info.Data = append(info.Data, model.FormChargerReportDetail{
							Key:   "总重",
							Value: sum,
							Unit:  "kg",
						})
						//if sum != "0" {
						data = append(data, info)
						//}
					}
				} else if i == 0 { //本次计算的第一条数据
					//查询在100条数据的上一条数据
					sortOpt := options.FindOneOptions{Sort: bson.M{"_id": -1}}
					err = global.FormPlcHisDataColl.FindOne(context.TODO(), bson.M{"createTime": bson.M{"$lt": startTime}}, &sortOpt).Decode(&lastBox)
					if err != nil {
						continue
					}
					if lastBox.Data[0].Detail[k].Value == "0" { //在100条数据的上一条数据
						var info model.FormChargerReport
						info.CreateTime = box.CreateTime
						var sum string
						for _, dataDetail := range box.Data[0].Detail {
							if strings.Contains(dataDetail.Key, "Y称重实际值") {
								info.Data = append(info.Data, model.FormChargerReportDetail{
									OriginKey: dataDetail.Key,
									Value:     dataDetail.Value,
									Unit:      dataDetail.Unit,
								})
								sum = stringAdd(sum, dataDetail.Value)
							}
						}
						info.Crucible = "西锅"
						formChargerReportFormat(&info) //格式处理
						info.Data = append(info.Data, model.FormChargerReportDetail{
							Key:   "总重",
							Value: sum,
							Unit:  "kg",
						})
						//if sum != "0" {
						data = append(data, info)
						//}
					}
				}

			} else if strings.Contains(detail.Key, "Y下东锅") && strings.Contains(detail.Value, "1") { ////东锅，此次信号值为一，下次信号值为0
				if i > 0 {
					if boxDB[i-1].Data[0].Detail[k].Value == "0" {
						var info model.FormChargerReport
						info.CreateTime = box.CreateTime
						var sum string
						for _, dataDetail := range box.Data[0].Detail {
							if strings.Contains(dataDetail.Key, "Y称重实际值") {
								info.Data = append(info.Data, model.FormChargerReportDetail{
									OriginKey: dataDetail.Key,
									Value:     dataDetail.Value,
									Unit:      dataDetail.Unit,
								})
								sum = stringAdd(sum, dataDetail.Value)
							}
						}
						info.Crucible = "东锅"
						formChargerReportFormat(&info) //格式处理
						info.Data = append(info.Data, model.FormChargerReportDetail{
							Key:   "总重",
							Value: sum,
							Unit:  "kg",
						})
						//if sum != "0" {
						data = append(data, info)
						//}

					}
				} else if i == 0 { //本次第一条数据计算
					//查询在100条数据的上一条数据
					sortOpt := options.FindOneOptions{Sort: bson.M{"_id": -1}}
					err = global.FormPlcHisDataColl.FindOne(context.TODO(), bson.M{"createTime": bson.M{"$lt": startTime}}, &sortOpt).Decode(&lastBox)
					if err != nil {
						continue
					}
					if lastBox.Data[0].Detail[k].Value == "0" { //在100条数据的上一条数据
						var info model.FormChargerReport
						info.CreateTime = box.CreateTime
						var sum string
						for _, dataDetail := range box.Data[0].Detail {
							if strings.Contains(dataDetail.Key, "Y称重实际值") {
								info.Data = append(info.Data, model.FormChargerReportDetail{
									OriginKey: dataDetail.Key,
									Value:     dataDetail.Value,
									Unit:      dataDetail.Unit,
								})
								sum = stringAdd(sum, dataDetail.Value)
							}
						}
						info.Crucible = "东锅"
						formChargerReportFormat(&info) //格式处理
						info.Data = append(info.Data, model.FormChargerReportDetail{
							Key:   "总重",
							Value: sum,
							Unit:  "kg",
						})
						//if sum != "0" {
						data = append(data, info)
						//}

					}
				}
			}
		}

	}

	//向前计算时间更新
	formChargerEndTime.UpdateTime = boxDB[len(boxDB)-1].CreateTime
	if err = global.FormChargerReportEndTime.FindOneAndUpdate(context.TODO(), bson.M{"createTime": "2022-06-14 21:12:11"}, bson.D{{"$set", formChargerEndTime}}).Decode(&bson.M{}); err != nil {
		fmt.Println("压型配料表向前计算时间更新失败", err)
		return
	}

	if data == nil {
		fmt.Println("数据长度符合要求，但没有符合计算压型配料报表的数据")
		formChargerEndTime.UpdateTime = boxDB[len(boxDB)-1].CreateTime
		if err = global.FormChargerReportEndTime.FindOneAndUpdate(context.TODO(), bson.M{"createTime": "2022-06-14 21:12:11"}, bson.D{{"$set", formChargerEndTime}}).Decode(&bson.M{}); err != nil {
			fmt.Println("压型配料表向前计算时间更新失败", err)
			return
		}
		return
	}

	//数据库插入
	for i := range data {
		err = global.FormChargerReportForwardColl.FindOneAndUpdate(context.TODO(), bson.M{"createTime": data[i].CreateTime}, bson.D{{"$set", data[i]}}).Decode(&bson.M{})
		if err != nil {
			_, err = global.FormChargerReportForwardColl.InsertOne(context.TODO(), data[i])
			if err == nil {
				fmt.Println("压型配料报表向前计算数据存储成功")
			}
		}
	}

}
