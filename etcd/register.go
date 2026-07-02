package etcd

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/juhaozero/util/log"
	"go.uber.org/zap"
)

var (
	ErrRegistrarClosed = errors.New("service registrar is closed")
)

type serviceEntry struct {
	key string
	val string
}

// ServiceRegister 基于租约的服务注册器。
type ServiceRegister struct {
	client *clientv3.Client
	ttl    int64

	mu            sync.Mutex
	lease         clientv3.Lease
	leaseResp     *clientv3.LeaseGrantResponse
	cancelFunc    context.CancelFunc
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	services      []serviceEntry
	closed        bool
}

// WithRegister 初始化时创建租约注册器。
func WithRegister(ttl int64) EdOption {
	return func(e *Etcd) error {
		_, err := e.NewServiceRegister(ttl)
		return err
	}
}

// NewServiceRegister 创建并绑定租约注册器。
func (d *Etcd) NewServiceRegister(ttl int64) (*ServiceRegister, error) {
	if err := d.ensureOpen(); err != nil {
		return nil, err
	}

	sr := &ServiceRegister{
		client: d.client,
		ttl:    ttl,
	}
	if err := sr.grantLease(context.Background()); err != nil {
		return nil, err
	}

	d.ServiceRegister = sr
	go sr.listenLeaseKeepAlive()
	return sr, nil
}

// NewRegistrar 创建独立的租约注册器。
func NewRegistrar(client *clientv3.Client, ttl int64) (*ServiceRegister, error) {
	if client == nil {
		return nil, errors.New("etcd client is nil")
	}
	sr := &ServiceRegister{
		client: client,
		ttl:    ttl,
	}
	if err := sr.grantLease(context.Background()); err != nil {
		return nil, err
	}
	go sr.listenLeaseKeepAlive()
	return sr, nil
}

// grantLease 授予租约并监听续期
func (sr *ServiceRegister) grantLease(ctx context.Context) error {
	lease := clientv3.NewLease(sr.client)

	grantCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	leaseResp, err := lease.Grant(grantCtx, sr.ttl)
	if err != nil {
		return err
	}

	keepCtx, cancelFunc := context.WithCancel(context.Background())
	keepAliveChan, err := lease.KeepAlive(keepCtx, leaseResp.ID)
	if err != nil {
		cancelFunc()
		return err
	}

	sr.mu.Lock()
	defer sr.mu.Unlock()

	if sr.cancelFunc != nil {
		sr.cancelFunc()
	}

	sr.lease = lease
	sr.leaseResp = leaseResp
	sr.cancelFunc = cancelFunc
	sr.keepAliveChan = keepAliveChan
	return nil
}

// listenLeaseKeepAlive 监听租约续期
func (sr *ServiceRegister) listenLeaseKeepAlive() {
	for {
		sr.mu.Lock()
		ch := sr.keepAliveChan
		closed := sr.closed
		leaseID := clientv3.NoLease
		if sr.leaseResp != nil {
			leaseID = sr.leaseResp.ID
		}
		sr.mu.Unlock()

		if closed {
			return
		}

		resp, ok := <-ch
		if !ok || resp == nil {
			log.Default().Warn("etcd 租约续期中断，尝试重建", zap.Int64("leaseID", int64(leaseID)))
			if err := sr.rebuildLease(); err != nil {
				log.Default().Error("etcd 租约重建失败", zap.Error(err))
				time.Sleep(time.Second)
			}
		}
	}
}

// rebuildLease 重建租约
func (sr *ServiceRegister) rebuildLease() error {
	sr.mu.Lock()
	if sr.closed {
		sr.mu.Unlock()
		return ErrRegistrarClosed
	}
	services := append([]serviceEntry(nil), sr.services...)
	sr.mu.Unlock()

	if err := sr.grantLease(context.Background()); err != nil {
		return err
	}

	for _, item := range services {
		if err := sr.putServiceLocked(item.key, item.val); err != nil {
			return err
		}
	}
	return nil
}

// PutService 注册服务到 etcd。
func (sr *ServiceRegister) PutService(key, val string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	return sr.putServiceLocked(key, val)
}

func (sr *ServiceRegister) putServiceLocked(key, val string) error {
	if sr.closed {
		return ErrRegistrarClosed
	}
	if sr.leaseResp == nil {
		return errors.New("lease is not ready")
	}

	kv := clientv3.NewKV(sr.client)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := kv.Put(ctx, key, val, clientv3.WithLease(sr.leaseResp.ID)); err != nil {
		return err
	}

	sr.upsertService(key, val)
	return nil
}

// upsertService 更新或添加服务
func (sr *ServiceRegister) upsertService(key, val string) {
	for i, item := range sr.services {
		if item.key == key {
			sr.services[i].val = val
			return
		}
	}
	sr.services = append(sr.services, serviceEntry{key: key, val: val})
}

// PutServiceCtx 带 context 的服务注册。
func (sr *ServiceRegister) PutServiceCtx(ctx context.Context, key, val string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if sr.closed {
		return ErrRegistrarClosed
	}
	if sr.leaseResp == nil {
		return errors.New("lease is not ready")
	}

	kv := clientv3.NewKV(sr.client)
	if _, err := kv.Put(ctx, key, val, clientv3.WithLease(sr.leaseResp.ID)); err != nil {
		return err
	}
	sr.upsertService(key, val)
	return nil
}

// RevokeLease 撤销租约并下线服务。
func (sr *ServiceRegister) RevokeLease() error {
	return sr.Close(context.Background())
}

// Close 停止续租并撤销租约。
func (sr *ServiceRegister) Close(ctx context.Context) error {
	sr.mu.Lock()
	if sr.closed {
		sr.mu.Unlock()
		return nil
	}
	sr.closed = true

	cancelFunc := sr.cancelFunc
	lease := sr.lease
	leaseID := clientv3.NoLease
	if sr.leaseResp != nil {
		leaseID = sr.leaseResp.ID
	}
	sr.mu.Unlock()

	if cancelFunc != nil {
		cancelFunc()
	}

	if lease == nil || leaseID == clientv3.NoLease {
		return nil
	}

	revokeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := lease.Revoke(revokeCtx, leaseID)
	return err
}

// LeaseID 返回当前租约 ID。
func (sr *ServiceRegister) LeaseID() clientv3.LeaseID {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	if sr.leaseResp == nil {
		return clientv3.NoLease
	}
	return sr.leaseResp.ID
}
