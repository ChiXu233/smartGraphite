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

func ParseDTUData(buf []byte, n int) {
	//设备编码是前四位ZZ02
	if n < 4 {
		fmt.Println("未知数据", string(buf[:n]), buf[:n])
		return
	}
	DTUId := string(buf[:4]) //设备编码
	var payload []string     //原始16进制数据串
	for i := 0; i < len(buf); i++ {
		//转换为16进制
		payload = append(payload, strconv.FormatInt(int64(buf[i]), 16))
	}
	fmt.Println(DTUId, payload[:n])
	var dataGroups [][]string //设备的每个传感器数据分组
	//求出传感器编码位置
	for i, _ := range payload {
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
			if err := checkCRC(sensorsData); err != nil {
				fmt.Println(err.Error())
				continue
			}
			dataGroups = append(dataGroups, sensorsData)
		}
	}
	//数据库设备传感器信息
	var DBSensors []model.Sensors
	if err := getSensors(DTUId, &DBSensors); err != nil {
		return
	}
	var DTUDatas []model.DTUData
	//传感器编码找数据库传感器信息进行解析
	for _, DBSensor := range DBSensors {
		for _, sensorData := range dataGroups {
			//接收到的传感器code与数据库传感code相比
			if DBSensor.Code == sensorData[0] {
				var DTUData model.DTUData
				DTUData.SensorId = DBSensor.Code   //传感器code赋值
				DTUData.SensorName = DBSensor.Name //传感器name赋值
				valueLoc := 3                      //此时valueLoc为数据开始位置
				for i := 0; i < len(DBSensor.DetectionValue); i++ {
					//两个字节拼接转换为10进制是其值 0 79
					// -32768---32767 正数0X0000-0X7FFF 0X8000-X0FFF  //负数为其补码
					value, _ := strconv.ParseInt(sensorData[valueLoc]+sensorData[valueLoc+1], 16, 32)
					if value > 0x7fff {
						value = -(0xffff - value)
					}
					DTUData.DTUDataDetail = append(DTUData.DTUDataDetail, model.DTUDataDetail{
						Key:   DBSensor.DetectionValue[i].Key,
						Value: strconv.FormatInt(value, 10),
						Unit:  DBSensor.DetectionValue[i].Unit,
					})
					valueLoc += 2 //下一个值解析
				}
				DTUDatas = append(DTUDatas, DTUData)
			}
		}
	}
	StoreDTUData(DTUId, payload[4:n], DTUDatas)
}

//获取设备上传感器信息
func getSensors(deviceCode string, sensors *[]model.Sensors) error {
	var device model.Device
	if err := global.DeviceColl.FindOne(context.TODO(), bson.M{"code": deviceCode, "isValid": true}).Decode(&device); err != nil {
		fmt.Println(deviceCode+"设备不存在", err.Error())
		return err
	}
	var deviceType model.DeviceType
	if err := global.DeviceTypeColl.FindOne(context.TODO(), bson.M{"_id": device.DeviceTypeId}).Decode(&deviceType); err != nil {
		fmt.Println(deviceCode+"设备类型不存在", err.Error())
		return err
	}
	if device.IsCustom {
		*sensors = device.Sensors
	} else {
		*sensors = deviceType.Sensors
	}
	return nil
}

//crc校验
func checkCRC(sensorsData []string) error {
	var dataInt []uint16
	for i := range sensorsData[:len(sensorsData)-2] { //后两位为传过来的crc
		a, _ := strconv.ParseInt(sensorsData[i], 16, 32)
		dataInt = append(dataInt, uint16(a))
	}
	crc := uint16(0xffff) // Initial value
	for j := 0; j < len(dataInt); j++ {
		crc = crc ^ dataInt[j] // crc ^= *data; data++;
		for i := 0; i < 8; i++ {
			if (crc & 1) != 0 {
				crc = (crc >> 1) ^ 0xA001 // 0xA001 = reverse 0x8005
			} else {
				crc >>= 1
			}
		}
	}
	crc16 := strconv.FormatInt(int64(crc), 16)
	switch len(crc16) {
	case 3:
		crc16 = "0" + crc16 //转换16进制3位加0
	case 2:
		crc16 = "00" + crc16
	case 1:
		crc16 = "000" + crc16
	}
	if sensorsData[len(sensorsData)-1]+sensorsData[len(sensorsData)-2] != crc16 {
		return errors.New(crc16 + "CRC校验错误")
	}
	return nil
}
func StoreDTUData(DTUId string, payloadArr []string, DTUDatas []model.DTUData) {
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
		//历史表插入
		if _, err := global.DTUHisDataColl.InsertOne(context.TODO(), DTUInfo); err != nil {
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
				fmt.Println("焙烧更新表创建新纪录", res.InsertedID)
			}
		}
	}
}
