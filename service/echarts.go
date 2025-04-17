package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"SmartGraphite-server/utils"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func EchartsOperation(name, code, interval string) {

	//Box分钟数据格式
	type RetMinData2 struct {
		CreateTime string                `json:"createTime"` //创建时间
		Data       []model.BoxDataDetail `json:"data"`
	}

	var coll *mongo.Collection
	var limit int64
	switch interval {
	case "5":
		coll = global.GraphitingFiveDataColl
		limit = 577
	case "15":
		coll = global.GraphitinFifteenDataColl
		limit = 193
	default:
		return
	}

	count, _ := coll.CountDocuments(context.TODO(), bson.M{})
	if count < limit {
		if count < 2 {
			return
		}
		limit = count
	}

	endTime := utils.TimeFormat(time.Now())
	startTime := utils.TimeFormat(time.Now().Add(-48 * time.Hour))
	filter := bson.M{
		"createTime": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}
	opts := options.FindOptions{
		Limit: &limit,
		Sort:  bson.M{"_id": -1},
	}
	res, err := coll.Find(context.TODO(), filter, &opts)
	if err != nil {
		return
	}

	var infoDB []model.Box
	if err = res.All(context.TODO(), &infoDB); err != nil {
		return
	}

	//var infoDB []model.Box
	var info []RetMinData2
	if infoDB == nil {
		return
	}
	//开始计算位移变化量
	//查找所需要的值,记录对应元素在数组中的位置
	keyMap := make(map[string]int)
	for i := 0; i < 1; i++ {
		for key := range infoDB[0].Data[0].Detail {
			keyMap[infoDB[0].Data[0].Detail[key].Key] = key
		}
	}
	if len(keyMap) == 0 { //记录失败
		return
	}
	//for i := 0; i < int(limit)-1; i++ {
	dataLen := len(infoDB) - 1
	for i := 0; i < dataLen; i++ {
		//大位移变化
		detail := model.BoxDataDetail{
			Key:   "位移变化量",
			Value: stringAdd(infoDB[i].Data[0].Detail[keyMap["位移"]].Value, "-"+infoDB[i+1].Data[0].Detail[keyMap["位移"]].Value),
		}
		//小位移变化
		sdetail := model.BoxDataDetail{
			Key:   "小位移变化量",
			Value: stringAdd(infoDB[i].Data[0].Detail[keyMap["小位移"]].Value, "-"+infoDB[i+1].Data[0].Detail[keyMap["小位移"]].Value),
		}
		//电量
		Adetail := model.BoxDataDetail{
			Key:   "有功电量",
			Value: stringAdd(infoDB[i].Data[0].Detail[keyMap["电量"]].Value, "-"+infoDB[i+1].Data[0].Detail[keyMap["有功电量"]].Value),
		}
		infoDB[i].Data[0].Detail = append(infoDB[i].Data[0].Detail, detail)
		infoDB[i].Data[0].Detail = append(infoDB[i].Data[0].Detail, sdetail)
		infoDB[i].Data[0].Detail = append(infoDB[i].Data[0].Detail, Adetail)
		info = append(info, RetMinData2{
			CreateTime: infoDB[i].CreateTime,
			Data:       infoDB[i].Data[0].Detail,
		})

	}

	//大位移变化量索引，小位移变化量索引，有功电量索引
	keyMap["位移变化量"] = len(infoDB[0].Data[0].Detail) - 3
	keyMap["小位移变化量"] = len(infoDB[0].Data[0].Detail) - 2
	keyMap["有功电量"] = 4
	//获取设备状态
	var device model.Device
	err = global.DeviceColl.FindOne(context.TODO(), bson.M{"code": code}).Decode(&device)
	if err != nil {
		return
	}

	//websocket推送echarts
	var resData [][]string
	resData = append(resData, []string{"createTime", "有功功率", "位移", "位移变化量", "小位移", "小位移变化量", "IA(A1)", "IB(A2)", "IC(A3)", "炉阻", "直流电流", "直流电压", "直流功率", "有功电量", "档位显示", "吨/电量", "规格", "本体重量", "附属品重量"})
	if info == nil || info[0].Data == nil {
		return
	}

	//添加数据
	for i := 0; i < len(info)-1; i++ {
		if len(info[i].Data) < 55 {
			continue
		}
		resData = append(resData, []string{
			info[i].CreateTime,
			info[i].Data[keyMap["有功功率"]].Value,
			info[i].Data[keyMap["位移"]].Value,
			info[i].Data[keyMap["位移变化量"]].Value,
			info[i].Data[keyMap["小位移"]].Value,
			info[i].Data[keyMap["小位移变化量"]].Value,
			info[i].Data[keyMap["IA(A1)"]].Value,
			info[i].Data[keyMap["IB(A2)"]].Value,
			info[i].Data[keyMap["IC(A3)"]].Value,
			info[i].Data[keyMap["炉阻"]].Value,
			info[i].Data[keyMap["直流电流"]].Value,
			info[i].Data[keyMap["直流电压"]].Value,
			info[i].Data[keyMap["直流功率"]].Value,
			info[i].Data[keyMap["有功电量"]].Value,
			info[i].Data[keyMap["档位显示"]].Value,
			info[i].Data[keyMap["吨/电量"]].Value,
			info[i].Data[keyMap["规格"]].Value,
			info[i].Data[keyMap["本体重量"]].Value,
			info[i].Data[keyMap["附属品重量"]].Value,
		})
	}

	UnitMap := make(map[string]string) //记录单位
	UnitMap["有功功率"] = info[0].Data[keyMap["有功功率"]].Unit
	UnitMap["位移变化量"] = info[0].Data[keyMap["位移变化量"]].Unit
	UnitMap["小位移变化量"] = info[0].Data[keyMap["小位移变化量"]].Unit
	UnitMap["IA(A1)"] = info[0].Data[keyMap["IA(A1)"]].Unit
	UnitMap["IB(A2)"] = info[0].Data[keyMap["IB(A2)"]].Unit
	UnitMap["IC(A3)"] = info[0].Data[keyMap["IC(A3)"]].Unit
	UnitMap["位移"] = info[0].Data[keyMap["位移"]].Unit
	UnitMap["小位移"] = info[0].Data[keyMap["小位移"]].Unit
	UnitMap["炉阻"] = info[0].Data[keyMap["炉阻"]].Unit
	UnitMap["直流电流"] = info[0].Data[keyMap["直流电流"]].Unit
	UnitMap["直流电压"] = info[0].Data[keyMap["直流电压"]].Unit
	UnitMap["直流功率"] = info[0].Data[keyMap["直流功率"]].Unit
	UnitMap["有功电量"] = info[0].Data[keyMap["有功电量"]].Unit
	UnitMap["档位显示"] = info[0].Data[keyMap["档位显示"]].Unit
	UnitMap["吨/电量"] = info[0].Data[keyMap["吨/电量"]].Unit
	UnitMap["规格"] = info[0].Data[keyMap["规格"]].Unit
	UnitMap["本体重量"] = info[0].Data[keyMap["本体重量"]].Unit
	UnitMap["附属品重量"] = info[0].Data[keyMap["附属品重量"]].Unit

	var data model.Echarts
	data.Name = name
	data.Data = resData
	data.Code = code + "_" + interval
	data.UpdateTime = endTime
	data.Unit = UnitMap
	err = global.EchartsColl.FindOneAndUpdate(context.TODO(), bson.M{"code": code + "_" + interval}, bson.M{"$set": data}).Decode(&bson.M{})
	if err == mongo.ErrNoDocuments { //第一次插入
		data.CreateTime = data.UpdateTime
		data.UpdateTime = ""
		_, err = global.EchartsColl.InsertOne(context.TODO(), data)
		if err != nil {
			fmt.Println(code+"_"+interval+" echarts创建失败", err)
			return
		}
		fmt.Println(code + "_" + interval + " echarts创建成功")
	} else {
		fmt.Println(code + "_" + interval + " echarts更新成功")
	}
}
