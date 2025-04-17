package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"SmartGraphite-server/utils"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// 获取项目
func FindProjects() {
	URL := "http://sukon-cloud.com/api/v1/base/projects"
	urlValues := url.Values{}
	urlValues.Add("token", global.Token)
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
	var data model.Project
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println(err)
		return
	}
	if data.Data == nil && data.Msg == "token已过期" {
		fmt.Println("project数据为空,获取失败---", data)
		utils.FindToken()
		return
	}
	fmt.Println(data.Data)
	for _, project := range data.Data {
		if project.Id == "rKWw9LNBQYH" {
			continue
		}
		FindProjectBoxes(global.Token, project.Id)
	}
}

// 获取Box
func FindProjectBoxes(token, projectId string) {
	URL := "http://sukon-cloud.com/api/v1/base/projectBoxes"
	urlValues := url.Values{}
	urlValues.Add("token", token)
	urlValues.Add("projectId", projectId)
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
	var data model.ProjectBox
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println(err)
		return
	}
	if data.Success == false {
		fmt.Println("获取box异常", data)
	}
	for _, Box := range data.Data {
		if Box.BoxId == "3d34d1b2385c4eafb94ffe89cfd6f43d" {
			continue
		}
		//设备不在线更改设备状态
		if Box.Status == "0" {
			updateTime := time.Now().Format("2006-01-02 15:04:05")
			update := bson.M{"$set": bson.M{"status": "离线", "updateTime": updateTime}}
			err := global.DeviceColl.FindOneAndUpdate(context.TODO(), bson.M{"code": Box.BoxId}, update).Decode(bson.M{})
			if err != nil {
				fmt.Println(err)
				continue
			}
			//fmt.Println(Box.Name + "设备离线")
			continue
		} else {
			updateTime := time.Now().Format("2006-01-02 15:04:05")
			update := bson.M{"$set": bson.M{"status": "正常", "updateTime": updateTime}}
			err := global.DeviceColl.FindOneAndUpdate(context.TODO(), bson.M{"code": Box.BoxId}, update).Decode(bson.M{})
			if err != nil {
				fmt.Println(err)
				continue
			}
			//fmt.Println(Box.Name + "设备在线")
		}
		FindBoxPlc(token, Box.BoxId)
	}
}

// 获取plc
func FindBoxPlc(token, boxId string) {
	URL := "http://sukon-cloud.com/api/v1/base/boxPlcs"
	urlValues := url.Values{}
	urlValues.Add("token", token)
	urlValues.Add("boxId", boxId)
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
	var data model.BoxPlc
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println(err)
		return
	}
	urlValues.Del("boxId")
	if data.Data == nil {
		fmt.Println("plcId为空", data)
		return
	}
	fmt.Println(data.Data)
	for _, a := range data.Data {
		FindVariant(token, boxId, a.PlcId)
	}
}

// 获取变量
func FindVariant(token, boxId, plcId string) {
	URL := "http://sukon-cloud.com/api/v1/base/boxVariants"
	urlValues := url.Values{}
	urlValues.Add("token", token)
	urlValues.Add("boxId", boxId)
	//获取每个sid下变量
	var data model.BoxVariant
	fmt.Println(plcId, "plcID")
	urlValues.Add("plcId", plcId)
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
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println(err)
		return
	}
	urlValues.Del("boxId")
	urlValues.Del("plcId")
	if data.Success == false {
		fmt.Println("获取变量失败")
		return
	}
	//得到用于获取实时数据的变量字符串
	var variantIds string
	for i, variant := range data.Data {
		if len(data.Data) == 1 {
			variantIds = boxId + "(" + variant.VariantId + ")"
		} else {
			if i == len(data.Data)-1 {
				variantIds = variantIds + variant.VariantId + ")"
			} else {
				if i == 0 {
					variantIds = variantIds + boxId + "("
				}
				variantIds = variantIds + variant.VariantId + ":"
			}
		}
	}
	//box数据
	var box model.Box
	box.BoxId = boxId
	var detail []model.BoxDataDetail
	for _, a := range data.Data {
		detail = append(detail, model.BoxDataDetail{
			Key:   a.Name,
			Value: "",
			Unit:  "",
		})
	}
	box.Data = append(box.Data, model.BoxData{
		SensorId:   "",
		SensorName: "",
		Detail:     detail,
	})
	//获取实时数据
	FindRealTimeData(token, variantIds, box)
}

// 获取实时数据
func FindRealTimeData(token, variantIds string, box model.Box) {
	var data model.RealtimeData
	URL := "http://sukon-cloud.com/api/v1/data/realtimeDatas"
	urlValues := url.Values{}
	urlValues.Add("token", token)
	urlValues.Add("variantIds", variantIds)
	//fmt.Println(URL, "URL")
	//fmt.Println(urlValues, "urlValues")
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
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println(err)
		return
	}
	urlValues.Del("variantIds")
	////浸渍
	//if box.BoxId == "2da580adb26b4a12accd4aec80e04656" {
	//	storeDippingData(box, data)
	//}
	////西跨吸料天车
	//if box.BoxId == "b46a0faf11cc4000a4c290eba5cc949a" {
	//	storeWestAirCarData(box, data)
	//}
	////东跨吸料天车
	//if box.BoxId == "f73fe0d8688046e088bb073849aa0c3f" {
	//	storeEarthAirCarData(box, data)
	//}
	////隧道窑
	//if box.BoxId == "9f62bc0edbd542b2bec159ac8f023509" {
	//	storeTunnelData(box, data)
	//}
	//石墨化
	if box.BoxId == "be67c2b8216e49e8981a95663413f115" {
		storeGraphitingData(box, data)
	}
	////坩埚
	//if box.BoxId == "9bd62f734af94dc0b0641817ac2807e9" {
	//	storeCrucibleData(box, data)
	//}
	////隧道窑湿电
	//if box.BoxId == "5cba298477bc456ab1a2bd06e35cb0d8" {
	//	storeTunnelWetElectricData(box, data)
	//}
	////焙烧湿电
	//if box.BoxId == "01e844f884844aa2bb5d1cab87316c17" {
	//	storeRoastingWetElectricData(box, data)
	//}
	////石墨化湿电
	//if box.BoxId == "69fb82a9cba744188cab9da766787f25" {
	//	storeGraphiteWetElectricData(box, data)
	//}
}

// 存储浸渍分钟数据
func storeDippingData(box model.Box, data model.RealtimeData) {
	//查找设备
	var BoxDevice model.Device
	if err := global.DeviceColl.FindOne(context.TODO(), bson.M{"code": box.BoxId}).Decode(&BoxDevice); err != nil {
		log.Println(box.BoxId+"设备不存在", err)
		return
	}
	box.DeviceTypeId = BoxDevice.DeviceTypeId
	box.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	//处理单位
	if len(BoxDevice.Sensors) == 1 {
		box.Data[0].SensorId = BoxDevice.Sensors[0].Code
		box.Data[0].SensorName = BoxDevice.Sensors[0].Name
		for i, a := range box.Data[0].Detail {
			for _, b := range BoxDevice.Sensors[0].DetectionValue {
				if a.Key == b.Key {
					box.Data[0].Detail[i].Unit = b.Unit
					continue
				}
			}
		}
	}
	//处理值
	for i, _ := range box.Data[0].Detail {
		for _, b := range data.Data {
			array := strings.Split(b.Id, "2da580adb26b4a12accd4aec80e04656:")
			id := strconv.Itoa(i)
			if array[1] == id {
				box.Data[0].Detail[i].Value = b.Value
				continue
			}
		}
	}
	//历史表插入
	_, err := global.DipDataHisColl.InsertOne(context.Background(), box)
	if err != nil {
		fmt.Println(err)
	}
	//更新数据更新表
	box.UpdateTime = box.CreateTime
	box.CreateTime = ""
	err = global.DipDataColl.FindOneAndUpdate(context.TODO(), bson.M{"boxId": box.BoxId}, bson.M{"$set": box}).Decode(bson.M{})
	if err == mongo.ErrNoDocuments {
		//第一次无记录更新，需要插入新纪录
		box.CreateTime = box.UpdateTime
		if res, err := global.DipDataColl.InsertOne(context.TODO(), box); err != nil {
			fmt.Println("浸渍数据更新存储出错", err.Error())
		} else {
			fmt.Println("浸渍更新表创建新纪录", res.InsertedID)
		}
	}
	fmt.Println("浸渍分钟数据更新存储完毕")
	return
}

// 存储西跨吸料天车分钟数据
func storeWestAirCarData(box model.Box, data model.RealtimeData) {
	//查找设备
	var BoxDevice model.Device
	if err := global.DeviceColl.FindOne(context.TODO(), bson.M{"code": box.BoxId}).Decode(&BoxDevice); err != nil {
		log.Println(box.BoxId+"设备不存在", err)
		return
	}
	box.DeviceTypeId = BoxDevice.DeviceTypeId
	box.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	//处理单位
	if len(BoxDevice.Sensors) == 1 {
		box.Data[0].SensorId = BoxDevice.Sensors[0].Code
		box.Data[0].SensorName = BoxDevice.Sensors[0].Name
		for i, a := range box.Data[0].Detail {
			for _, b := range BoxDevice.Sensors[0].DetectionValue {
				if a.Key == b.Key {
					box.Data[0].Detail[i].Unit = b.Unit
					continue
				}
			}
		}
	}
	//处理值
	for i, _ := range box.Data[0].Detail {
		for _, b := range data.Data {
			array := strings.Split(b.Id, "b46a0faf11cc4000a4c290eba5cc949a:")
			id := strconv.Itoa(i)
			if array[1] == id {
				box.Data[0].Detail[i].Value = b.Value
				continue
			}
		}
	}
	//历史表插入
	_, err := global.WestAirCarHisDataColl.InsertOne(context.Background(), box)
	if err != nil {
		fmt.Println(err)
	}
	//更新数据更新表
	box.UpdateTime = box.CreateTime
	box.CreateTime = ""
	err = global.WestAirCarDataColl.FindOneAndUpdate(context.TODO(), bson.M{"boxId": box.BoxId}, bson.M{"$set": box}).Decode(bson.M{})
	if err == mongo.ErrNoDocuments {
		//第一次无记录更新，需要插入新纪录
		box.CreateTime = box.UpdateTime
		if res, err := global.WestAirCarDataColl.InsertOne(context.TODO(), box); err != nil {
			fmt.Println("西跨吸料天车数据更新存储出错", err.Error())
		} else {
			fmt.Println("西跨吸料天车更新表创建新纪录", res.InsertedID)
		}
	}
	fmt.Println("西跨吸料天车分钟数据更新存储完毕")
	return
}

// 存储东跨跨吸料天车分钟数据
func storeEarthAirCarData(box model.Box, data model.RealtimeData) {
	//查找设备
	var BoxDevice model.Device
	if err := global.DeviceColl.FindOne(context.TODO(), bson.M{"code": box.BoxId}).Decode(&BoxDevice); err != nil {
		log.Println(box.BoxId+"设备不存在", err)
		return
	}
	box.DeviceTypeId = BoxDevice.DeviceTypeId
	box.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	//处理单位
	if len(BoxDevice.Sensors) == 1 {
		box.Data[0].SensorId = BoxDevice.Sensors[0].Code
		box.Data[0].SensorName = BoxDevice.Sensors[0].Name
		for i, a := range box.Data[0].Detail {
			for _, b := range BoxDevice.Sensors[0].DetectionValue {
				if a.Key == b.Key {
					box.Data[0].Detail[i].Unit = b.Unit
					continue
				}
			}
		}
	}
	//处理值
	for i, _ := range box.Data[0].Detail {
		for _, b := range data.Data {
			array := strings.Split(b.Id, "f73fe0d8688046e088bb073849aa0c3f:")
			id := strconv.Itoa(i)
			if array[1] == id {
				box.Data[0].Detail[i].Value = b.Value
				continue
			}
		}
	}
	//历史表插入
	_, err := global.EastAirCarHisDataColl.InsertOne(context.Background(), box)
	if err != nil {
		fmt.Println(err)
	}
	//更新数据更新表
	box.UpdateTime = box.CreateTime
	box.CreateTime = ""
	err = global.EastAirCarDataColl.FindOneAndUpdate(context.TODO(), bson.M{"boxId": box.BoxId}, bson.M{"$set": box}).Decode(bson.M{})
	if err == mongo.ErrNoDocuments {
		//第一次无记录更新，需要插入新纪录
		box.CreateTime = box.UpdateTime
		if res, err := global.EastAirCarDataColl.InsertOne(context.TODO(), box); err != nil {
			fmt.Println("东跨吸料天车数据更新存储出错", err.Error())
		} else {
			fmt.Println("东跨吸料天车更新表创建新纪录", res.InsertedID)
		}
	}
	fmt.Println("东跨吸料天车分钟数据更新存储完毕")
	return
}

// 存储隧道窑分钟数据
func storeTunnelData(box model.Box, data model.RealtimeData) {
	//查找设备
	var BoxDevice model.Device
	if err := global.DeviceColl.FindOne(context.TODO(), bson.M{"code": box.BoxId}).Decode(&BoxDevice); err != nil {
		log.Println(box.BoxId+"设备不存在", err)
		return
	}
	box.DeviceTypeId = BoxDevice.DeviceTypeId
	box.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	//处理单位
	if len(BoxDevice.Sensors) == 1 {
		box.Data[0].SensorId = BoxDevice.Sensors[0].Code
		box.Data[0].SensorName = BoxDevice.Sensors[0].Name
		for i, a := range box.Data[0].Detail {
			for _, b := range BoxDevice.Sensors[0].DetectionValue {
				if a.Key == b.Key {
					box.Data[0].Detail[i].Unit = b.Unit
					continue
				}
			}
		}
	}
	//处理值
	for i, _ := range box.Data[0].Detail {
		for _, b := range data.Data {
			array := strings.Split(b.Id, "9f62bc0edbd542b2bec159ac8f023509:")
			id := strconv.Itoa(i)
			if array[1] == id {
				box.Data[0].Detail[i].Value = b.Value
				continue
			}
		}
	}
	//历史表插入
	if _, err := global.TunnelHisDataColl.InsertOne(context.TODO(), box); err != nil {
		fmt.Println("隧道窑数据存储出错", err.Error())
	}
	box.UpdateTime = box.CreateTime
	box.CreateTime = ""
	//更新数据更新表
	err := global.TunnelDataColl.FindOneAndUpdate(context.TODO(), bson.M{"boxId": box.BoxId}, bson.M{"$set": box}).Decode(bson.M{})
	if err == mongo.ErrNoDocuments {
		//第一次无记录更新，需要插入新纪录
		box.CreateTime = box.UpdateTime
		if res, err := global.TunnelDataColl.InsertOne(context.TODO(), box); err != nil {
			fmt.Println("隧道窑数据更新存储出错", err.Error())
		} else {
			fmt.Println("隧道窑更新表创建新纪录", res.InsertedID)
		}
	}
	fmt.Println("隧道窑分钟数据更新存储完毕")
	return
}

// 存储石墨化分钟数据
func storeGraphitingData(box model.Box, data model.RealtimeData) {
	//查找设备
	var BoxDevice model.Device
	if err := global.DeviceColl.FindOne(context.TODO(), bson.M{"code": box.BoxId}).Decode(&BoxDevice); err != nil {
		log.Println(box.BoxId+"设备不存在", err)
		return
	}
	box.DeviceTypeId = BoxDevice.DeviceTypeId
	box.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	//处理单位
	if len(BoxDevice.Sensors) == 1 {
		box.Data[0].SensorId = BoxDevice.Sensors[0].Code
		box.Data[0].SensorName = BoxDevice.Sensors[0].Name
		for i, a := range box.Data[0].Detail {
			for _, b := range BoxDevice.Sensors[0].DetectionValue {
				if a.Key == b.Key {
					box.Data[0].Detail[i].Unit = b.Unit
					continue
				}
			}
		}
	}
	//处理值
	var hour string
	var day string
	var minute string
	var month string
	var year string
	var power string        //有功功率
	var electricity string  //有功电量
	var displacement string //位移对应的小位移
	for i, _ := range box.Data[0].Detail {
		for _, b := range data.Data {
			array := strings.Split(b.Id, "be67c2b8216e49e8981a95663413f115:") //速控云上对应每个变量的id
			id := strconv.Itoa(i)
			if array[1] == "29" {
				intVal, _ := strconv.Atoi(b.Value)
				if intVal < 10 {
					hour = "0" + b.Value
				} else {
					hour = b.Value
				}
			}
			if array[1] == "3" {
				power = b.Value
			}
			if array[1] == "4" {
				electricity = b.Value
			}
			if array[1] == "30" {
				intVal, _ := strconv.Atoi(b.Value)
				if intVal < 10 && intVal != 0 {
					day = "0" + b.Value
				} else {
					day = b.Value
				}
			}
			if array[1] == "31" {
				intVal, _ := strconv.Atoi(b.Value)
				if intVal < 10 {
					minute = "0" + b.Value
				} else {
					minute = b.Value
				}
			}
			if array[1] == "32" {
				intVal, _ := strconv.Atoi(b.Value)
				if intVal < 10 && intVal != 0 {
					month = "0" + b.Value
				} else {
					month = b.Value
				}
			}
			if array[1] == "45" {
				year = b.Value
			}
			if array[1] == "47" {
				//石墨化炉号
				box.Data[0].StoveNumber = b.Value
			}
			//如果分割后id和设备表上对应设备的id一样，说明他们是一个变量，然后赋值
			if array[1] == id {
				box.Data[0].Detail[i].Value = b.Value
				continue
			}
		}
	}

	//根据位移对照表将大位移换算为小位移
	var excel model.GraphitingDisplacementExcel
	if err := global.GraphitingdisplacementColl.FindOne(context.TODO(), bson.M{}).Decode(&excel); err != nil {
		fmt.Println("石墨化PLC位移对照失败", err.Error())
	}
	for x := range excel.ExcelData {
		if box.Data[0].Detail[15].Value == excel.ExcelData[x].Key {
			displacement = excel.ExcelData[x].Value
			break
		}
	}

	//若未匹配到对应值，使用公式计算 y=x*0.4555+0.2101
	if displacement == "" {
		x, _ := strconv.ParseFloat(box.Data[0].Detail[15].Value, 10)
		displacement = fmt.Sprintf("%.1f", 0.4555*x+0.2101)
	}
	box.Data[0].Detail = append(box.Data[0].Detail, model.BoxDataDetail{
		Key:   "小位移",
		Value: displacement,
		Unit:  "mm",
	})
	//停电、送电Mqtt发布消息
	//GraphitePubAlarm(BoxDevice, power)
	//石墨化送电时刻
	sendEleTime := year + "-" + month + "-" + day + " " + hour + ":" + minute + ":00"
	box.Data[0].StartTime = sendEleTime
	//当前有功电量小于1
	intElectricity, _ := strconv.Atoi(electricity)
	if intElectricity < 1 {
		//获取上一个有功电量
		var lastData model.Box
		_ = global.GraphitingDataColl.FindOne(context.TODO(), bson.M{}).Decode(&lastData)
		electricity2 := lastData.Data[0].Detail[4].Value
		intElectricity2, _ := strconv.Atoi(electricity2)
		//前一个有功电量大于1则记录当前电量时刻值，同时存储报表
		if intElectricity2 > 1 {
			var nowElectric model.GraElectric
			nowElectric.CreateTime = box.CreateTime
			nowElectric.StoveNumber = box.Data[0].StoveNumber
			//获取上一个电量时刻数据时间 存储报表
			var lastElectric []model.GraElectric
			findOptions := new(options.FindOptions)
			findOptions = &options.FindOptions{}
			findOptions.SetSort(bson.D{{"createTime", -1}})
			findOptions.SetLimit(1) //每页数据数量
			cur, err := global.GraElectricTimeColl.Find(context.TODO(), bson.M{}, findOptions)
			if err != nil {
				fmt.Println(err)
			}
			if err = cur.All(context.TODO(), &lastElectric); err != nil {
				fmt.Println(err)
			}
			storeGraReportByElectricity(lastElectric[0].CreateTime, nowElectric.CreateTime)
			//存储时刻数据
			_, err = global.GraElectricTimeColl.InsertOne(context.TODO(), nowElectric)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	//历史表插入
	//_, err := global.GraphitingHisDataColl.InsertOne(context.Background(), box)
	//if err != nil {
	//	fmt.Println(err)
	//}
	////更新数据更新表
	//box.UpdateTime = box.CreateTime
	//box.CreateTime = ""
	//err = global.GraphitingDataColl.FindOneAndUpdate(context.TODO(), bson.M{"boxId": box.BoxId}, bson.M{"$set": box}).Decode(bson.M{})
	//if err == mongo.ErrNoDocuments {
	//	//第一次无记录更新，需要插入新纪录
	//	box.CreateTime = box.UpdateTime
	//	if res, err := global.GraphitingDataColl.InsertOne(context.TODO(), box); err != nil {
	//		fmt.Println("数据更新存储出错", err.Error())
	//	} else {
	//		fmt.Println("石墨化数据更新表创建新纪录", res.InsertedID)
	//	}
	//}
	fmt.Println("石墨化分钟数据更新存储完毕")
	//有功功率小于1的话计算结束时间然后创建石墨化报表数据
	intPower, _ := strconv.Atoi(power)
	if intPower < 1 && sendEleTime != "0-0-0 00:00:00" {
		//添加有功功率对应的时刻值
		var graPower model.GraPower
		graPower.CreateTime = box.UpdateTime
		graPower.StartTime = box.Data[0].StartTime
		graPower.StoveNumber = box.Data[0].StoveNumber
		_, err := global.GraPowerTimeColl.InsertOne(context.TODO(), graPower)
		if err != nil {
			fmt.Println(err)
		}
		storeGraReportData(box)
	}

	//存储5分钟数据
	utils.Try(func() {
		storeGraFiveData(box)
	})

	return
}

// 存储石墨化5分钟数据
func storeGraFiveData(box model.Box) {

	//判断分钟时间是否5的倍数
	createTime, err := strconv.Atoi(box.UpdateTime[14:16])
	if err != nil {
		return
	}
	if (createTime % 5) != 0 { //不是5的倍数
		return
	}
	box.CreateTime = box.UpdateTime
	box.UpdateTime = ""
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
	//数据表插入
	_, err = global.GraphitingFiveDataColl.InsertOne(context.TODO(), box)
	if err != nil {
		fmt.Println("石墨化5分钟数据创建失败")
	}
	fmt.Println("石墨化5分钟数据创建成功")

	return
}

// 存储坩埚分钟数据
func storeCrucibleData(box model.Box, data model.RealtimeData) {
	//查找设备
	var BoxDevice model.Device
	if err := global.DeviceColl.FindOne(context.TODO(), bson.M{"code": box.BoxId}).Decode(&BoxDevice); err != nil {
		log.Println(box.BoxId+"设备不存在", err)
		return
	}
	box.DeviceTypeId = BoxDevice.DeviceTypeId
	box.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	//处理单位
	if len(BoxDevice.Sensors) == 1 {
		box.Data[0].SensorId = BoxDevice.Sensors[0].Code
		box.Data[0].SensorName = BoxDevice.Sensors[0].Name
		for i, a := range box.Data[0].Detail {
			for _, b := range BoxDevice.Sensors[0].DetectionValue {
				if a.Key == b.Key {
					box.Data[0].Detail[i].Unit = b.Unit
					continue
				}
			}
		}
	}
	//处理值
	for i, _ := range box.Data[0].Detail {
		for _, b := range data.Data {
			array := strings.Split(b.Id, "9bd62f734af94dc0b0641817ac2807e9:")
			id := strconv.Itoa(i)
			if array[1] == id {
				box.Data[0].Detail[i].Value = b.Value
				continue
			}
		}
	}
	//历史表插入
	if _, err := global.CrucibleHisDataColl.InsertOne(context.TODO(), box); err != nil {
		fmt.Println("坩埚数据存储出错", err.Error())
	}
	box.UpdateTime = box.CreateTime
	box.CreateTime = ""
	//更新数据更新表
	err := global.CrucibleDataColl.FindOneAndUpdate(context.TODO(), bson.M{"boxId": box.BoxId}, bson.M{"$set": box}).Decode(bson.M{})
	if err == mongo.ErrNoDocuments {
		//第一次无记录更新，需要插入新纪录
		box.CreateTime = box.UpdateTime
		if res, err := global.CrucibleDataColl.InsertOne(context.TODO(), box); err != nil {
			fmt.Println("坩埚数据更新存储出错", err.Error())
		} else {
			fmt.Println("坩埚更新表创建新纪录", res.InsertedID)
		}
	}
	fmt.Println("坩埚分钟数据更新存储完毕")
	return
}

// 存储隧道窑湿电数据
func storeTunnelWetElectricData(box model.Box, data model.RealtimeData) {
	//查找设备
	var BoxDevice model.Device
	if err := global.DeviceColl.FindOne(context.TODO(), bson.M{"code": box.BoxId}).Decode(&BoxDevice); err != nil {
		log.Println(box.BoxId+"设备不存在", err)
		return
	}
	box.DeviceTypeId = BoxDevice.DeviceTypeId
	box.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	//处理单位
	if len(BoxDevice.Sensors) == 1 {
		box.Data[0].SensorId = BoxDevice.Sensors[0].Code
		box.Data[0].SensorName = BoxDevice.Sensors[0].Name
		for i, a := range box.Data[0].Detail {
			for _, b := range BoxDevice.Sensors[0].DetectionValue {
				if a.Key == b.Key {
					box.Data[0].Detail[i].Unit = b.Unit
					continue
				}
			}
		}
	}
	//处理值
	for i, _ := range box.Data[0].Detail {
		for _, b := range data.Data {
			array := strings.Split(b.Id, "5cba298477bc456ab1a2bd06e35cb0d8:") //速控云上对应每个变量的id
			id := strconv.Itoa(i)
			//如果分割后id和设备表上对应设备的id一样，说明他们是统统一个变量，然后赋值
			if array[1] == id {
				box.Data[0].Detail[i].Value = b.Value
				continue
			}
		}
	}
	//历史表插入
	_, err := global.TunWetElectricHisDataColl.InsertOne(context.Background(), box)
	if err != nil {
		fmt.Println(err)
	}
	//更新数据更新表
	box.UpdateTime = box.CreateTime
	box.CreateTime = ""
	err = global.TunWetElectricDataColl.FindOneAndUpdate(context.TODO(), bson.M{"boxId": box.BoxId}, bson.M{"$set": box}).Decode(bson.M{})
	if err == mongo.ErrNoDocuments {
		//第一次无记录更新，需要插入新纪录
		box.CreateTime = box.UpdateTime
		if res, err := global.TunWetElectricDataColl.InsertOne(context.TODO(), box); err != nil {
			fmt.Println("隧道窑湿电数据更新存储出错", err.Error())
		} else {
			fmt.Println("隧道窑湿电数据更新表创建新纪录", res.InsertedID)
		}
	}
	fmt.Println("隧道窑湿电分钟数据更新存储完毕")
	return
}

// 存储焙烧湿电数据
func storeRoastingWetElectricData(box model.Box, data model.RealtimeData) {
	//查找设备
	var BoxDevice model.Device
	if err := global.DeviceColl.FindOne(context.TODO(), bson.M{"code": box.BoxId}).Decode(&BoxDevice); err != nil {
		log.Println(box.BoxId+"设备不存在", err)
		return
	}
	box.DeviceTypeId = BoxDevice.DeviceTypeId
	box.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	//处理单位
	if len(BoxDevice.Sensors) == 1 {
		box.Data[0].SensorId = BoxDevice.Sensors[0].Code
		box.Data[0].SensorName = BoxDevice.Sensors[0].Name
		for i, a := range box.Data[0].Detail {
			for _, b := range BoxDevice.Sensors[0].DetectionValue {
				if a.Key == b.Key {
					box.Data[0].Detail[i].Unit = b.Unit
					continue
				}
			}
		}
	}
	//处理值
	for i, _ := range box.Data[0].Detail {
		for _, b := range data.Data {
			array := strings.Split(b.Id, "01e844f884844aa2bb5d1cab87316c17:") //速控云上对应每个变量的id
			id := strconv.Itoa(i)
			//如果分割后id和设备表上对应设备的id一样，说明他们是统统一个变量，然后赋值
			if array[1] == id {
				box.Data[0].Detail[i].Value = b.Value
				continue
			}
		}
	}
	//历史表插入
	_, err := global.RoastWetElectricHisDataColl.InsertOne(context.Background(), box)
	if err != nil {
		fmt.Println(err)
	}
	//更新数据更新表
	box.UpdateTime = box.CreateTime
	box.CreateTime = ""
	err = global.RoastWetElectricDataColl.FindOneAndUpdate(context.TODO(), bson.M{"boxId": box.BoxId}, bson.M{"$set": box}).Decode(bson.M{})
	if err == mongo.ErrNoDocuments {
		//第一次无记录更新，需要插入新纪录
		box.CreateTime = box.UpdateTime
		if res, err := global.RoastWetElectricDataColl.InsertOne(context.TODO(), box); err != nil {
			fmt.Println("焙烧湿电数据更新存储出错", err.Error())
		} else {
			fmt.Println("焙烧湿电数据更新表创建新纪录", res.InsertedID)
		}
	}
	fmt.Println("焙烧湿电分钟数据更新存储完毕")
	return
}

// 存储石墨化湿电数据
func storeGraphiteWetElectricData(box model.Box, data model.RealtimeData) {
	//查找设备
	var BoxDevice model.Device
	if err := global.DeviceColl.FindOne(context.TODO(), bson.M{"code": box.BoxId}).Decode(&BoxDevice); err != nil {
		log.Println(box.BoxId+"设备不存在", err)
		return
	}
	box.DeviceTypeId = BoxDevice.DeviceTypeId
	box.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	//处理单位
	if len(BoxDevice.Sensors) == 1 {
		box.Data[0].SensorId = BoxDevice.Sensors[0].Code
		box.Data[0].SensorName = BoxDevice.Sensors[0].Name
		for i, a := range box.Data[0].Detail {
			for _, b := range BoxDevice.Sensors[0].DetectionValue {
				if a.Key == b.Key {
					box.Data[0].Detail[i].Unit = b.Unit
					continue
				}
			}
		}
	}
	//处理值
	for i, _ := range box.Data[0].Detail {
		for _, b := range data.Data {
			array := strings.Split(b.Id, "69fb82a9cba744188cab9da766787f25:") //速控云上对应每个变量的id
			id := strconv.Itoa(i)
			//如果分割后id和设备表上对应设备的id一样，说明他们是统统一个变量，然后赋值
			if array[1] == id {
				box.Data[0].Detail[i].Value = b.Value
				continue
			}
		}
	}
	//历史表插入
	_, err := global.GraWetElectricHisDataColl.InsertOne(context.Background(), box)
	if err != nil {
		fmt.Println(err)
	}
	//更新数据更新表
	box.UpdateTime = box.CreateTime
	box.CreateTime = ""
	err = global.GraWetElectricDataColl.FindOneAndUpdate(context.TODO(), bson.M{"boxId": box.BoxId}, bson.M{"$set": box}).Decode(bson.M{})
	if err == mongo.ErrNoDocuments {
		//第一次无记录更新，需要插入新纪录
		box.CreateTime = box.UpdateTime
		if res, err := global.GraWetElectricDataColl.InsertOne(context.TODO(), box); err != nil {
			fmt.Println("石墨化湿电数据更新存储出错", err.Error())
		} else {
			fmt.Println("石墨化湿电数据更新表创建新纪录", res.InsertedID)
		}
	}
	fmt.Println("石墨化湿电分钟数据更新存储完毕")
	return
}
