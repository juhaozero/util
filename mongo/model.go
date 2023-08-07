package mongo

import "go.mongodb.org/mongo-driver/bson"

type Config struct {
	UserName string
	PassWord string
	Addr     string
	Port     int32
}
type Bson interface {
	bson.D | bson.M | bson.A
}

const (
	Desc = -1
	Aes  = 1
)
