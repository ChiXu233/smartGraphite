package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"time"
)

func ParseDTUData2(buf []byte, n int) {
	if time.Now().Second() != 0 || n != 57 { //时间和数据长度严格规定
		return
	}
	var data string //原始16进制数据串
	var payload []string
	//fmt.Println(buf[:n])
	for i := 0; i < len(buf[:n]); i++ {
		p := strconv.FormatInt(int64(buf[i]), 16)
		payload = append(payload, p)
		//转换为16进制
		if len(p) < 2 {
			p = "0" + p
		}
		data += p
	}
	fmt.Println(time.Now(), payload)
	DTUId := "ZZ02"               //aa 66 1 4 0 0 0
	data = data[14 : len(data)-4] //字符串数据

	//求出传感器编码位置
	//数据库设备传感器信息
	var DBSensors []model.Sensors
	if err := getSensors(DTUId, &DBSensors); err != nil {
		fmt.Println("错误dtuId")
		return
	}
	var DTUDatas []model.DTUData
	//传感器编码找数据库传感器信息进行解析
	var DTUData model.DTUData
	DTUData.SensorId = DBSensors[0].Code   //传感器code赋值
	DTUData.SensorName = DBSensors[0].Name //传感器name赋值
	valueLoc := 0                          //此时valueLoc为数据开始位置
	for i := 0; i < len(DBSensors[0].DetectionValue); i++ {
		//两个字节拼接转换为10进制是其值 0 79
		// -32768---32767 正数0X0000-0X7FFF 0X8000-X0FFF  //负数为其补码
		b := data[valueLoc : valueLoc+4]
		if valueLoc == 4 {
			b = b[0:2] + b[3:4]
		}
		value, _ := strconv.ParseInt(b, 16, 32)
		if value > 0x7fff {
			value = -(0xffff - value)
		}
		DTUData.DTUDataDetail = append(DTUData.DTUDataDetail, model.DTUDataDetail{
			Key:   DBSensors[0].DetectionValue[i].Key,
			Value: strconv.FormatInt(value, 10),
			Unit:  DBSensors[0].DetectionValue[i].Unit,
		})
		valueLoc += 4 //下一个值解析
	}
	DTUDatas = append(DTUDatas, DTUData)
	//fmt.Println(DTUDatas)
	StoreDTUData2(DTUId, payload, DTUDatas)
}
func StoreDTUData2(DTUId string, payloadArr []string, DTUDatas []model.DTUData) {
	payload := DTUId
	for _, p := range payloadArr {
		payload += " " + p
	}
	DTUInfo := model.DTU{
		Id:         primitive.ObjectID{},
		DTUId:      DTUId,
		DTUData:    DTUDatas,
		Payload:    payload,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	if DTUInfo.DTUData == nil {
		updateTime := time.Now().Format("2006-01-02 15:04:05")
		update := bson.M{"$set": bson.M{"status": "离线", "updateTime": updateTime}}
		err := global.DeviceColl.FindOneAndUpdate(context.TODO(), bson.M{"code": DTUId}, update).Decode(bson.M{})
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Dtu设备离线")
		return
	} else {
		//设备状态更新
		updateTime := time.Now().Format("2006-01-02 15:04:05")
		update := bson.M{"$set": bson.M{"status": "正常", "updateTime": updateTime}}
		err := global.DeviceColl.FindOneAndUpdate(context.TODO(), bson.M{"code": DTUId}, update).Decode(bson.M{})
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Dtu设备在线")
		//温度值/10
		for i, a := range DTUInfo.DTUData {
			for j, b := range a.DTUDataDetail {
				value, err := strconv.ParseFloat(b.Value, 64)
				if err != nil {
					fmt.Println(err)
				}
				DTUInfo.DTUData[i].DTUDataDetail[j].Value = fmt.Sprintf("%.3f", value/10)
			}
		}
		//历史表插入 //如果时间相同则更新
		upsert := true
		_, err = global.DTUHisDataColl.UpdateOne(context.TODO(), bson.M{"DTUId": DTUInfo.DTUId, "createTime": DTUInfo.CreateTime}, bson.M{"$set": DTUInfo}, &options.UpdateOptions{Upsert: &upsert})
		if err != nil {
			fmt.Println("数据存储出错", err.Error())
		}
		DTUInfo.UpdateTime = DTUInfo.CreateTime
		DTUInfo.CreateTime = ""
		//更新数据更新表
		err = global.DTUDataColl.FindOneAndUpdate(context.TODO(), bson.M{"DTUId": DTUInfo.DTUId}, bson.M{"$set": DTUInfo}).Decode(bson.M{})
		if err == mongo.ErrNoDocuments {
			//第一次无记录更新，需要插入新纪录
			DTUInfo.CreateTime = DTUInfo.UpdateTime
			if res, err := global.DTUDataColl.InsertOne(context.TODO(), DTUInfo); err != nil {
				fmt.Println("数据更新存储出错", err.Error())
			} else {
				fmt.Println("隧道窑更新表创建新纪录", res.InsertedID)
			}
		}
	}
}
