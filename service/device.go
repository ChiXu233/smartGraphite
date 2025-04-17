package service

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"SmartGraphite-server/utils"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateDevice(device model.Device) utils.Response {

	var deviceType model.DeviceType
	if err := global.DeviceTypeColl.FindOne(context.TODO(), bson.M{"_id": device.DeviceTypeId}).Decode(&deviceType); err != nil {
		if err == mongo.ErrNoDocuments {
			return utils.ErrorMess("设备类型不存在", err.Error())
		} else {
			return utils.ErrorMess("查找设备类型时遇到其它错误", err.Error())
		}
	}
	//不是自定义设备
	if !device.IsCustom {
		device.Sensors = deviceType.Sensors
	}
	if res, err := global.DeviceColl.InsertOne(context.TODO(), device); err != nil {
		return utils.ErrorMess("添加失败", err.Error())
	} else {
		device.Id = res.InsertedID.(primitive.ObjectID)
		return utils.SuccessMess("添加成功", device)
	}
}
