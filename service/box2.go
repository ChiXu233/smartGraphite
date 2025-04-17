package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/robfig/cron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var token string
var spec string

// plc数据定时器
func TokenTimer2() {
	if runtime.GOOS != "linux" {
		return
	}
	GetToken()
	c := cron.New() //新建一个定时任务对象
	//定时获取token
	_ = c.AddFunc(global.Spec, GetToken)
	//每分钟存储生产工艺数据
	_ = c.AddFunc("0 */1 * * * *", GetProject)
	c.Start() //开始
	select {} //阻塞住,保持程序运行
}

// 首次获取token
func GetToken() {
	timeUnix := time.Now().UnixNano() / 1e6
	uid := "00bdd125113744de83b690a7a896b69b"
	sid := "b3bd0acff064472db2944173a8470640"
	random := fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000)) //生成0-1000随机字符串
	timestamp := fmt.Sprintf("%v", timeUnix)
	ctx := md5.New()
	ctx.Write([]byte(uid + sid + random + timestamp))
	signature := strings.ToUpper(hex.EncodeToString(ctx.Sum(nil))) //签名转换成字符串和大写32位
	//http请求调用初始化token接口
	URL := "http://sukon-cloud.com/api/v1/token/initToken"
	urlValues := url.Values{}
	urlValues.Add("uid", uid)
	urlValues.Add("sid", sid)
	urlValues.Add("random", random)
	urlValues.Add("timestamp", timestamp)
	urlValues.Add("signature", signature)
	res, err := http.PostForm(URL, urlValues) //发送请求
	if err != nil {
		fmt.Println("plc2获取token失败", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Println("文件流关闭错误", err)
		}
	}()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	var data model.Token
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println(err)
	}
	token = data.Data.Token
	hour := int(math.Floor(float64(data.Data.Expire / 3600)))

	if hour <= 0 {
		fmt.Println("hour", hour)
		GetToken()
		fmt.Println("token时效为0，重新获取token")
		return
	} else {
		if hour == 24 {
			hour = hour - 1
		}
		t := strconv.Itoa(hour)
		spec = "0 0 */" + t + " * * *"
		fmt.Println("获取token成功")
	}
	return
}

// 获取项目
func GetProject() {
	URL := "http://sukon-cloud.com/api/v1/base/projects"
	urlValues := url.Values{}
	urlValues.Add("token", token)
	res, err := http.PostForm(URL, urlValues) //发送请求
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	var project model.Project
	err = json.Unmarshal(body, &project)
	if err != nil {
		fmt.Println("body转project错误", err)
		return
	}
	if project.Data == nil && project.Msg == "token已过期" {
		fmt.Println("project数据为空", project)
		GetToken()
		return
	}
	fmt.Println(project.Data, "fanhuide")
	for i := range project.Data {
		if project.Data[i].Id == "rKWw9LNBQYH" {
			continue
		}
		GetProjectBoxes(token, project.Data[i].Id)
	}
}

// 获取box
func GetProjectBoxes(token, projectId string) {
	URL := "http://sukon-cloud.com/api/v1/base/projectBoxes"
	urlValues := url.Values{}
	urlValues.Add("token", token)
	urlValues.Add("projectId", projectId)
	res, err := http.PostForm(URL, urlValues) //发送请求
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	var box model.ProjectBox
	err = json.Unmarshal(body, &box)
	if err != nil {
		fmt.Println(err)
		return
	}
	if box.Success == false {
		fmt.Println("获取box异常", box)
	}
	for i := range box.Data {
		if box.Data[i].BoxId == "3d34d1b2385c4eafb94ffe89cfd6f43d" {
			continue
		}
		updateTime := time.Now().Format("2006-01-02 15:04:05")
		if box.Data[i].Status == "0" { //设备不在线，更改对应设备状态
			update := bson.M{"$set": bson.M{"status": "离线", "updateTime": updateTime}}
			err = global.DeviceColl.FindOneAndUpdate(context.TODO(), bson.M{"code": box.Data[i].BoxId}, update).Decode(bson.M{})
			if err != nil {
				fmt.Println(box.Data[i].BoxId, err)
				continue
			}
			fmt.Println(box.Data[i].Name + "设备离线")
			continue
		} else {
			update := bson.M{"$set": bson.M{"status": "正常", "updateTime": updateTime}}
			err := global.DeviceColl.FindOneAndUpdate(context.TODO(), bson.M{"code": box.Data[i].BoxId}, update).Decode(bson.M{})
			if err != nil {
				fmt.Println(box.Data[i].BoxId, err)
				continue
			}
			fmt.Println(box.Data[i].Name + "设备在线")
		}
		GetBoxPlc(token, box.Data[i].BoxId)
	}
}

// 获取plc
func GetBoxPlc(token, boxId string) {
	URL := "http://sukon-cloud.com/api/v1/base/boxPlcs"
	urlValues := url.Values{}
	urlValues.Add("token", token)
	urlValues.Add("boxId", boxId)
	res, err := http.PostForm(URL, urlValues) //发送请求
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Println(err)
		}
	}()
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
	fmt.Println(data.Data, "sssss")
	for _, a := range data.Data {
		GetVariant(token, boxId, a.PlcId)
	}
}

// 获取变量
func GetVariant(token, boxId, plcId string) {
	URL := "http://sukon-cloud.com/api/v1/base/boxVariants"
	urlValues := url.Values{}
	urlValues.Add("token", token)
	urlValues.Add("boxId", boxId)
	urlValues.Add("plcId", plcId)
	res, err := http.PostForm(URL, urlValues) //发送请求
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err = res.Body.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	//获取每个sid对应的变量
	var data model.BoxVariant
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
	GetRealTimeData(token, variantIds, box)
}

// 获取实时数据
func GetRealTimeData(token, variantIds string, box model.Box) {
	var data model.RealtimeData
	URL := "http://sukon-cloud.com/api/v1/data/realtimeDatas"
	urlValues := url.Values{}
	urlValues.Add("token", token)
	urlValues.Add("variantIds", variantIds)
	res, err := http.PostForm(URL, urlValues)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err = res.Body.Close(); err != nil {
			fmt.Println(err)
		}
	}()
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
	PlcCollSwitch(box, data)
}

// 表选择，数据存储
func PlcCollSwitch(box model.Box, data model.RealtimeData) {
	var coll *mongo.Collection
	var collHis *mongo.Collection
	var collName string
	if box.BoxId != "2da580adb26b4a12accd4aec80e04656" && box.BoxId != "b46a0faf11cc4000a4c290eba5cc949a" && box.BoxId != "be67c2b8216e49e8981a95663413f115" && box.BoxId != "5cba298477bc456ab1a2bd06e35cb0d8" && box.BoxId != "01e844f884844aa2bb5d1cab87316c17" && box.BoxId != "69fb82a9cba744188cab9da766787f25" && box.BoxId != "f73fe0d8688046e088bb073849aa0c3f" && box.BoxId != "9bd62f734af94dc0b0641817ac2807e9" && box.BoxId != "9f62bc0edbd542b2bec159ac8f023509" {
		fmt.Println("new box", box.BoxId)
	}
	switch box.BoxId {
	//成型plc
	case "65d27a491d744a0e91b4d8e6db628887":
		coll = global.FormPlcDataColl
		collHis = global.FormPlcHisDataColl
		collName = "成型PLC"

	//煅烧脱硝
	case "ef62aa2e44204b5d82463b72a86f9621":
		coll = global.CalDenitrification
		collHis = global.CalDenitrificationHis
		collName = "煅烧脱硝"

	//焙烧脱硝
	case "52980204e2dc4ce9907196441c6f9a32":
		coll = global.RoastDenitrification
		collHis = global.RoastDenitrificationHis
		collName = "焙烧脱硝"

	default:
		fmt.Println("new boxId", box.BoxId)
		return
	}

	//查找设备
	var BoxDevice model.Device
	if err := global.DeviceColl.FindOne(context.TODO(), bson.M{"code": box.BoxId}).Decode(&BoxDevice); err != nil {
		log.Println(box.BoxId+"设备不存在", err)
		return
	}
	box.DeviceTypeId = BoxDevice.DeviceTypeId
	box.CreateTime = time.Now().Format("2006-01-02 15:04:05")

	//初次运行设备详细信息key赋值
	//for i, detail := range box.Data[0].Detail {
	//	BoxDevice.Sensors[0].DetectionValue = append(BoxDevice.Sensors[0].DetectionValue, model.DetectionValue{})
	//	BoxDevice.Sensors[0].DetectionValue[i].Key = detail.Key
	//}
	//
	//opts := options.FindOneAndUpdate().SetUpsert(false)
	//update := bson.D{{"$set", BoxDevice}}
	//res := global.DeviceColl.FindOneAndUpdate(context.TODO(), bson.M{"code": BoxDevice.Code}, update, opts)
	//
	//if err := res.Decode(&BoxDevice);err!=nil{
	//	fmt.Println("key更新错误")
	//	return
	//}
	//初次运行设备详细信息key赋值

	//处理单位
	if len(BoxDevice.Sensors) == 1 {
		//fmt.Println(BoxDevice.Sensors[0].DetectionValue,BoxDevice.Code)
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
			array := strings.Split(b.Id, box.BoxId+":")
			id := strconv.Itoa(i)
			if array[1] == id {
				box.Data[0].Detail[i].Value = b.Value
				continue
			}
		}
	}
	fmt.Println(box, "box")
	//历史表插入
	_, err := collHis.InsertOne(context.Background(), box)
	if err != nil {
		fmt.Println(collName, err)
	}
	//更新数据更新表
	box.UpdateTime = box.CreateTime
	box.CreateTime = ""
	err = coll.FindOneAndUpdate(context.TODO(), bson.M{"boxId": box.BoxId}, bson.M{"$set": box}).Decode(bson.M{})
	if err == mongo.ErrNoDocuments {
		//第一次无记录更新，需要插入新纪录
		box.CreateTime = box.UpdateTime
		if res, err := coll.InsertOne(context.TODO(), box); err != nil {
			fmt.Println(collName+"数据更新存储出错", err.Error())
		} else {
			fmt.Println(collName+"更新表创建新纪录", res.InsertedID)
		}
	}
	fmt.Println(collName + "分钟数据更新存储完毕")
	return
}
