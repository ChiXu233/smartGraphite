package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// 石墨化原料表（规格；本体重量；附属品重量）
type GraphiteOriginList struct {
	Id           primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Name         string             `json:"name" bson:"name"`
	RealWeight   float64            `json:"realWeight" bson:"realWeight"`
	AccessWeight float64            `json:"accessWeight" bson:"accessWeight"`
}
