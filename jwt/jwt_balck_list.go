package jwt

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

type JwtBlackList struct {
	Store Store
	Key   string
}

// Store 缓存接口
type Store interface {
	Get(key string) (string, error)
	Set(key string, value string, expiration time.Duration) error
	Del(key string) error
}

// NewJwtBlackList 创建黑名单
func NewJwtBlackList(store Store, key string) *JwtBlackList {
	return &JwtBlackList{
		Store: store,
		Key:   key,
	}
}

// blacklistKey 生成黑名单 key
// 使用 sha256 加密 token 并拼接 key 避免 key 过长
func (j *JwtBlackList) blacklistKey(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%s:%x", j.Key, hash[:])
}

// AddBlackList 添加黑名单
func (j *JwtBlackList) AddBlackList(token string, expiration time.Duration) error {
	return j.Store.Set(j.blacklistKey(token), token, expiration)
}

// IsBlackList 判断 token 是否在黑名单中
func (j *JwtBlackList) IsRevoked(token string) (bool, error) {
	val, err := j.Store.Get(j.blacklistKey(token))
	if err != nil {
		return false, err
	}
	return val != "", nil
}

// DelBlackList 删除黑名单
func (j *JwtBlackList) DelBlackList(token string) error {
	return j.Store.Del(j.blacklistKey(token))
}

// 本地缓存
func NewLocalStore() *LocalStore {
	return &LocalStore{
		CacheList: make(map[string]string),
	}
}

type LocalStore struct {
	CacheList map[string]string
	sync.RWMutex
}

func (d *LocalStore) Get(key string) (string, error) {
	d.RLock()
	defer d.RUnlock()
	return d.CacheList[key], nil
}

func (d *LocalStore) Set(key string, value string, expiration time.Duration) error {
	d.Lock()
	defer d.Unlock()
	d.CacheList[key] = value
	return nil
}

func (d *LocalStore) Del(key string) error {
	d.Lock()
	defer d.Unlock()
	delete(d.CacheList, key)
	return nil
}
