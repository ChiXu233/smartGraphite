package global

import "go.mongodb.org/mongo-driver/mongo"

//动态协议计算时间间隔时用
var (
	CollMap map[string]map[int]*mongo.Collection //通过设备编号，时间间隔来指定表
)
