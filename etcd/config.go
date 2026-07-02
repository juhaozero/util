package etcd

import "time"

// Config etcd 客户端配置。
type Config struct {
	Endpoints   []string
	DialTimeout time.Duration
	Username    string
	Password    string
}

func (c Config) withDefaults() Config {
	if c.DialTimeout <= 0 {
		c.DialTimeout = 5 * time.Second
	}
	return c
}
