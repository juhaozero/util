package etcd

import (
	"context"
	"log/slog"
	"maps"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

// ServiceEventType 服务变更类型。
type ServiceEventType int

const (
	ServicePut ServiceEventType = iota
	ServiceDelete
)

// ServiceEvent 服务发现变更事件。
type ServiceEvent struct {
	Type  ServiceEventType
	Key   string
	Value string
}

// ServiceChangeHandler 服务变更回调。
type ServiceChangeHandler func(event ServiceEvent)

// GetDiscovery 按前缀获取服务列表，并确保只启动一个 watcher。
func (ed *Etcd) GetDiscovery(prefix string, op ...clientv3.OpOption) ([]string, error) {
	if err := ed.ensureOpen(); err != nil {
		return nil, err
	}

	opts := append([]clientv3.OpOption{clientv3.WithPrefix()}, op...)
	resp, err := ed.GetCtx(context.Background(), prefix, opts...)
	if err != nil {
		return nil, err
	}

	addrs := ed.extractAddrs(resp)
	ed.ensureWatcher(prefix)
	return addrs, nil
}

// WatchServices 监听指定前缀下的服务变更。
func (ed *Etcd) WatchServices(ctx context.Context, prefix string, handler ServiceChangeHandler) error {
	if err := ed.ensureOpen(); err != nil {
		return err
	}
	if handler != nil {
		ed.locker.Lock()
		prev := ed.onChange
		ed.onChange = func(event ServiceEvent) {
			if prev != nil {
				prev(event)
			}
			handler(event)
		}
		ed.locker.Unlock()
	}
	ed.ensureWatcherWithContext(ctx, prefix)
	return nil
}

// StopWatch 停止指定前缀的 watcher。
func (ed *Etcd) StopWatch(prefix string) {
	ed.watchMu.Lock()
	defer ed.watchMu.Unlock()
	if cancel, ok := ed.watchCancel[prefix]; ok {
		cancel()
		delete(ed.watchCancel, prefix)
	}
}

func (ed *Etcd) ensureWatcher(prefix string) {
	ed.ensureWatcherWithContext(context.Background(), prefix)
}

func (ed *Etcd) ensureWatcherWithContext(parent context.Context, prefix string) {
	ed.watchMu.Lock()
	defer ed.watchMu.Unlock()
	if _, ok := ed.watchCancel[prefix]; ok {
		return
	}

	ctx, cancel := context.WithCancel(parent)
	ed.watchCancel[prefix] = cancel
	go ed.watcher(ctx, prefix)
}

func (ed *Etcd) watcher(ctx context.Context, prefix string) {
	rch := ed.client.Watch(ctx, prefix, clientv3.WithPrefix())
	for wresp := range rch {
		if wresp.Err() != nil {

		}
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				ed.setServiceList(string(ev.Kv.Key), string(ev.Kv.Value))
			case mvccpb.DELETE:
				ed.delServiceList(string(ev.Kv.Key))
			}
		}
	}

	ed.watchMu.Lock()
	if cancel, ok := ed.watchCancel[prefix]; ok && cancel != nil {
		delete(ed.watchCancel, prefix)
	}
	ed.watchMu.Unlock()
}

func (ed *Etcd) extractAddrs(resp *clientv3.GetResponse) []string {
	addrs := make([]string, 0)
	if resp == nil || resp.Kvs == nil {
		return addrs
	}
	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			ed.setServiceList(string(resp.Kvs[i].Key), string(v))
			addrs = append(addrs, string(v))
		}
	}
	return addrs
}

func (ed *Etcd) setServiceList(key, val string) {
	ed.locker.Lock()
	ed.discovery[key] = val
	handler := ed.onChange
	ed.locker.Unlock()
	slog.Info("setServiceList", "key", key, "val", val)
	if handler != nil {
		handler(ServiceEvent{Type: ServicePut, Key: key, Value: val})
	}
}

func (ed *Etcd) delServiceList(key string) {
	ed.locker.Lock()
	delete(ed.discovery, key)
	handler := ed.onChange
	ed.locker.Unlock()
	if handler != nil {
		handler(ServiceEvent{Type: ServiceDelete, Key: key})
	}
}

// GetServiceList 获取当前服务发现快照。
func (ed *Etcd) GetServiceList() map[string]string {
	ed.locker.RLock()
	defer ed.locker.RUnlock()
	tmp := make(map[string]string, len(ed.discovery))
	maps.Copy(tmp, ed.discovery)
	return tmp
}

// GetServiceValues 获取所有服务地址值。
func (ed *Etcd) GetServiceValues() []string {
	ed.locker.RLock()
	defer ed.locker.RUnlock()
	values := make([]string, 0, len(ed.discovery))
	for _, v := range ed.discovery {
		values = append(values, v)
	}
	return values
}
