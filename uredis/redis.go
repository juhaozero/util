package uredis

import (
	"errors"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

type RedisUtil struct {
	client           *redis.Client
	repeatedTimes    int
	repeatedInterval time.Duration
}

const (
	Repeated_Times    = 5   //次数
	Repeated_Interval = 200 //毫秒

	Default_Time = time.Duration(-1) * time.Second
)

func NewRedisClient(config *RedisUtil) *RedisUtil {
	return &RedisUtil{
		client:           config.client,
		repeatedTimes:    config.repeatedTimes,
		repeatedInterval: config.repeatedInterval,
	}
}
func (r *RedisUtil) TTL(key string) (time.Duration, error) {
	return r.client.TTL(key).Result()
}

// 设置过期时间
func (r *RedisUtil) Expire(key string, t int64) error {
	return r.client.Expire(key, time.Duration(t)*time.Second).Err()
}

// 删除key
func (r *RedisUtil) DelKey(key ...string) error {
	return r.client.Del(key...).Err()
}

// ### string 类型

// 获取string类型数据
func (r *RedisUtil) GetString(key string) (string, error) {
	count := 1
	for {
		result, err := r.client.Get(key).Result()
		if err == nil || err == redis.Nil {
			return result, nil
		}
		count++
		if count > r.repeatedTimes {
			return "", err
		} else {
			time.Sleep(r.repeatedInterval * time.Millisecond) //重试间隔
		}
	}
}

// 设置string类型数据
func (r *RedisUtil) SetString(key string, value interface{}, ex ...int64) error {
	count := 1
	var t = Default_Time
	if len(ex) > 0 {
		t = time.Duration(ex[0]) * time.Second
	}
	for {
		err := r.client.Set(key, value, t).Err()
		if err == nil {
			return nil
		}
		count++
		if count > r.repeatedTimes {
			return err
		} else {
			time.Sleep(r.repeatedInterval * time.Millisecond) //重试间隔
		}
	}
}

// ### zset 类型

// 添加有序集合
func (r *RedisUtil) ZSortSet(key string, value ...redis.Z) error {
	if err := r.client.ZAdd(key, value...).Err(); err != nil {

		return err
	}
	return nil
}

// 设置对应分数
func (r *RedisUtil) ZScore(key string, value string) (float64, error) {
	return r.client.ZScore(key, value).Result()

}

// 数量
func (r *RedisUtil) ZCard(key string, value string) (int64, error) {
	count, err := r.client.ZCard(key).Result()
	if err != nil {

		return count, err
	}
	return count, nil
}

// 获取升序集合数据 (从小到大, num[0]=start num[1]=end) index 区间
func (r *RedisUtil) GetZRageSort(key string, num ...int64) ([]string, error) {
	var n1, n2 int64 = 0, -1
	if len(num) > 0 {
		n1 = num[0]
	}
	if len(num) > 1 {
		n2 = num[1]
	}
	arr, err := r.client.ZRange(key, n1, n2).Result()
	if err != nil {
		return nil, err
	}
	return arr, nil
}

// 获取降序集合数据  (从大到小,num[0]=start num[1]=end,默认全部) index 区间
func (r *RedisUtil) GetZRevRangeSort(key string, num ...int64) ([]string, error) {
	var n1, n2 int64 = 0, -1
	if len(num) > 0 {
		n1 = num[0]
	}
	if len(num) > 1 {
		n2 = num[1]
	}
	arr, err := r.client.ZRevRange(key, n1, n2).Result()
	if err != nil {

		return nil, err
	}
	return arr, nil
}

// 获取score过滤后的升序集合数据  (从小到大)/min,max 分数区间
func (r *RedisUtil) GetZRevRangeByScoreSort(key string, fields ...string) ([]string, error) {

	if len(fields) == 0 {
		return nil, errors.New("参数错误")
	}

	arr, err := r.client.ZRangeByScore(key, redis.ZRangeBy{
		Min: fields[0],
		Max: fields[1],
	}).Result()
	if err != nil {
		return nil, err
	}
	return arr, nil
}

// 获取score过滤后的降序集合数据  (从小到大,)/min,max 分数区间
func (r *RedisUtil) GetZRageRangeByScoreSort(key string, fields ...string) ([]string, error) {

	if len(fields) == 0 {
		return nil, errors.New("参数错误")
	}
	arr, err := r.client.ZRevRangeByScore(key, redis.ZRangeBy{
		Min: fields[0],
		Max: fields[1],
	}).Result()
	if err != nil {
		return nil, err
	}
	return arr, nil
}

// 获取区间内用户信息  (从小到大,)/min,max 分数区间
func (r *RedisUtil) GetZRevRangeUserWithScores(key string, num ...int64) ([]redis.Z, error) {
	var n1, n2 int64 = 0, -1
	if len(num) > 0 {
		n1 = num[0]
	}
	if len(num) > 1 {
		n2 = num[1]
	}

	arr, err := r.client.ZRevRangeWithScores(key, n1, n2).Result()
	if err != nil {
		return nil, err
	}
	return arr, nil
}

// 删除指定区间数据
func (r *RedisUtil) DelZSortData(key string, num ...int64) error {
	if len(num) == 0 {
		return errors.New("参数错误")
	}
	s := strconv.FormatInt(num[0], 10)
	e := strconv.FormatInt(num[1], 10)
	err := r.client.ZRemRangeByLex(key, s, e).Err()
	return err
}

// ### list 类型

// 某个元素是否存在
func (r *RedisUtil) HashExists(key string, field string) (bool, error) {
	return r.client.HExists(key, field).Result()
}

// 压入链表
func (r *RedisUtil) LPush(key string, value interface{}) (int64, error) {
	return r.client.LPush(key, value).Result()
}

// 截断链表
func (r *RedisUtil) LTrim(key string, num ...int64) error {
	return r.client.LTrim(key, num[0], num[1]).Err()
}

// 弹出链表
func (r *RedisUtil) LPop(key string) (string, error) {
	return r.client.LPop(key).Result()
}

// 设置某个下标元素的值
func (r *RedisUtil) LSet(key string, index int64, value interface{}) (string, error) {
	return r.client.LSet(key, index, value).Result()
}

// 获取指定范围的链表
func (r *RedisUtil) LRange(key string, num ...int64) ([]string, error) {
	var n1, n2 int64 = 0, -1
	if len(num) > 0 {
		n1 = num[0]
	}
	if len(num) > 1 {
		n2 = num[1]
	}

	arr, err := r.client.LRange(key, n1, n2).Result()
	if err != nil {
		return nil, err
	}
	return arr, nil
}

// 根据值移除
func (r *RedisUtil) LRem(key string, count int64, value interface{}) error {
	return r.client.LRem(key, count, value).Err()
}

// 链表数量
func (r *RedisUtil) LLen(key string) (int64, error) {
	return r.client.LLen(key).Result()
}

// ### set 类型

// 添加元素
func (r *RedisUtil) SAdd(key string, value interface{}, ex ...int64) error {
	err := r.client.SAdd(key, value).Err()
	if err == nil && len(ex) > 0 {
		var t = Default_Time
		t = time.Duration(ex[0]) * time.Second
		err = r.client.Expire(key, t).Err()
		return err
	}
	return err
}

// 获取所有元素
func (r *RedisUtil) SMembers(key string) ([]string, error) {
	return r.client.SMembers(key).Result()

}

// 元素是否存在集合中
func (r *RedisUtil) SIsMember(key string, field string) (bool, error) {
	arr, err := r.client.SIsMember(key, field).Result()
	return arr, err
}

// 删除元素
func (r *RedisUtil) SRem(key string, field ...string) error {
	return r.client.SRem(key, field).Err()
}

// 随机获取一个元素
func (r *RedisUtil) SPop(key string, field ...string) (string, error) {
	return r.client.SPop(key).Result()
}

// 数量
func (r *RedisUtil) SCard(key string, field ...string) (int64, error) {
	return r.client.SCard(key).Result()
}

// ### hash 类型

// 获取Hash类型数据
func (r *RedisUtil) GetHash(key string, field string) (string, error) {
	count := 1
	for {
		result, err := r.client.HGet(key, field).Result()
		if err == nil {
			return result, nil
		}
		if err == redis.Nil {
			return "", nil
		}
		count++
		if count > r.repeatedTimes {
			return "", err
		} else {
			time.Sleep(r.repeatedInterval * time.Millisecond) //重试间隔
		}
	}
}

// 获取Hash全部数据数据
func (r *RedisUtil) GetHashAll(key string) (map[string]string, error) {
	count := 1
	for {
		result, err := r.client.HGetAll(key).Result()
		if err == nil {
			return result, nil
		}
		if err == redis.Nil {
			return nil, nil
		}
		count++
		if count > Repeated_Times {

			return nil, err
		} else {
			time.Sleep(Repeated_Interval * time.Millisecond) //重试间隔
		}
	}
}

// 获取hash的长度
func (r *RedisUtil) GetHashLen(key string) (int64, error) {
	result, err := r.client.HLen(key).Result()
	return result, err
}

// SetHash 设置Hash集合
func (r *RedisUtil) SetHash(key, field string, value []byte) error {
	count := 1
	for {
		err := r.client.HSet(key, field, value).Err()
		if err == nil {
			return nil
		}
		count++
		if count > r.repeatedTimes {

			return err
		} else {
			time.Sleep(r.repeatedInterval * time.Millisecond) //重试间隔
		}
	}
}

// 删除hash
func (r *RedisUtil) DelHash(key string, field ...string) error {
	err := r.client.HDel(key, field...).Err()
	return err
}

func (r *RedisUtil) SetNx(key string, value interface{}, ex ...int64) (bool, error) {
	if len(ex) > 0 {
		t := time.Duration(ex[0]) * time.Second
		return r.client.SetNX(key, value, t).Result()
	} else {
		return r.client.SetNX(key, value, 0).Result()
	}
}

func (r *RedisUtil) Keys(key string) ([]string, error) {
	return r.client.Keys(key).Result()
}

func (r *RedisUtil) IsExists(key string) (bool, error) {
	ENum, err := r.client.Exists(key).Result()
	if err == nil {
		if ENum > 0 {
			return true, err
		}
	}
	return false, err
}

func (r *RedisUtil) IncrKey(key string) (int64, error) {
	i, err := r.client.Incr(key).Result()
	if err != nil {
		return 0, err
	}
	return i, err
}
func (r *RedisUtil) DecrKey(key string) (int64, error) {
	i, err := r.client.Decr(key).Result()
	if err != nil {
		return -1, err
	}
	return i, err
}

// 自增
func (r *RedisUtil) Incr(key string) error {

	return r.client.Incr(key).Err()
}

func (r *RedisUtil) IncrBy(key string, val int64) (int64, error) {
	return r.client.IncrBy(key, val).Result()
}

// 自减
func (r *RedisUtil) Decr(key string) error {
	return r.client.Decr(key).Err()

}

func (r *RedisUtil) DecrBy(key string, val int64) error {
	return r.client.DecrBy(key, val).Err()
}
