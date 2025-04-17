package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"SmartGraphite-server/utils"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
)

func CreateDeviceType(deviceType model.DeviceType) utils.Response {
	n, err := global.DeviceTypeColl.InsertOne(context.TODO(), deviceType)
	if err != nil {
		fmt.Println(err.Error())
	}
	return utils.SuccessMess("", n)
}
func GetDeviceType() utils.Response {
	var deviceType []bson.M
	if err := utils.Find(global.DeviceTypeColl, &deviceType, bson.M{}); err != nil {
		return utils.ErrorMess("查找失败", err.Error())
	}
	return utils.SuccessMess("查找成功", deviceType)
}
