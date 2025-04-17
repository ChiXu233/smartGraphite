package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
	"time"
)

func ParseDTUThreeElectricity(buf []byte, n int) {

	if n < 12 || len(buf) < 12 {
		fmt.Println("未知数据", buf[:n])
	}

	var payload []string //原始16进制数据串
	for i := 0; i < len(buf); i++ {
		//转换为16进制
		payload = append(payload, strconv.FormatInt(int64(buf[i]), 16))

	}

	DTUId := payload[11]   //设备编码
	payload = payload[12:] //省略前面
	fmt.Println(DTUId, payload[:n])
	fmt.Println(buf[:n])
	var dataGroups [][]string //设备的每个传感器数据分组
	//求出传感器编码位置
	//for i, _ := range payload {
	for i := 0; i < (len(payload) - 1); i++ {
		if payload[i] == "aa" && payload[i+1] == "55" {
			dataLen, _ := strconv.ParseInt(payload[i+4], 16, 32) //具体数据长度
			var sensorsData []string
			//地址位 功能码 数据长度
			sensorsData = append(sensorsData, payload[i+2], payload[i+3], payload[i+4])
			//详细数据
			sensorsData = append(sensorsData, payload[i+5:i+5+int(dataLen)]...)
			//crc校验码
			sensorsData = append(sensorsData, payload[i+5+int(dataLen):i+5+int(dataLen)+2]...)
			//crc16modbus低位和高位
			if err := checkCRC3(sensorsData); err != nil {
				fmt.Println(err.Error())
				continue
			}
			dataGroups = append(dataGroups, sensorsData)
		}
	}

	//数据格式化
	var err error
	dataGroups, err = formMatDataGroup(dataGroups)
	if err != nil {
		return
	}
	//数据库设备传感器信息
	var DBSensors []model.Sensors
	if err := getSensors3(DTUId, &DBSensors); err != nil {
		return
	}
	var DTUDatas []model.DTUData
	//传感器编码找数据库传感器信息进行解析
	for _, DBSensor := range DBSensors {
		for _, sensorData := range dataGroups {
			//fmt.Println(DBSensor.Code, "sensorCode", sensorData[0])
			//接收到的传感器code与数据库传感code相比
			if len(sensorData[0]) < 2 {
				sensorData[0] = "0" + sensorData[0]
			}
			if DBSensor.Code == sensorData[0]+DTUId {
				var DTUData model.DTUData
				DTUData.SensorId = DBSensor.Code   //传感器code赋值
				DTUData.SensorName = DBSensor.Name //传感器name赋值
				valueLoc := 3                      //此时valueLoc为数据开始位置

				for i := 0; i < len(DBSensor.DetectionValue); i++ {
					//两个字节拼接转换为10进制是其值 0 79
					// -32768---32767 正数0X0000-0X7FFF 0X8000-X0FFF  //负数为其补码
					value, _ := strconv.ParseInt(sensorData[valueLoc]+sensorData[valueLoc+1]+sensorData[valueLoc+2]+sensorData[valueLoc+3], 16, 32)
					if value > 0x7fff {
						value = -(0xffff - value)
					}
					if DBSensor.DetectionValue[i].Unit == "V" || DBSensor.DetectionValue[i].Unit == "W" || DBSensor.DetectionValue[i].Unit == "Var" {
						DTUData.DTUDataDetail = append(DTUData.DTUDataDetail, model.DTUDataDetail{
							Key:   DBSensor.DetectionValue[i].Key,
							Value: strconv.FormatInt(value/120/10, 10),
							Unit:  DBSensor.DetectionValue[i].Unit,
						})
					} else if DBSensor.DetectionValue[i].Unit == "A" || DBSensor.DetectionValue[i].Unit == "Kwh" {
						DTUData.DTUDataDetail = append(DTUData.DTUDataDetail, model.DTUDataDetail{
							Key:   DBSensor.DetectionValue[i].Key,
							Value: strconv.FormatInt(value/120/100, 10),
							Unit:  DBSensor.DetectionValue[i].Unit,
						})
					}
					valueLoc += 4 //下一个值解析
				}
				DTUDatas = append(DTUDatas, DTUData)
			}
		}
	}
	StoreDTUThreeElectricityData(DTUId, payload[4:n], DTUDatas)

}

//仅在此页面使用,三相电表
func formMatDataGroup(dataGroups [][]string) ([][]string, error) {

	if len(dataGroups) != 2 {
		return nil, errors.New("三相电表数组长度错误")
	}

	var dataGroup [][]string
	dataGroup = append(dataGroup, dataGroups[0])
	dataGroup[0] = append(dataGroup[0], dataGroups[1][3:]...)

	return dataGroup, nil
}

func StoreDTUThreeElectricityData(DTUId string, payloadArr []string, DTUDatas []model.DTUData) {
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
		fmt.Println(DTUId + "设备离线")
		return
	} else {
		//设备状态更新
		updateTime := time.Now().Format("2006-01-02 15:04:05")
		update := bson.M{"$set": bson.M{"status": "正常", "updateTime": updateTime}}
		err := global.DeviceColl.FindOneAndUpdate(context.TODO(), bson.M{"code": DTUId}, update).Decode(bson.M{})
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(DTUId + "设备在线")
		//温度值/10，焙烧DTU不用除10
		for i, a := range DTUInfo.DTUData {
			for j, b := range a.DTUDataDetail {
				value, err := strconv.ParseFloat(b.Value, 64)
				if err != nil {
					fmt.Println(err)
				}
				if value < 0 {
					value = 0
				}
				//DTUInfo.DTUData[i].DTUDataDetail[j].Value = fmt.Sprintf("%.3f", value/10)
				DTUInfo.DTUData[i].DTUDataDetail[j].Value = fmt.Sprintf("%.3f", value)
			}
		}
		//历史表插入
		if _, err := global.ThreeElectricityHisDataColl.InsertOne(context.TODO(), DTUInfo); err != nil {
			fmt.Println("三相电表数据存储出错", err.Error())
		}
		DTUInfo.UpdateTime = DTUInfo.CreateTime
		DTUInfo.CreateTime = ""
		//更新数据更新表
		err = global.ThreeElectricityDataColl.FindOneAndUpdate(context.TODO(), bson.M{"DTUId": DTUInfo.DTUId}, bson.M{"$set": DTUInfo}).Decode(bson.M{})
		if err == mongo.ErrNoDocuments {
			//第一次无记录更新，需要插入新纪录
			DTUInfo.CreateTime = DTUInfo.UpdateTime
			if res, err := global.ThreeElectricityDataColl.InsertOne(context.TODO(), DTUInfo); err != nil {
				fmt.Println("三相电表数据更新存储出错", err.Error())
			} else {
				fmt.Println("三相电表更新表创建新纪录", res.InsertedID)
			}
		}
	}
}

func GetUnitMap(DTUId string) (model.Unit, error) {

	var unitDB model.Unit
	err := global.ElectricityUnitColl.FindOne(context.TODO(), bson.M{"DTUId": DTUId, "isValid": true}).Decode(&unitDB)
	if err != nil {
		return model.Unit{}, err
	}

	return unitDB, nil

}
