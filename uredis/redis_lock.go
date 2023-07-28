package uredis

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/go-redis/redis"
)

const (
	Lock_Key_Pre = "redis_lock:"
	Lock_Expire  = 2000 //2000毫秒

	Err_Is_Exist  = "get lock fail"
	Lock_Interval = 5  //毫秒
	Lock_Repeated = 20 //重复次数
)

// 单点redis，多点注意
type RedisLock struct {
	lockKey   string
	lockValue string
	Key       string
	Field     string
	Expire    int64 //毫秒
	con       *redis.Client
}

// 保证原子性（redis是单线程），避免del删除了，其他client获得的lock
var delScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end`)

func NewRedisLock(key string, con *redis.Client, lockExpire ...int64) *RedisLock {
	lock := new(RedisLock)
	lock.lockKey = Lock_Key_Pre + key
	lock.Key = key
	lock.con = con

	b := make([]byte, 16)
	rand.Read(b)
	lock.lockValue = base64.StdEncoding.EncodeToString(b)

	lock.Expire = Lock_Expire
	if len(lockExpire) > 0 {
		if lockExpire[0] > 0 {
			lock.Expire = lockExpire[0]
		}
	}
	return lock
}

// 加锁 获取锁失败马上返回
func (lock *RedisLock) LockNoWait() error {
	lockReply, err := lock.con.SetNX(lock.lockKey, lock.lockValue, time.Duration(lock.Expire)*time.Millisecond).Result()
	if err != nil {
		return errors.New("redis fail")
	}
	if !lockReply {
		return errors.New(Err_Is_Exist)
	}
	return nil
}

// 等待锁 等待50ms
func (lock *RedisLock) Lock() error {
	var b = false
	for i := 0; i < Lock_Repeated; i++ {
		err := lock.LockNoWait()
		if err != nil {
			if err.Error() != Err_Is_Exist {
				return err
			}
			time.Sleep(Lock_Interval * time.Millisecond) //重试间隔
		} else {
			b = true
			break
		}
	}
	if !b {
		return errors.New(Err_Is_Exist)
	}
	return nil
}

// 解锁
func (lock *RedisLock) Unlock() error {
	return delScript.Run(lock.con, []string{lock.lockKey}, lock.lockValue).Err()
}
