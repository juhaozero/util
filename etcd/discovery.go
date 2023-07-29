package etcd

import (
	"context"
	"log"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

// GetDiscovery 获取所有指定服务键的值
// op： etcd 的组件
func (ed *Etcd) GetDiscovery(prefix string, op ...clientv3.OpOption) ([]string, error) {
	resp, err := ed.client.Get(context.Background(), prefix, op...)
	if err != nil {
		return nil, err
	}
	addrs := ed.extractAddrs(resp)

	go ed.watcher(prefix)
	return addrs, nil
}

// watcher 监听订阅
func (ed *Etcd) watcher(prefix string) {
	rch := ed.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				// 修改或者新增
				ed.setServiceList(string(ev.Kv.Key), string(ev.Kv.Value))
			case mvccpb.DELETE:
				ed.delServiceList(string(ev.Kv.Key))
			}
		}
	}
}

// extractAddrs 获取对应的值
func (ed *Etcd) extractAddrs(resp *clientv3.GetResponse) []string {
	addrs := make([]string, 0)
	if resp == nil || resp.Kvs == nil {
		return addrs
	}
	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			ed.setServiceList(string(resp.Kvs[i].Key), string(resp.Kvs[i].Value))
			addrs = append(addrs, string(v))
		}
	}
	return addrs
}

// SetServiceList 设置服务发现
func (ed *Etcd) setServiceList(key, val string) {
	ed.locker.Lock()
	defer ed.locker.Unlock()
	ed.discovery[key] = val
	log.Printf("发现服务 key=%s,value=%s", key, val)
}

// DelServiceList 删除服务发现
func (ed *Etcd) delServiceList(key string) {
	ed.locker.Lock()
	defer ed.locker.Unlock()
	delete(ed.discovery, key)
	log.Printf("服务下线: %s", key)
}

// GetServiceList 获取服务发现的值
func (ed *Etcd) GetServiceList() map[string]string {
	tmp := map[string]string{}
	ed.locker.RLock()
	defer ed.locker.RUnlock()
	for k, v := range ed.discovery {
		tmp[k] = v
	}
	return tmp
}
