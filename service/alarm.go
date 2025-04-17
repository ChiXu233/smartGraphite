package service

//func GraphitePubAlarm(device model.Device, power string) {
//	now := time.Now()
//	var one model.Box
//	err := global.GraphitingDataColl.FindOne(context.TODO(), bson.M{}).Decode(&one)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	//上一个有功功率
//	lastPower := one.Data[0].Detail[3].Value
//	lastIntPower, err := strconv.Atoi(lastPower)
//	if err != nil {
//		fmt.Println(err)
//	}
//	//现在有功功率
//	nowIntPower, err := strconv.Atoi(power)
//	if err != nil {
//		fmt.Println(err)
//	}
//	//送电：现在有功功率大于一 前一个小于一
//	if nowIntPower > 1 && lastIntPower < 1 {
//		alarm := model.Alarm{
//			Device:  device,
//			Type:    "生产",
//			Content: "送电:有功功率大于1",
//			Time:    now.Format("2006-01-02 15:04:05"),
//		}
//		alarmData, err := json.Marshal(alarm)
//		if err != nil {
//			return
//		}
//		global.MqttPubAlarm.Publish("alarm", 2, false, alarmData)
//	}
//	//停电：现在有功功率小于一 前一个大于一
//	if nowIntPower < 1 && lastIntPower > 1 {
//		alarm := model.Alarm{
//			Device:  device,
//			Type:    "生产",
//			Content: "停电:有功功率小于1",
//			Time:    now.Format("2006-01-02 15:04:05"),
//		}
//		alarmData, err := json.Marshal(alarm)
//		if err != nil {
//			return
//		}
//		global.MqttPubAlarm.Publish("alarm", 2, false, alarmData)
//	}
//	return
//}
