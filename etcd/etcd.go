package etcd

import (
	"context"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
)

type Etcd struct {
	client           *clientv3.Client
	discovery        map[string]string
	locker           *sync.RWMutex
	*ServiceRegister // 租约注册
}

func NewEtcd(addr []string, userName, password string, op ...EdOption) (*Etcd, error) {
	conf := clientv3.Config{
		Endpoints:   addr,
		DialTimeout: 5 * time.Second,
		Username:    userName,
		Password:    password,
	}
	if client, err := clientv3.New(conf); err == nil {
		return &Etcd{
			client:    client,
			discovery: map[string]string{},
			locker:    &sync.RWMutex{},
		}, nil
	} else {
		return nil, err
	}
}

func (d *Etcd) Put(key, value string, opts ...clientv3.OpOption) (resp *clientv3.PutResponse, err error) {
	resp, err = d.client.Put(context.Background(), key, value, opts...)
	return resp, err
}

func (d *Etcd) Get(key string, opts ...clientv3.OpOption) (resp *clientv3.GetResponse, err error) {
	resp, err = d.client.Get(context.Background(), key, opts...)
	return resp, err
}
func (d *Etcd) Del(key string, opts ...clientv3.OpOption) (resp *clientv3.DeleteResponse, err error) {
	resp, err = d.client.Delete(context.Background(), key, opts...)
	return resp, err
}
