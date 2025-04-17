package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"strconv"
	"time"
)

// 焙烧温度定时存储
func RoastDataOperation(interval time.Duration) {
	endTime := time.Now().Format("2006-01-02 15:04:05")
	startTime := time.Now().Add(-interval).Format("2006-01-02 15:04:05")
	//根据焙烧温度传感器类型查询设备
	deviceTypeId, _ := primitive.ObjectIDFromHex("62874bb27cc89967383a5b80")

	var devices []model.Device
	curr, err := global.DeviceColl.Find(context.TODO(), bson.M{"deviceTypeId": deviceTypeId, "isValid": true, "status": "正常"})
	if err != nil {
		log.Println("焙烧设备", err)
		return
	}
	if err := curr.All(context.TODO(), &devices); err != nil {
		log.Println("焙烧设备", err)
		return
	}

	//使用下标索引
	for devI := range devices { //单个设备依次计算
		//待插入数据库中的字段
		var info model.SensorData
		//切片空间增加
		//info.DTUDataDetail = append(info.DTUDataDetail, model.DTUDataDetail{})
		//根据设备绑定的dtu编码在历史表中查询数据
		var dtuDB []model.SensorData
		filter := bson.M{
			"createTime": bson.M{
				"$gte": startTime,
				"$lte": endTime,
			},
			"sensorId": devices[devI].Code,
		}
		//根据filter条件查询历史表
		curr, err := global.CRoastHisDataColl.Find(context.TODO(), filter)
		if err != nil {
			log.Println("焙烧历史数据", err)
			return
		}
		if err = curr.All(context.TODO(), &dtuDB); err != nil {
			log.Println("焙烧历史数据", err)
			return
		}
		if dtuDB == nil {
			log.Println(devices[devI].Name, "此设备在此时间段没有数据")
			continue
		}

		//数据计算
		for i := range dtuDB { //单个设备数据计算
			info.CreateTime = endTime
			//数据详细信息
			if i == 0 { //第一条数据
				info.CreateTime = endTime //创建时间
				info.Code = dtuDB[0].Code //赋值编号,名称
				info.Name = dtuDB[0].Name
				for detail := range dtuDB[0].DTUDataDetail { //初始赋值,数据格式化
					info.DTUDataDetail = append(info.DTUDataDetail, []model.DTUDataDetail{
						{
							Key:   dtuDB[0].DTUDataDetail[detail].Key + "最大值",
							Value: dtuDB[0].DTUDataDetail[detail].Value,
							Unit:  dtuDB[0].DTUDataDetail[detail].Unit,
						},
						{
							Key:   dtuDB[0].DTUDataDetail[detail].Key + "最小值",
							Value: dtuDB[0].DTUDataDetail[detail].Value,
							Unit:  dtuDB[0].DTUDataDetail[detail].Unit,
						},
						{
							Key:   dtuDB[0].DTUDataDetail[detail].Key + "平均值",
							Value: dtuDB[0].DTUDataDetail[detail].Value,
							Unit:  dtuDB[0].DTUDataDetail[detail].Unit,
						},
					}...)
				}
				//fmt.Println(len(info.DTUData[0].DTUDataDetail))
			} else { //数据在遍历的同时进行比较和累加，最后再求平均值
				detailLen := len(dtuDB[0].DTUDataDetail)
				for detail := 0; detail < detailLen; detail++ {
					temp := detail * 3
					if dtuDB[i].DTUDataDetail[detail].Key+"最大值" == info.DTUDataDetail[temp].Key {
						//最大值
						info.DTUDataDetail[temp].Value = stringMax(info.DTUDataDetail[temp].Value, dtuDB[i].DTUDataDetail[detail].Value)
						//最小值
						info.DTUDataDetail[temp+1].Value = stringMin(info.DTUDataDetail[temp+1].Value, dtuDB[i].DTUDataDetail[detail].Value)
						//平均值
						info.DTUDataDetail[temp+2].Value = stringAdd(info.DTUDataDetail[temp+2].Value, dtuDB[i].DTUDataDetail[detail].Value)
					} else {
						fmt.Println("计算错误")
					}

				}
			}
		}
		//数据累加完毕，求平均
		dtuDBLen := len(dtuDB)
		detailLen := len(info.DTUDataDetail)
		for i := 2; i < detailLen; i += 3 {
			info.DTUDataDetail[i].Value = DivisionTen(info.DTUDataDetail[i].Value, dtuDBLen)
		}

		//数据存储
		roastData(info, interval)
	}

}

// 字符串转10进制除n，再转字符串
func DivisionTen(x string, n int) string {
	xf, _ := strconv.ParseFloat(x, 10)
	return fmt.Sprintf("%.3f", xf/float64(n))
}

// 字符串转10进制后相加，再将结果转为字符串
func stringAdd(x, y string) string {
	xf, _ := strconv.ParseFloat(x, 10)
	yf, _ := strconv.ParseFloat(y, 10)
	res := xf + yf

	return fmt.Sprintf("%.3f", res)
}

// 字符串转10进制后比较谁大，再将结果转为字符串
func stringMax(x, y string) string {
	xf, _ := strconv.ParseFloat(x, 10)
	yf, _ := strconv.ParseFloat(y, 10)

	if xf >= yf {
		return fmt.Sprintf("%.3f", xf)
	}
	return fmt.Sprintf("%.3f", yf)
}

// 字符串转10进制后比较谁小，再将结果转为字符串
func stringMin(x, y string) string {
	xf, _ := strconv.ParseFloat(x, 10)
	yf, _ := strconv.ParseFloat(y, 10)
	if xf <= yf {
		return fmt.Sprintf("%.3f", xf)
	}
	return fmt.Sprintf("%.3f", yf)
}

func roastData(DTU model.SensorData, interval time.Duration) {
	var collectionHis *mongo.Collection
	var s string
	switch interval {
	case time.Minute * 10:
		collectionHis = global.CRoastTenMinDataColl
		s = "10分钟数据存储成功"
	case time.Minute * 30:
		collectionHis = global.CRoastThirtyDataColl
		s = "30分钟数据存储成功"
	case time.Hour:
		collectionHis = global.CRoastHourDataColl
		s = "60分钟数据存储成功"
	default:
		return
	}
	if _, err := collectionHis.InsertOne(context.TODO(), DTU); err != nil {
		fmt.Println("数据存储出错", err.Error())
		return
	}
	fmt.Println("焙烧温度" + s)
}
