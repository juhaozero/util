package etcd

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
)

// WithLock 获取分布式锁后执行 fn，支持 context 超时与错误返回。
func (e *Etcd) WithLock(ctx context.Context, key string, ttl int, fn func() error) error {
	if err := e.ensureOpen(); err != nil {
		return err
	}
	if fn == nil {
		return fmt.Errorf("lock callback is nil")
	}

	session, err := concurrency.NewSession(e.client, concurrency.WithTTL(ttl))
	if err != nil {
		return fmt.Errorf("create etcd session: %w", err)
	}
	defer session.Close()

	mu := concurrency.NewMutex(session, key)
	if err := mu.Lock(ctx); err != nil {
		return fmt.Errorf("acquire etcd lock: %w", err)
	}

	defer func() {
		unlockCtx := context.WithoutCancel(ctx)
		if unlockCtx.Err() != nil {
			unlockCtx = context.Background()
		}
		_ = mu.Unlock(unlockCtx)
	}()

	return fn()
}
