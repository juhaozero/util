package uredis

import (
	"errors"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

const (
	Repeated_Times    = 5   //次数
	Repeated_Interval = 200 //毫秒

	Default_Time = time.Duration(-1) * time.Second
)

func TTL(con *redis.Client, key string) (time.Duration, error) {
	return con.TTL(key).Result()
}

// 设置过期时间
func Expire(con *redis.Client, key string, t int64) error {
	return con.Expire(key, time.Duration(t)*time.Second).Err()
}

// 删除key
func DelKey(con *redis.Client, key ...string) error {
	return con.Del(key...).Err()
}

// ### string 类型

// 获取string类型数据
func GetString(con *redis.Client, key string) (string, error) {
	count := 1
	for {
		result, err := con.Get(key).Result()
		if err == nil {
			return result, nil
		}
		if err == redis.Nil {
			return "", nil
		}
		count++
		if count > Repeated_Times {

			return "", err
		} else {
			time.Sleep(Repeated_Interval * time.Millisecond) //重试间隔
		}
	}
}

// 设置string类型数据
func SetString(con *redis.Client, key string, value interface{}, ex ...int64) error {
	count := 1
	var t = Default_Time
	if len(ex) > 0 {
		t = time.Duration(ex[0]) * time.Second
	}
	for {
		err := con.Set(key, value, t).Err()
		if err == nil {
			return nil
		}
		count++
		if count > Repeated_Times {
			return err
		} else {
			time.Sleep(Repeated_Interval * time.Millisecond) //重试间隔
		}
	}
}

// ### zset 类型

// 添加有序集合
func ZSortSet(con *redis.Client, key string, value ...redis.Z) error {
	if err := con.ZAdd(key, value...).Err(); err != nil {

		return err
	}
	return nil
}

// 设置对应分数
func ZScore(con *redis.Client, key string, value string) (float64, error) {
	return con.ZScore(key, value).Result()

}

// 数量
func ZCard(con *redis.Client, key string, value string) (int64, error) {
	count, err := con.ZCard(key).Result()
	if err != nil {

		return count, err
	}
	return count, nil
}

// 获取升序集合数据 (从小到大, num[0]=start num[1]=end) index 区间
func GetZRageSort(con *redis.Client, key string, num ...int64) ([]string, error) {
	var n1, n2 int64 = 0, -1
	if len(num) > 0 {
		n1 = num[0]
	}
	if len(num) > 1 {
		n2 = num[1]
	}
	arr, err := con.ZRange(key, n1, n2).Result()
	if err != nil {
		return nil, err
	}
	return arr, nil
}

// 获取降序集合数据  (从大到小,num[0]=start num[1]=end,默认全部) index 区间
func GetZRevRangeSort(con *redis.Client, key string, num ...int64) ([]string, error) {
	var n1, n2 int64 = 0, -1
	if len(num) > 0 {
		n1 = num[0]
	}
	if len(num) > 1 {
		n2 = num[1]
	}
	arr, err := con.ZRevRange(key, n1, n2).Result()
	if err != nil {

		return nil, err
	}
	return arr, nil
}

// 获取score过滤后的升序集合数据  (从小到大)/min,max 分数区间
func GetZRevRangeByScoreSort(con *redis.Client, key string, fields ...string) ([]string, error) {

	if len(fields) == 0 {
		return nil, errors.New("参数错误")
	}

	arr, err := con.ZRangeByScore(key, redis.ZRangeBy{
		Min: fields[0],
		Max: fields[1],
	}).Result()
	if err != nil {
		return nil, err
	}
	return arr, nil
}

// 获取score过滤后的降序集合数据  (从小到大,)/min,max 分数区间
func GetZRageRangeByScoreSort(con *redis.Client, key string, fields ...string) ([]string, error) {

	if len(fields) == 0 {
		return nil, errors.New("参数错误")
	}
	arr, err := con.ZRevRangeByScore(key, redis.ZRangeBy{
		Min: fields[0],
		Max: fields[1],
	}).Result()
	if err != nil {
		return nil, err
	}
	return arr, nil
}

// 获取区间内用户信息  (从小到大,)/min,max 分数区间
func GetZRevRangeUserWithScores(con *redis.Client, key string, num ...int64) ([]redis.Z, error) {
	var n1, n2 int64 = 0, -1
	if len(num) > 0 {
		n1 = num[0]
	}
	if len(num) > 1 {
		n2 = num[1]
	}

	arr, err := con.ZRevRangeWithScores(key, n1, n2).Result()
	if err != nil {
		return nil, err
	}
	return arr, nil
}

// 删除指定区间数据
func DelZSortData(con *redis.Client, key string, num ...int64) error {
	if len(num) == 0 {
		return errors.New("参数错误")
	}
	s := strconv.FormatInt(num[0], 10)
	e := strconv.FormatInt(num[1], 10)
	err := con.ZRemRangeByLex(key, s, e).Err()
	return err
}

// ### list 类型

// 某个元素是否存在
func HashExists(con *redis.Client, key string, field string) (bool, error) {
	return con.HExists(key, field).Result()
}

// 压入链表
func LPush(con *redis.Client, key string, value interface{}) (int64, error) {
	return con.LPush(key, value).Result()
}

// 截断链表
func LTrim(con *redis.Client, key string, num ...int64) error {
	return con.LTrim(key, num[0], num[1]).Err()
}

// 弹出链表
func LPop(con *redis.Client, key string) (string, error) {
	return con.LPop(key).Result()
}

// 设置某个下标元素的值
func LSet(con *redis.Client, key string, index int64, value interface{}) (string, error) {
	return con.LSet(key, index, value).Result()
}

// 获取指定范围的链表
func LRange(con *redis.Client, key string, num ...int64) ([]string, error) {
	var n1, n2 int64 = 0, -1
	if len(num) > 0 {
		n1 = num[0]
	}
	if len(num) > 1 {
		n2 = num[1]
	}

	arr, err := con.LRange(key, n1, n2).Result()
	if err != nil {
		return nil, err
	}
	return arr, nil
}

// 根据值移除
func LRem(con *redis.Client, key string, count int64, value interface{}) error {
	return con.LRem(key, count, value).Err()
}

// 链表数量
func LLen(con *redis.Client, key string) (int64, error) {
	return con.LLen(key).Result()
}

// ### set 类型

// 添加元素
func SAdd(con *redis.Client, key string, value interface{}, ex ...int64) error {
	err := con.SAdd(key, value).Err()
	if err == nil && len(ex) > 0 {
		var t = Default_Time
		t = time.Duration(ex[0]) * time.Second
		err = con.Expire(key, t).Err()
		return err
	}
	return err
}

// 获取所有元素
func SMembers(con *redis.Client, key string) ([]string, error) {
	return con.SMembers(key).Result()

}

// 元素是否存在集合中
func SIsMember(con *redis.Client, key string, field string) (bool, error) {
	arr, err := con.SIsMember(key, field).Result()
	return arr, err
}

// 删除元素
func SRem(con *redis.Client, key string, field ...string) error {
	return con.SRem(key, field).Err()
}

// 随机获取一个元素
func SPop(con *redis.Client, key string, field ...string) (string, error) {
	return con.SPop(key).Result()
}

// 数量
func SCard(con *redis.Client, key string, field ...string) (int64, error) {
	return con.SCard(key).Result()
}

// ### hash 类型

// 获取Hash类型数据
func GetHash(con *redis.Client, key string, field string) (string, error) {
	count := 1
	for {
		result, err := con.HGet(key, field).Result()
		if err == nil {
			return result, nil
		}
		if err == redis.Nil {
			return "", nil
		}
		count++
		if count > Repeated_Times {
			return "", err
		} else {
			time.Sleep(Repeated_Interval * time.Millisecond) //重试间隔
		}
	}
}

// 获取Hash全部数据数据
func GetHashAll(con *redis.Client, key string) (map[string]string, error) {
	count := 1
	for {
		result, err := con.HGetAll(key).Result()
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
func GetHashLen(con *redis.Client, key string) (int64, error) {
	result, err := con.HLen(key).Result()
	return result, err
}

// SetHash 设置Hash集合
func SetHash(con *redis.Client, key, field string, value []byte) error {
	count := 1
	for {
		err := con.HSet(key, field, value).Err()
		if err == nil {
			return nil
		}
		count++
		if count > Repeated_Times {

			return err
		} else {
			time.Sleep(Repeated_Interval * time.Millisecond) //重试间隔
		}
	}
}

// 删除hash
func DelHash(con *redis.Client, key string, field ...string) error {
	err := con.HDel(key, field...).Err()
	return err
}

func SetNx(con *redis.Client, key string, value interface{}, ex ...int64) (bool, error) {
	if len(ex) > 0 {
		t := time.Duration(ex[0]) * time.Second
		return con.SetNX(key, value, t).Result()
	} else {
		return con.SetNX(key, value, 0).Result()
	}
}

func Keys(con *redis.Client, key string) ([]string, error) {
	return con.Keys(key).Result()
}

func IsExists(con *redis.Client, key string) (bool, error) {
	ENum, err := con.Exists(key).Result()
	if err == nil {
		if ENum > 0 {
			return true, err
		}
	}
	return false, err
}

func IncrKey(con *redis.Client, key string) (int64, error) {
	i, err := con.Incr(key).Result()
	if err != nil {
		return 0, err
	}
	return i, err
}
func DecrKey(con *redis.Client, key string) (int64, error) {
	i, err := con.Decr(key).Result()
	if err != nil {
		return -1, err
	}
	return i, err
}

func Incr(con *redis.Client, key string) error {

	return con.Incr(key).Err()
}

func IncrBy(con *redis.Client, key string, val int64) (int64, error) {
	return con.IncrBy(key, val).Result()
}

func Decr(con *redis.Client, key string) error {
	return con.Decr(key).Err()

}
