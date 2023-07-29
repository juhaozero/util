package etcd

import (
	"context"
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
)

// 注册租约服务
type ServiceRegister struct {
	client        *clientv3.Client
	lease         clientv3.Lease //租约
	leaseResp     *clientv3.LeaseGrantResponse
	canclefunc    func()
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
}

// 注册
func WithRegister(timeNum int64) EdOption {
	return func(e *Etcd) { e.NewServiceRegister(timeNum) }
}

// NewServiceReg 租约服务
func (d *Etcd) NewServiceRegister(timeNum int64) (*ServiceRegister, error) {
	ser := &ServiceRegister{
		client: d.client,
	}
	// 注册
	d.ServiceRegister = ser

	if err := ser.setLease(timeNum); err != nil {
		return nil, err
	}
	go ser.listenLeaseRespChan()
	return ser, nil
}

// setLease 设置租约
func (sr *ServiceRegister) setLease(timeNum int64) error {
	lease := clientv3.NewLease(sr.client)

	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
	defer cancel()
	leaseResp, err := lease.Grant(ctx, timeNum)
	if err != nil {
		return err
	}

	ctx, cancelFunc := context.WithCancel(context.TODO())
	leaseRespChan, errs := lease.KeepAlive(ctx, leaseResp.ID)
	if errs != nil {
		cancelFunc()
		return errs
	}

	sr.lease = lease
	sr.leaseResp = leaseResp
	sr.canclefunc = cancelFunc
	sr.keepAliveChan = leaseRespChan
	return nil
}

// listenLeaseRespChan 监听续租情况
func (sr *ServiceRegister) listenLeaseRespChan() {
	for {
		leaseKeepResp := <-sr.keepAliveChan
		if leaseKeepResp == nil {
			log.Printf("已经关闭续租功能 %d", sr.leaseResp.ID)
			return
		}
	}

}

// 注册租约
func (sr *ServiceRegister) PutService(key, val string) error {
	kv := clientv3.NewKV(sr.client)
	_, err := kv.Put(context.TODO(), key, val, clientv3.WithLease(sr.leaseResp.ID))
	return err
}

// 撤销租约
func (sr *ServiceRegister) RevokeLease() error {
	sr.canclefunc()
	time.Sleep(2 * time.Second)
	_, err := sr.lease.Revoke(context.TODO(), sr.leaseResp.ID)
	return err
}
