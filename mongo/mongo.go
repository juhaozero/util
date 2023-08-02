package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	client *mongo.Client
	Contex context.Context
)

func NewMangoClient(conf *Config) {
	Contex = context.Background()
	opts := NewMangoOpts(conf)
	c, err := mongo.Connect(Contex, opts)
	if err != nil {
		panic(err)
		//log.Fatal(err)
	}
	client = c
}

func NewMangoOpts(conf *Config) *options.ClientOptions {
	var clientOpts = options.Client().
		SetAuth(options.Credential{
			AuthMechanism: "SCRAM-SHA-1",
			AuthSource:    "admin",
			Username:      conf.UserName,
			Password:      conf.PassWord,
		}).
		SetConnectTimeout(10 * time.Second).
		SetHosts([]string{fmt.Sprintf("%s:%s", conf.Addr, fmt.Sprint(conf.Port))}).
		SetMaxPoolSize(20).
		SetMinPoolSize(5).
		SetReadPreference(readpref.Primary())
	//	SetReplicaSet("replicaSetDb")

	return clientOpts
}

// 获取db库
func GetClientDataBase(dbName string, opts ...*options.DatabaseOptions) *mongo.Database {
	return client.Database(dbName, opts...)
}

// 获取数据库表
// collectionName 表名
func GetDataBaseCollection(db *mongo.Database, collectionName string) *mongo.Collection {
	return db.Collection(collectionName)
}
