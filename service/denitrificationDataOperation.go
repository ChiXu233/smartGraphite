package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"SmartGraphite-server/utils"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

func DenitrificationDataOperation(interval time.Duration) {
	endTime := time.Now().Format("2006-01-02 15:04:05")
	startTime := time.Now().Add(-interval).Format("2006-01-02 15:04:05")
	//根据成型PLC类型查询设备
	deviceTypeId, _ := primitive.ObjectIDFromHex("642ab65fab32bd22bd9e1eae")

	var devices []model.Device
	curr, err := global.DeviceColl.Find(context.TODO(), bson.M{"deviceTypeId": deviceTypeId, "isValid": true, "status": "正常"})
	if err != nil {
		log.Println("脱硝设备", err)
		return
	}
	if err := curr.All(context.TODO(), &devices); err != nil {
		log.Println("脱硝设备", err)
		return
	}
	var coll *mongo.Collection
	for _, device := range devices { //设备遍历循环

		switch device.Code {
		//煅烧脱硝
		case "ef62aa2e44204b5d82463b72a86f9621":
			coll = global.CalDenitrificationHis
		//焙烧脱硝
		case "52980204e2dc4ce9907196441c6f9a32":
			coll = global.RoastDenitrificationHis
		default:
			fmt.Println("脱硝" + "时间粒度计算设备编号设置错误")
			return
		}

		//查询对应设备的数据
		var boxDB []model.Box
		filter := bson.M{
			"createTime": bson.M{
				"$gte": startTime,
				"$lte": endTime,
			},
			"boxId": device.Code,
		}
		//最新的一条数据的信号值作为最后的信号值
		opts := options.FindOptions{
			Sort: bson.M{"_id": -1},
		}
		//根据filter条件查询历史表
		curr, err = coll.Find(context.TODO(), filter, &opts)
		if err != nil {
			log.Println("脱硝设备历史数据", err)
			return
		}
		if err = curr.All(context.TODO(), &boxDB); err != nil {
			log.Println("脱硝设备历史数据", err)
			return
		}
		//fmt.Println("test", 1111)
		if boxDB == nil {
			log.Println(device.Code, "此设备在此时间段没有数据")
			return
		}

		//数据计算
		var info model.Box
		temp := len(boxDB) - 1
		for i := range boxDB {
			info.CreateTime = endTime
			if i == 0 { //数据格式初始化，数据初始化
				info.DeviceTypeId = boxDB[i].DeviceTypeId
				info.BoxId = boxDB[i].BoxId
				info.Data = append(info.Data, model.BoxData{})
				//详细数据格式处理 区分信号值和其他可求最小，最大，平均的数据项
				for k, detail := range boxDB[i].Data[0].Detail {
					fmt.Println(detail.Value, "value")
					//if detail.Value == "0" || detail.Value == "1" {
					//	info.Data[0].Detail = append(info.Data[0].Detail, model.BoxDataDetail{
					//		Key:   detail.Key,
					//		Value: detail.Value,
					//		Unit:  detail.Unit,
					//	})
					//} else {
					info.Data[0].Detail = append(info.Data[0].Detail, []model.BoxDataDetail{
						{
							Key:   boxDB[i].Data[0].Detail[k].Key + "最大值",
							Value: boxDB[i].Data[0].Detail[k].Value,
							Unit:  boxDB[i].Data[0].Detail[k].Unit,
						},
						{
							Key:   boxDB[i].Data[0].Detail[k].Key + "最小值",
							Value: boxDB[i].Data[0].Detail[k].Value,
							Unit:  boxDB[i].Data[0].Detail[k].Unit,
						},
						{
							Key:   boxDB[i].Data[0].Detail[k].Key + "平均值",
							Value: boxDB[i].Data[0].Detail[k].Value,
							Unit:  boxDB[i].Data[0].Detail[k].Unit,
						},
					}...)
				}

			} else { //数据计算
				dataLen := len(boxDB[i].Data[0].Detail)
				count := 0 //用于标记经过了多少个可求最小，最大，平均的数据项
				for k := 0; k < dataLen; k++ {
					j := 2*count + k //info中detail的key值对应
					if boxDB[i].Data[0].Detail[k].Value == "0" || boxDB[i].Data[0].Detail[k].Value == "1" {
						//目前不用计算 信号计算
						//info.Data[0].Detail[k].Value = stringAdd(info.Data[0].Detail[k].Value, boxDB[i].Data[0].Detail[k].Value)
					} else { //其他可求最小，最大，平均的数据项计算
						if j+2 < len(info.Data[0].Detail) {
							info.Data[0].Detail[j].Value = stringMax(info.Data[0].Detail[j].Value, boxDB[i].Data[0].Detail[k].Value)
							info.Data[0].Detail[j+1].Value = stringMin(info.Data[0].Detail[j+1].Value, boxDB[i].Data[0].Detail[k].Value)
							info.Data[0].Detail[j+2].Value = stringAdd(info.Data[0].Detail[j+2].Value, boxDB[i].Data[0].Detail[k].Value)
							count++
							if i == temp { //最后一次累加，其他可求最小，最大，平均的数据项除以累加的次数
								info.Data[0].Detail[j+2].Value = DivisionTen(info.Data[0].Detail[j+2].Value, temp+1)
							}
						}
					}
				}
			}
		}
		denitrificationColl(info, interval)
	}
}

func denitrificationColl(box model.Box, interval time.Duration) {
	var coll *mongo.Collection
	var cOll *mongo.Collection
	var msg string
	switch interval {
	case 10 * time.Minute:
		coll = global.CalDenitrificationTen
		cOll = global.RoastDenitrificationTen
		msg = "10分钟"
	case 30 * time.Minute:
		coll = global.CalDenitrificationThirty
		cOll = global.RoastDenitrificationThirty
		msg = "30分钟"
	case time.Hour:
		coll = global.CalDenitrificationHour
		cOll = global.RoastDenitrificationHour
		msg = "小时"
	default:
		fmt.Println("脱硝设备时间粒度选择错误")
		return
	}
	_, err := coll.InsertOne(context.TODO(), box)
	if err != nil {
		fmt.Println("脱硝设备" + msg + "存储失败")
		return
	}
	_, erre := cOll.InsertOne(context.TODO(), box)
	if erre != nil {
		fmt.Println("焙烧脱销设备" + msg + "存储失败")
		return
	}
	fmt.Println("脱硝设备" + msg + "存储成功")
	return
}

// 煅烧脱硝10分钟写入NOx反馈值
func WriteDenitrificationDataOperation() {
	////与整点相差多少分。相差多少sleep多少
	//nowtimeMinut := time.Now().Minute()
	//if nowtimeMinut%10 != 0 {
	//	sleepTime := 60 * (10 - nowtimeMinut%10)
	//	time.Sleep(time.Duration(sleepTime) * time.Second)
	//}
	////找出焙烧CEMS中最新10条数据
	//endTime := time.Now().Format("2006-01-02 15:04:05")
	//startTime := time.Now().Add(-interval).Format("2006-01-02 15:04:05")
	//code := "130424LRTTS001"
	//var limit int64 = 10
	//filters := bson.M{
	//	"deviceCode": code,
	//	"createTime": bson.M{
	//		//time要为string类型
	//		"$gte": startTime,
	//		"$lte": endTime,
	//	},
	//}
	//findOptions := options.FindOptions{
	//	Limit: &limit,
	//	Sort:  bson.M{"_id": -1},
	//}
	//res, err := global.DataMinuteHisColl.Find(context.TODO(), filters, &findOptions)
	////若此时间段内没有十分钟数据
	//if err == mongo.ErrNoDocuments {
	//	variantVal := "0"
	//	writeWriteDenitrificationData(variantVal)
	//}
	//if err != nil {
	//	fmt.Println("获取10条数据失败", err.Error())
	//	return
	//}
	//var data []model.DeviceData
	//err = res.All(context.TODO(), &data)
	//if err != nil {
	//	fmt.Println("获取前10条数据失败", err.Error())
	//}
	////拿出其中的NOx值并累加计算平均值
	//var value float64
	//var sum float64
	//var avg float64
	//for i := 0; i < len(data); i++ {
	//	//累加和sum
	//	value, err = strconv.ParseFloat(data[i].Dataset[42].Value, 64)
	//	if err != nil {
	//		fmt.Println("类型转化出错", err.Error())
	//		return
	//	}
	//	sum += value
	//}
	//avg = sum / 10

	//分钟表查找每分钟氮氧化物最大值并写入
	var data model.DeviceData
	if err := global.DataMinuteHisColl.FindOne(context.TODO(), bson.M{"deviceCode": "130424LRTTS001"}, &options.FindOneOptions{
		Sort: bson.M{"_id": -1},
	}).Decode(&data); err != nil {
		fmt.Println("查找最新一条分钟值失败", err.Error())
		return
	}
	variantVal := data.Dataset[40].Value
	writeWriteDenitrificationData(variantVal)
	writeWriteDenitrificationData2(variantVal)
}

// 煅烧脱硝
func writeWriteDenitrificationData(variantVal string) {
	URL := "http://sukon-cloud.com/api/v1/mndraw/setVariantValue"
	urlValues := url.Values{}
	urlValues.Add("token", global.Token)
	urlValues.Add("variantVal", variantVal)
	urlValues.Add("variantId", "ef62aa2e44204b5d82463b72a86f9621:1")
	res, err := http.PostForm(URL, urlValues)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	var data model.WriteData
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println(err)
		return
	}
	if data.Msg == "token已过期" {
		fmt.Println("project数据为空,获取失败---", data)
		utils.FindToken()
		return
	}
	fmt.Println(data.Data, "返回值")
}

// 焙烧脱硝
func writeWriteDenitrificationData2(variantVal string) {
	URL := "http://sukon-cloud.com/api/v1/mndraw/setVariantValue"
	urlValues := url.Values{}
	urlValues.Add("token", global.Token)
	urlValues.Add("variantVal", variantVal)
	urlValues.Add("variantId", "52980204e2dc4ce9907196441c6f9a32:1")
	res, err := http.PostForm(URL, urlValues)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	var data model.WriteData
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println(err)
		return
	}
	if data.Msg == "token已过期" {
		fmt.Println("project数据为空,获取失败---", data)
		utils.FindToken()
		return
	}
	fmt.Println(data.Data, "返回值")
}
