package utils

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Find 查询
func Find(collection *mongo.Collection, result interface{}, filter interface{}, opts ...*options.FindOptions) error {
	cur, err := collection.Find(context.TODO(), filter, opts...)
	if err != nil {
		return err
	}
	if err = cur.All(context.TODO(), result); err != nil {
		return err
	}
	return nil
}

