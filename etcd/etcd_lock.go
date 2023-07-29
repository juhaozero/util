package etcd

import (
	"context"
	"log"

	"github.com/coreos/etcd/clientv3/concurrency"
)

// EtcdLock 上锁处理
func (e *Etcd) EtcdLock(key string, times int, f func()) {
	s, _ := concurrency.NewSession(e.client, concurrency.WithTTL(times))
	defer s.Close()

	mu := concurrency.NewMutex(s, key)
	if err := mu.Lock(context.TODO()); err != nil {
		log.Fatal("m lock err: ", err)
	}

	//do something
	f()

	if err := mu.Unlock(context.TODO()); err != nil {
		log.Fatal("m unlock err: ", err)
	}
}
