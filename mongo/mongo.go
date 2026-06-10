package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoClient struct {
	client *mongo.Client
	contex context.Context // 全局上下文
	db     *mongo.Database
}

func NewMangoClient(conf *Config) (*MongoClient, error) {
	ctx := context.Background()
	opts := NewMangoOpts(conf.UserName, conf.PassWord, conf.Addr, conf.Port)
	c, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
		//log.Fatal(err)
	}

	client := &MongoClient{
		client: c,
		contex: ctx,
	}
	client.db = client.GetClientDataBase(conf.DbName)
	return client, nil
}

// 初始化一个mango配置项
func NewMangoOpts(userName, passWord, addr string, port int32) *options.ClientOptions {
	var clientOpts = options.Client().
		SetAuth(options.Credential{
			AuthMechanism: "SCRAM-SHA-1",
			AuthSource:    "admin",
			Username:      userName,
			Password:      passWord,
		}).
		SetConnectTimeout(10 * time.Second).
		SetHosts([]string{fmt.Sprintf("%s:%s", addr, fmt.Sprint(port))}).
		SetMaxPoolSize(20).
		SetMinPoolSize(5).
		SetReadPreference(readpref.Primary())
	//	SetReplicaSet("replicaSetDb")

	return clientOpts
}

// 获取db库
func (m *MongoClient) GetClientDataBase(dbName string, opts ...*options.DatabaseOptions) *mongo.Database {
	return m.client.Database(dbName, opts...)
}

// 获取数据库表
// collectionName 表名
func (m *MongoClient) GetMongoCollection(collectionName string) *mongo.Collection {
	return m.db.Collection(collectionName)
}
