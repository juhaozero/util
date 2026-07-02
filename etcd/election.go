package etcd

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
)

// LeaderElection Leader 选举封装，负责 session 生命周期。
type LeaderElection struct {
	Election *concurrency.Election
	session  *concurrency.Session
}

// Close 释放 Leader 身份并关闭 session。
func (le *LeaderElection) Close(ctx context.Context) error {
	if le == nil {
		return nil
	}
	var err error
	if le.Election != nil {
		if resignErr := le.Election.Resign(ctx); resignErr != nil && err == nil {
			err = resignErr
		}
	}
	if le.session != nil {
		if closeErr := le.session.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return err
}

// Campaign 参与 Leader 选举。
func (e *Etcd) Campaign(ctx context.Context, prefix, proposal string) (*LeaderElection, error) {
	if err := e.ensureOpen(); err != nil {
		return nil, err
	}

	session, err := concurrency.NewSession(e.client)
	if err != nil {
		return nil, fmt.Errorf("create etcd session: %w", err)
	}

	election := concurrency.NewElection(session, prefix)
	if err := election.Campaign(ctx, proposal); err != nil {
		session.Close()
		return nil, fmt.Errorf("campaign leader: %w", err)
	}

	return &LeaderElection{Election: election, session: session}, nil
}

// ObserveLeader 监听 Leader 变更，返回当前 Leader proposal。
func (e *Etcd) ObserveLeader(ctx context.Context, prefix string) (string, error) {
	if err := e.ensureOpen(); err != nil {
		return "", err
	}

	session, err := concurrency.NewSession(e.client)
	if err != nil {
		return "", fmt.Errorf("create etcd session: %w", err)
	}
	defer session.Close()

	election := concurrency.NewElection(session, prefix)
	ch := election.Observe(ctx)

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case resp, ok := <-ch:
		if !ok || len(resp.Kvs) == 0 {
			return "", fmt.Errorf("leader observation closed")
		}
		return string(resp.Kvs[0].Value), nil
	}
}

// ResignLeader 主动放弃 Leader 身份。
func ResignLeader(ctx context.Context, le *LeaderElection) error {
	if le == nil {
		return fmt.Errorf("leader election is nil")
	}
	return le.Close(ctx)
}

// WithLeader 成为 Leader 后执行 fn，函数返回或 ctx 取消时自动释放 Leader。
func (e *Etcd) WithLeader(ctx context.Context, prefix, proposal string, fn func(ctx context.Context) error) error {
	if fn == nil {
		return fmt.Errorf("leader callback is nil")
	}

	le, err := e.Campaign(ctx, prefix, proposal)
	if err != nil {
		return err
	}
	defer func() {
		resignCtx := context.WithoutCancel(ctx)
		if resignCtx.Err() != nil {
			resignCtx = context.Background()
		}
		_ = le.Close(resignCtx)
	}()

	return fn(ctx)
}
