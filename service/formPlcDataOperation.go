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
	"log"
	"time"
)

func FormPlcDataOperation(interval time.Duration) {
	endTime := time.Now().Format("2006-01-02 15:04:05")
	startTime := time.Now().Add(-interval).Format("2006-01-02 15:04:05")
	//根据成型PLC类型查询设备
	deviceTypeId, _ := primitive.ObjectIDFromHex("62a888b0d8a1cbe11b55b494")

	var devices []model.Device
	curr, err := global.DeviceColl.Find(context.TODO(), bson.M{"deviceTypeId": deviceTypeId, "isValid": true, "status": "正常"})
	if err != nil {
		log.Println("成型PLC设备", err)
		return
	}
	if err := curr.All(context.TODO(), &devices); err != nil {
		log.Println("成型PLC设备", err)
		return
	}
	for _, device := range devices { //设备遍历循环
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
		curr, err := global.FormPlcHisDataColl.Find(context.TODO(), filter, &opts)
		if err != nil {
			log.Println("成型PLC设备历史数据", err)
			return
		}
		if err = curr.All(context.TODO(), &boxDB); err != nil {
			log.Println("成型PLC设备历史数据", err)
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
					if (detail.Unit == "" && detail.Value == "0") || (detail.Unit == "" && detail.Value == "1") {
						info.Data[0].Detail = append(info.Data[0].Detail, model.BoxDataDetail{
							Key:   detail.Key,
							Value: detail.Value,
							Unit:  detail.Unit,
						})
					} else {
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
				}
			} else { //数据计算
				dataLen := len(boxDB[i].Data[0].Detail)
				count := 0 //用于标记经过了多少个可求最小，最大，平均的数据项
				for k := 0; k < dataLen; k++ {
					j := 2*count + k //info中detail的key值对应
					if (boxDB[i].Data[0].Detail[k].Unit == "" && boxDB[i].Data[0].Detail[k].Value == "0") || (boxDB[i].Data[0].Detail[k].Unit == "" && boxDB[i].Data[0].Detail[k].Value == "1") {
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
		//过滤掉多余值
		if len(info.Data[0].Detail) >= 368 {
			info.Data[0].Detail = info.Data[0].Detail[:368]
		}
		formPlcColl(info, interval)
	}

}

func formPlcColl(box model.Box, interval time.Duration) {
	var coll *mongo.Collection
	var msg string
	switch interval {
	case 10 * time.Minute:
		coll = global.FormPlcTenMinDataColl
		msg = "10分钟"
	case 30 * time.Minute:
		coll = global.FormPlcThirtyMinDataColl
		msg = "30分钟"
	case time.Hour:
		coll = global.FormPlcHourDataColl
		msg = "小时"
	default:
		fmt.Println("成型PLC时间粒度选择错误")
		return
	}
	_, err := coll.InsertOne(context.TODO(), box)
	if err != nil {
		fmt.Println("成型PLC" + msg + "存储失败")
		return
	}
	fmt.Println("成型PLC" + msg + "存储成功")
	return
}
