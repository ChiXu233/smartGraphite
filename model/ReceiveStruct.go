package model

// 接收返回的数据类型
// 接收获取token时返回的数据
type Token struct {
	Code int      `form:"code" json:"code"` //状态码
	Data struct { //数据集，状态success为true时返回，否则为空
		Expire int    `form:"expire" json:"expire"`
		Token  string `form:"token" json:"token"`
		Type   string `form:"type" json:"type"`
	} `form:"data" json:"data"`
	Msg     string `form:"msg" json:"msg"`         //返回消息
	Success bool   `form:"success" json:"success"` //状态
}

// 接收获取项目时返回的数据
type Project struct {
	Code int `form:"code" json:"code"`
	Data []struct {
		Id          string `form:"id" json:"id"`
		Name        string `form:"name" json:"name"`
		ProjectType string `form:"projectType" json:"projectType"`
		Status      string `form:"status" json:"status"` //状态 0：离线，1：在线
	} `form:"data" json:"data"`
	Msg     string `form:"msg" json:"msg"`
	Success bool   `form:"success" json:"success"`
}

// 接收获取项目Box时返回的数据
type ProjectBox struct {
	Code    int       `form:"code" json:"code"`
	Data    []BoxType `form:"data" json:"data"`
	Msg     string    `form:"msg" json:"msg"`
	Success bool      `form:"success" json:"success"`
}
type BoxType struct {
	BoxId       string `form:"boxId" json:"boxId"`
	Name        string `form:"name" json:"name"`
	ProjectType string `form:"projectType" json:"projectType"`
	Serlnum     string `form:"serlnum" json:"serlnum"`
	Status      string `form:"status" json:"status"`
}

// 接收获取项目Box Plc时返回的数据
type BoxPlc struct {
	Code    int       `form:"code" json:"code"`
	Data    []PlcType `form:"data" json:"data"`
	Msg     string    `form:"msg" json:"msg"`
	Success bool      `form:"success" json:"success"`
}
type PlcType struct {
	PlcId  string `form:"plcId" json:"plcId"`
	Name   string `form:"name" json:"name"`
	Status string `form:"status" json:"status"`
}

// 接收获取项目变量时返回的数据
type BoxVariant struct {
	Code    int           `form:"code" json:"code"`
	Data    []VariantType `form:"data" json:"data"`
	Msg     string        `form:"msg" json:"msg"`
	Success bool          `form:"success" json:"success"`
}
type VariantType struct {
	Addr      string `form:"addr" json:"addr"`
	Name      string `form:"name" json:"name"`
	Type      string `form:"type" json:"type"`
	VariantId string `form:"variantId" json:"variantId"`
}

// 接收实时数据变量时返回的数据
type RealtimeData struct {
	Code    int              `form:"code" json:"code"`
	Data    []RealtimeDataVo `form:"data" json:"data"`
	Msg     string           `form:"msg" json:"msg"`
	Success bool             `form:"success" json:"success"`
}
type RealtimeDataVo struct {
	Id     string `form:"id" json:"id"`
	Status string `form:"status" json:"status"`
	Time   int    `form:"time" json:"time"` //单位s
	Value  string `form:"value" json:"value"`
}

// 接收修改变量时返回的数据
type WriteData struct {
	Code    int         `form:"id" json:"id"`
	Success bool        `form:"success" json:"success"`
	Msg     string      `form:"msg" json:"msg"`
	Data    WriteDataVo `form:"data" json:"data"`
}
type WriteDataVo struct {
	VariantId     string `form:"variantId" json:"variantId"`
	ReturnSuccess bool   `form:"returnsuccess" json:"returnsuccess"`
	Msg           string `form:"msg" json:"msg"`
}
