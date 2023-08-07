package mongo

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

// 插入一个
func InsertOne[T Bson](collection *mongo.Collection, data T) (string, error) {
	var id string
	result, err := collection.InsertOne(contex, data)
	if err != nil {
		return id, err
	}
	id = result.InsertedID.(string)
	return id, nil
}

// 插入多个
func InsertMany[T Bson](collection *mongo.Collection, data []T) ([]string, error) {
	var ids = []string{}
	result, err := collection.InsertMany(contex, []interface{}{data}, options.InsertMany().SetOrdered(false))
	if err != nil {
		return ids, err
	}
	for _, v := range result.InsertedIDs {
		ids = append(ids, v.(string))
	}

	return ids, nil
}

// 查找全部排序
func FindSort[T Bson](collection *mongo.Collection, data T, sort int, opts ...*options.FindOptions) ([]bson.M, error) {
	var result []bson.M
	// 1升序或者-1降序
	find := options.Find().SetSort(sort)
	opts = append(opts, find)

	cursor, err := collection.Find(contex, data, opts...)
	if err != nil {
		return result, err
	}

	if err := cursor.All(contex, &result); err != nil {
		return result, err
	}

	return result, nil
}
