package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"SmartGraphite-server/utils"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
	"time"
)

func ParseDTUDataNew(buf []byte, n int) {

	var payload []string //原始16进制数据串
	for i := 0; i < len(buf); i++ {
		//转换为16进制
		payload = append(payload, strconv.FormatInt(int64(buf[i]), 16))
	}

	payload = payload[:]
	var DTUId string
	//寻找DTUId在第一个aa 55前面的第一个位置
	for i := 1; i < (len(payload) - 1); i++ {
		if payload[i] == "aa" && payload[i+1] == "55" {
			DTUId = payload[i-1]
			break
		}
	}

	if DTUId == "" {
		fmt.Println("未知数据包", buf[:n])
		return
	}

	if DTUId == "c2" {
		//三项电表协议设备编号与焙烧c2混了，焙烧设备设备编号+1
		DTUId = "c21"
	}

	//原始数据打印
	fmt.Println(DTUId, payload[:n])
	fmt.Println(DTUId, buf[:n])
	payload = payload[:n]
	//判断aa 55 后面是否为空值
	if len(payload) < 6 {
		fmt.Println("错误数据", payload[:n])
		return
	}

	//有效字节数的位置，每次计算的字节数，要省略的数据
	var dLen, bits, omit, sensorHead, crcLen int
	//是否使用补码运算，是否格式化时间，是否把全部传感器数据合并，是否使用数据单位计算，是否存在相同的传感器编号(aa 55数据包后的地址码)
	var complement, timeFormat, sensorMerge, unit, validCode, crc bool
	//保存的值格式
	var sprintf string

	//是否进行e2挤压报表计算
	var report string
	//根据设备编号选择配置
	switch DTUId {
	case "e2":
		sensorHead = 3
		crcLen = 2
		dLen = 3
		bits = 2
		omit = 1
		complement = false
		timeFormat = false
		sensorMerge = true
		unit = true
		validCode = true
		crc = false
		sprintf = "%0.3f"
		report = "e2"

	case "e4":
		sensorHead = 3
		crcLen = 2
		dLen = 3
		bits = 2
		omit = 1
		complement = false
		timeFormat = false
		sensorMerge = true
		unit = true
		validCode = true
		crc = false
		sprintf = "%0.3f"
		report = "e4"

	case "e5":
		sensorHead = 3
		crcLen = 2
		dLen = 3
		bits = 2
		omit = 1
		complement = false
		timeFormat = false
		sensorMerge = true
		unit = true
		validCode = true
		crc = false
		sprintf = "%0.3f"
		report = "e5"
	case "c1", "c21", "c3", "c4", "c5":
		sensorHead = 3
		crcLen = 2
		dLen = 3
		bits = 2
		omit = 1
		complement = false
		timeFormat = false
		sensorMerge = true
		unit = false
		validCode = true
		crc = false
		sprintf = "%0.3f"

	default:
		fmt.Println("未知设备：", DTUId)
		return
	}

	ParseDTU(payload, DTUId, omit, dLen, bits, sensorHead, crcLen, complement, timeFormat, sensorMerge, unit, validCode, crc, sprintf, report)

}

// ParseDTU 原始16进制数据，DTUId，DTUIdIndex(DTUId的位置),omit(需要省略的数据),数据包有效字节位置(payload[dLen])
// 每次计算的位数(bits),是否要使用补码计算(complement)
// 是否合并传感器数据(sensorMerge),是否要根据单位计算(unit),是否有相同的传感器编号(validCode),是否开启crc校验(crc)
// Sprintf字符串格式(sprintf)
func ParseDTU(payload []string, DTUId string, omit, dLen, bits, sensorHead, crcLen int, complement, timeFormat, sensorMerge, unit, validCode, crc bool, sprintf string, report string) {
	if len(payload) < omit {
		fmt.Println(DTUId + " 省略数据长度错误")
		return
	}
	var Len = len(payload)
	if payload[Len-1] == "55" && payload[Len-2] == "aa" {
		payload = payload[omit : Len-2]
	} else {
		payload = payload[omit:]
	}
	//设备数据,一个设备可以有多个传感器
	var deviceData [][]string
	payLen := len(payload) - 1
	for i := 0; i < payLen; i++ {
		if payload[i] == "aa" && payload[i+1] == "55" && payload[i+2] != "" {
			if (i + dLen + 1) >= len(payload) {
				fmt.Println("数据错误", payload)
				return
			}
			dataLen, _ := strconv.ParseInt(payload[i+dLen+1], 16, 32) //每个数据包具体的数据长度
			if dataLen < 1 {
				return
			}
			//跳过aa 55
			i += 2
			var sensorData []string
			//地址位 功能码 数据长度
			sensorData = append(sensorData, payload[i:i+sensorHead]...)
			//携带的数据
			sensorData = append(sensorData, payload[i+sensorHead:i+sensorHead+int(dataLen)]...)
			//crc校验码
			sensorData = append(sensorData, payload[i+sensorHead+int(dataLen):i+sensorHead+int(dataLen)+crcLen]...)
			//crc16modbus低位和高位
			if crc {
				if err := checkCRC3(sensorData); err != nil {
					fmt.Println(err.Error())
					continue
				}
			}
			deviceData = append(deviceData, sensorData)
		}
	}
	//获取数据协议信息
	var sensorDB []model.Sensors
	if err := getSensors3(DTUId, &sensorDB); err != nil {
		return
	}

	var DTUData []model.SensorData

	timeStr := utils.TimeFormat(time.Now())

	//是否需要根据单位计算
	var unitDB model.Unit
	//方便取值，时间复杂度减至o(1)，空间换时间
	var uMap map[string]model.UnitSensor
	var unitMap map[string]model.UnitMapDetail
	if unit {
		if err := global.ElectricityUnitColl.FindOne(context.TODO(), bson.M{"DTUId": DTUId}).Decode(&unitDB); err != nil {
			fmt.Println("设备编号为：" + DTUId + " 寻找单位计算协议出错")
			return
		}

		uMap = make(map[string]model.UnitSensor)
		unitMap = make(map[string]model.UnitMapDetail)
		for i := range unitDB.Sensor {
			uMap[unitDB.Sensor[i].SensorId] = unitDB.Sensor[i]
		}

	}

	if unit && uMap == nil {
		fmt.Println("设备编号为：" + DTUId + " 单位计算协议为空")
		return
	}

	if validCode { //有相同的传感器编号,合并相同的传感器编号的数据,按照先后顺序合并

		sensorMap := make(map[string][]string)
		for i := range deviceData {
			if _, ok := sensorMap[deviceData[i][0]]; !ok { //不存在这个键
				sensorMap[deviceData[i][0]] = deviceData[i][:len(deviceData[i])-2] //,记录数据并除去crc
				continue
			}
			//存在这个键，数据拼接
			sensorMap[deviceData[i][0]] = append(sensorMap[deviceData[i][0]], deviceData[i][dLen:len(deviceData[i])-2]...)

		}
		//对deviceData重新赋值
		deviceData = append(deviceData[:0]) //删除原数据
		for key := range sensorMap {
			deviceData = append(deviceData, sensorMap[key])
		}
	}

	if sensorDB == nil {
		fmt.Println("设备编号为：" + DTUId + " 未添加协议")
	}

	//传感器协议协议
	for i := range sensorDB {
		for _, sensorData := range deviceData {

			if sensorData[0] == sensorDB[i].Code { //传感器编码匹配
				var sensorInfo model.SensorData
				sensorInfo.Code = sensorDB[i].Code //传感器编号
				sensorInfo.Name = sensorDB[i].Name //传感器名称
				valueLoc := dLen                   //数据开始位置
				valueStr := ""
				//判断协议是否正确
				var value int64
				if validCode {
					//经过合并的传感器数据
					value = int64(len(sensorData[dLen:]))
				} else {
					value, _ = strconv.ParseInt(sensorData[dLen-1], 16, 32) //获取有效字节数
				}
				if int64(len(sensorDB[i].DetectionValue)) != value/int64(bits) {
					fmt.Println("设备编号为："+DTUId+" 的 "+sensorDB[i].Code+" 传感器协议不匹配", "协议长度：", len(sensorDB[i].DetectionValue), "数据个数长度：", value/int64(bits))
					continue
				}

				//单位值赋值
				unitMap = uMap[sensorDB[i].Code].UnitMap

				for k := 0; k < len(sensorDB[i].DetectionValue); k++ {

					//每次计算之前拼接
					for bit := 0; bit < bits; bit++ {
						//补0
						if len(sensorData[valueLoc+bit]) < 2 {
							valueStr += "0" + sensorData[valueLoc+bit]
						} else {
							valueStr += sensorData[valueLoc+bit]
						}
					}

					result, err := strconv.ParseUint(valueStr, 16, 16)
					if err != nil {
						fmt.Println("转换失败：", err)
						return
					}
					wordValue := int16(result) // 将无符号整型转换为有符号的WORD类型
					value = int64(wordValue)   //强制转化为int64类型
					//	//value, _ := strconv.ParseInt(sensorData[valueLoc]+sensorData[valueLoc+1], 16, 32)
					//	value, _ = strconv.ParseInt(valueStr, 16, 32)
					if complement && value > 0x7fff { //是否要进行补码运算
						value = -(0xffff - value)
					}

					//单位值是否需要计算
					if unit {
						if unitMap[sensorDB[i].DetectionValue[k].Key].Flag { //是否计算乘变比
							valueStr = fmt.Sprintf(sprintf, float64(value)*uMap[sensorDB[i].Code].Var*unitMap[sensorDB[i].DetectionValue[k].Key].Value)
						} else {
							valueStr = fmt.Sprintf(sprintf, float64(value)*unitMap[sensorDB[i].DetectionValue[k].Key].Value)
						}
					} else {
						valueStr = fmt.Sprintf(sprintf, float64(value))
					}

					sensorInfo.DTUDataDetail = append(sensorInfo.DTUDataDetail, model.DTUDataDetail{
						Key:   sensorDB[i].DetectionValue[k].Key,
						Value: valueStr,
						Unit:  sensorDB[i].DetectionValue[k].Unit,
					})

					valueStr = ""    //数值字符串置空
					valueLoc += bits //下一个值解析
				}

				//传感器数据记录
				DTUData = append(DTUData, sensorInfo)

			}

		}
	}
	//模拟数据c6
	var DTUDataC6 []model.SensorData
	if DTUId == "c1" {
		var DTUC6Details []model.DTUDataDetail
		//深拷贝,避免修改模拟值c6影响原本数据
		for i := range DTUData[0].DTUDataDetail {
			DTUC6Details = append(DTUC6Details, model.DTUDataDetail{Key: DTUData[0].DTUDataDetail[i].Key, Value: "-1999.000", Unit: DTUData[0].DTUDataDetail[i].Unit})
		}
		DTUDataC6 = append(DTUDataC6, model.SensorData{DTUDataDetail: DTUC6Details, Name: "温度传感器", Code: "1"})
	}

	//判断数据是否存在t6,避免停电数据丢失
	var flag bool
	for i := range DTUData {
		for k := range DTUData[i].DTUDataDetail {
			if DTUData[i].DTUDataDetail[k].Key == "t6" {
				if stringCompare("500", DTUData[i].DTUDataDetail[k].Value) {
					flag = true
					break
				}
				if stringCompare(DTUData[i].DTUDataDetail[k].Value, "500") {
					DTUData[i].DTUDataDetail[k].Value = "-999.000"
					flag = true
					break
				}
			} else {
				flag = false
			}
		}
	}
	if flag == false && DTUId != "c1" && DTUId != "c21" && DTUId != "c3" && DTUId != "c4" && DTUId != "c5" && DTUId != "c6" {
		var t6 model.DTUDataDetail
		t6.Key = "t6"
		t6.Value = "-999.000"
		t6.Unit = "℃"
		DTUData[0].DTUDataDetail = append(DTUData[0].DTUDataDetail, t6)
	}
	for i := 0; i < len(DTUData); i++ {
		DTUData[i].CreateTime = timeStr
	}

	//数据存储,按设备分表，同一个设备的传感器存入一个表，可自定义修改
	var coll *mongo.Collection

	//寻找存储表
	for _, sensor := range sensorDB {
		val := global.CollMap[DTUId+sensor.Code][1]
		if DTUId == "c1" || DTUId == "c21" || DTUId == "c3" || DTUId == "c4" || DTUId == "c5" {
			val = global.CRoastHisDataColl
		}
		if DTUId == "c1" {
			storeSensorData(global.CRoastHisDataColl, "c6", DTUDataC6, timeFormat, sensorMerge, report)
		}
		if val == nil {
			continue
		}
		coll = val
		storeSensorData(coll, DTUId, DTUData, timeFormat, sensorMerge, report)
	}
	//switch DTUId {
	//case "ba":
	//	//1号焙烧变频器
	//	coll = global.BATransducer
	//case "bb":
	//	//2号焙烧变频器
	//	coll = global.BBTransducer
	//case "bc":
	//	//3号焙烧变频器
	//	coll = global.BCTransducer
	//case "ee":
	//	coll = global.RoastTemp
	//
	//default:
	//	fmt.Println("设备编号为：" + DTUId + " 的设备没有分表")
	//	return
	//}
	//storeSensorData(coll, DTUId, DTUData, timeFormat, sensorMerge)

}

// 数据存储,按设备分表，同一个设备的传感器存入一个表，可自定义修改
func storeSensorData(coll *mongo.Collection, DTUId string, data []model.SensorData, timeFormat, sensorMerge bool, report string) {
	if coll == nil {
		fmt.Println("指定表不存在")
		return
	}

	//格式化时间
	updateTime := ""
	if timeFormat {
		updateTime = time.Now().Format("2006-01-02 15:04") + ":00"
	} else {
		updateTime = time.Now().Format("2006-01-02 15:04:05")
		if DTUId == "c6" {
			updateTime = time.Now().Add(time.Second * 5).Format("2006-01-02 15:04:05")
		}
	}

	var device model.Device
	//查看设备是否离线
	if data == nil { //设备离线
		update := bson.M{"$set": bson.M{"status": "离线", "updateTime": updateTime}}
		err := global.DeviceColl.FindOneAndUpdate(context.TODO(), bson.M{"code": DTUId}, update).Decode(bson.M{})
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("设备编号为：" + DTUId + " 的设备离线")
		return
	}

	update := bson.M{"$set": bson.M{"status": "正常", "updateTime": updateTime}}
	err := global.DeviceColl.FindOneAndUpdate(context.TODO(), bson.M{"code": DTUId}, update).Decode(&device)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("设备编号为：" + DTUId + " 的设备在线")

	//传感器数据是否合并
	if sensorMerge {
		//将所有传感器数据合并到第一个传感器中，新传感器编号为设备编号
		for i := 1; i < len(data); i++ {
			data[0].DTUDataDetail = append(data[0].DTUDataDetail, data[i].DTUDataDetail...)
		}
		data[0].Code = device.Code
		data[0].Name = device.Name
		data = append(data, data[0])
	}

	switch report {
	case "e2":
		E2extrusionTime(data)
	case "e5":
		E5extrusionTime(data)
	default:
		fmt.Println("不需要报表计算")
	}

	//数据存储
	for i := 0; i < len(data); i++ {
		data[i].CreateTime = updateTime
		err = coll.FindOneAndUpdate(context.TODO(), bson.M{"createTime": data[i].CreateTime}, bson.M{"$set": data[i]}).Decode(&bson.M{})
		if err != nil {
			_, err = coll.InsertOne(context.TODO(), data[i])
			if err != nil {
				fmt.Println("设备编号为：" + DTUId + " 数据存储更新失败")
			}
		}
	}
}
