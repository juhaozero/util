package retry

import (
	"github.com/avast/retry-go"
)

type RetryClient struct {
	retryStrategy []retry.Option
}

var (
	RetryStrategy = []retry.Option{}
)

// 重试配置
func NewRetryOption(retryStrategy []retry.Option) *RetryClient {
	return &RetryClient{
		retryStrategy: retryStrategy,
	}
}

// 重试请求
func (r *RetryClient) RetryFuncInterface(f func() error) error {
	err := retry.Do(f, r.retryStrategy...)
	if err != nil {
		return err
	}
	return nil
}
