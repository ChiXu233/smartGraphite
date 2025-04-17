package model

type FormChargerReportEndTime struct {
	CreateTime string `json:"createTime" bson:"createTime"` //初次创建时间
	UpdateTime string `json:"updateTime" bson:"updateTime"` //更新时间
	EndTime    string `json:"endTime" bson:"endTime"`       //计算结束时间
}
