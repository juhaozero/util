package etcd

import (
	"context"
	"errors"
	"sync"

	"github.com/coreos/etcd/clientv3"
)

var ErrClosed = errors.New("etcd client is closed")

// Etcd etcd 客户端封装。
type Etcd struct {
	client    *clientv3.Client
	discovery map[string]string
	locker    sync.RWMutex

	watchMu     sync.Mutex
	watchCancel map[string]context.CancelFunc
	onChange    ServiceChangeHandler

	closeOnce sync.Once
	closeErr  error
	closed    bool

	*ServiceRegister // 租约注册，兼容旧用法
}

// EdOption 初始化选项。
type EdOption func(*Etcd) error

// New 创建 etcd 客户端。
func New(cfg Config, opts ...EdOption) (*Etcd, error) {
	cfg = cfg.withDefaults()
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: cfg.DialTimeout,
		Username:    cfg.Username,
		Password:    cfg.Password,
	})
	if err != nil {
		return nil, err
	}

	e := &Etcd{
		client:      client,
		discovery:   map[string]string{},
		watchCancel: map[string]context.CancelFunc{},
	}

	for _, opt := range opts {
		if err := opt(e); err != nil {
			_ = e.Close(context.Background())
			return nil, err
		}
	}
	return e, nil
}

// Client 返回底层 clientv3.Client。
func (e *Etcd) Client() *clientv3.Client {
	return e.client
}

// OnServiceChange 设置服务变更回调，供服务发现使用。
func (e *Etcd) OnServiceChange(handler ServiceChangeHandler) {
	e.locker.Lock()
	defer e.locker.Unlock()
	e.onChange = handler
}

// Close 关闭客户端、停止所有 watcher，并撤销租约。
func (e *Etcd) Close(ctx context.Context) error {
	e.closeOnce.Do(func() {
		e.stopAllWatchers()

		if e.ServiceRegister != nil {
			e.closeErr = e.ServiceRegister.Close(ctx)
		}

		if e.client != nil {
			if err := e.client.Close(); err != nil && e.closeErr == nil {
				e.closeErr = err
			}
		}
		e.closed = true
	})
	return e.closeErr
}

func (e *Etcd) ensureOpen() error {
	if e == nil || e.closed {
		return ErrClosed
	}
	return nil
}

// Ping 检查 etcd 连接是否正常。
func (e *Etcd) Ping(ctx context.Context) error {
	if err := e.ensureOpen(); err != nil {
		return err
	}
	endpoints := e.client.Endpoints()
	if len(endpoints) == 0 {
		return errors.New("no etcd endpoints configured")
	}
	_, err := e.client.Status(ctx, endpoints[0])
	return err
}

func (e *Etcd) Put(key, value string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return e.PutCtx(context.Background(), key, value, opts...)
}

func (e *Etcd) PutCtx(ctx context.Context, key, value string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	if err := e.ensureOpen(); err != nil {
		return nil, err
	}
	return e.client.Put(ctx, key, value, opts...)
}

func (e *Etcd) Get(key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return e.GetCtx(context.Background(), key, opts...)
}

func (e *Etcd) GetCtx(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	if err := e.ensureOpen(); err != nil {
		return nil, err
	}
	return e.client.Get(ctx, key, opts...)
}

func (e *Etcd) Del(key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	return e.DelCtx(context.Background(), key, opts...)
}

func (e *Etcd) DelCtx(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	if err := e.ensureOpen(); err != nil {
		return nil, err
	}
	return e.client.Delete(ctx, key, opts...)
}

// Txn 创建事务。
func (e *Etcd) Txn(ctx context.Context) clientv3.Txn {
	return e.client.Txn(ctx)
}

// stopAllWatchers 停止所有 watcher
func (e *Etcd) stopAllWatchers() {
	e.watchMu.Lock()
	defer e.watchMu.Unlock()
	for prefix, cancel := range e.watchCancel {
		cancel()
		delete(e.watchCancel, prefix)
	}
}
