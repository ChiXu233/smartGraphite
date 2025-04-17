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
	"time"
)

// DeviceEchartsOperation 此echarts Code为设备编码+传感器编号
// DeviceEchartsOperation DeviceEchartsOperation 适用于使用动态协议解析的设备计算echarts,optKey 根据最大，小值还是平均值计算,inter指定时间数据表选择,interval计算多长时间的数据,基数为time.Minute,查找的条数上限(时间差值/limitBase)
func DeviceEchartsOperation(idStr, optKey string, inter, interval, limitBase int) {

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		fmt.Println(idStr, err)
		return
	}

	var deviceDB model.Device
	if err = global.DeviceColl.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&deviceDB); err != nil {
		fmt.Println(err)
		return
	}

	//判断对应设备的传感器数据表是否存在,保证协议正确，才开始下面的计算
	if len(deviceDB.Sensors) <= 0 {
		return
	} else {
		for i := range deviceDB.Sensors {
			//多个传感器
			//此时string key为：设备编号+传感器编号
			if _, ok := global.CollMap[deviceDB.Code+deviceDB.Sensors[i].Code][inter]; !ok {
				fmt.Println(deviceDB.Code, inter, "数据表未映射")
				return
			}
		}
	}

	//查找unit
	var unitDB model.Unit
	if err = global.ElectricityUnitColl.FindOne(context.TODO(), bson.M{"DTUId": deviceDB.Code}).Decode(&unitDB); err != nil {
		fmt.Println("数据单位表", err)
		return
	}

	if unitDB.Sensor == nil {
		fmt.Println("Echarts 传感器单位为空")
		return
	}
	var sensorDB []model.UnitSensor
	sensorDB = append(sensorDB, unitDB.Sensor...)
	unitMap := make(map[string]map[string]bool)
	for i := range sensorDB {
		unitMap[sensorDB[i].SensorId] = make(map[string]bool)
		unitMap[sensorDB[i].SensorId] = sensorDB[i].IsEcharts
		for key := range sensorDB[i].IsEcharts {
			unitMap[sensorDB[i].SensorId][key+optKey] = sensorDB[i].IsEcharts[key]
		}

	}
	//以上为预检

	//开始查询数据
	end := time.Now()
	start := time.Now().Add(-time.Duration(interval) * time.Minute)
	endTime := end.Format("2006-01-02 15:04:05")
	startTime := start.Format("2006-01-02 15:04:05")

	filter := bson.M{
		"createTime": bson.M{
			"$lte": endTime,
			"$gte": startTime,
		},
	}

	//获取条数上限
	var limit int64
	temp := end.Sub(start) //时间差值
	limit = int64(temp.Minutes() / float64(limitBase))

	//limit设置查询条数
	opts := options.FindOptions{
		Limit: &limit,
		Sort:  bson.M{"_id": -1},
	}

	//预检已经检查sensors长度是否小于0
	for _, sensor := range deviceDB.Sensors {
		//开始生成echarts
		SensorEchartsOperation(deviceDB.Code, sensor.Name, sensor.Code, inter, unitMap[sensor.Code], global.CollMap[deviceDB.Code+sensor.Code][inter], filter, &opts)
	}

}

// SensorEchartsOperation 基于传感器生成Echarts,设备编号(deviceCode),传感器名称(sensorName),传感器编号(sensorId),单位表isEcharts(unitMap),collMap[key1][key2](coll)
func SensorEchartsOperation(deviceCode, sensorName, sensorId string, inter int, unitMap map[string]bool, coll *mongo.Collection, filter interface{}, opts ...*options.FindOptions) {

	if coll == nil {
		fmt.Println("collMap[" + deviceCode + "][" + sensorId + "]为空")
		return
	}
	//在指定数据库查找数据
	res, err := coll.Find(context.TODO(), filter, opts...)
	if err != nil {
		fmt.Println("echarts "+deviceCode+" "+sensorId, err)
		return
	}

	var sensorDB []model.SensorData
	if err = res.All(context.TODO(), &sensorDB); err != nil {
		fmt.Println("echarts "+deviceCode+" "+sensorId, err)
		return
	}
	if sensorDB == nil {
		fmt.Println("echarts "+deviceCode+" "+sensorId, "没有找到数据")
		return
	}

	var echarts model.Echarts
	echarts.Name = sensorName
	echarts.Code = deviceCode + sensorId + "_" + fmt.Sprintf("%d", inter)

	//记录值的位置
	keyMap := make(map[string]int)
	var head []string                 //生成表头
	head = append(head, "createTime") //创建时间默认在第一列
	//注意map是无序的
	for i := range sensorDB[0].DTUDataDetail {

		if unitMap[sensorDB[0].DTUDataDetail[i].Key] { //是否需要计算
			head = append(head, sensorDB[0].DTUDataDetail[i].Key)
			keyMap[sensorDB[0].DTUDataDetail[i].Key] = i
		}

	}
	//表头添加
	echarts.Data = append(echarts.Data, head)

	//数据处理
	for i := range sensorDB {
		var data []string
		data = append(data, sensorDB[i].CreateTime)
		for k := range sensorDB[i].DTUDataDetail {
			if unitMap[sensorDB[i].DTUDataDetail[k].Key] { //匹配到要计算的值，bool默认值为false
				data = append(data, sensorDB[i].DTUDataDetail[k].Value)
			}
		}
		echarts.Data = append(echarts.Data, data)
	}

	//检查并更新
	echarts.UpdateTime = utils.TimeFormat(time.Now())
	err = global.EchartsColl.FindOneAndUpdate(context.TODO(), bson.M{"code": echarts.Code}, bson.M{"$set": echarts}).Decode(&bson.M{})
	if err != nil {
		echarts.CreateTime = utils.TimeFormat(time.Now())
		//第一次数据存储
		_, err = global.EchartsColl.InsertOne(context.TODO(), echarts)
		if err != nil {
			fmt.Println("echarts "+deviceCode+"设备"+sensorId+"传感器", "数据存储失败")
			return
		}
		fmt.Println("echarts "+deviceCode+"设备"+sensorId+"传感器", "第一次数据存储成功")
		return
	}

	fmt.Println("echarts "+deviceCode+"设备"+sensorId+"传感器", "数据更新成功")

}
