package common

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	json "github.com/json-iterator/go"
	"github.com/shopspring/decimal"
)

func GetTimeMicro() int64 {
	return time.Now().UnixMicro()
}
func GetTimeNano() int64 {
	return time.Now().UnixNano()
}
func GetTimeMs() int64 {
	return time.Now().UnixMilli()
}
func GetTime() int64 {
	return time.Now().Unix()
}
func GetTimeFormat() string {
	return time.Now().Format("2006-01-02")
}
func GetNumIsEven[T int | int32 | int64](data T) bool {
	return int64(data)&1 == 0
}
func ChangeDoubleToString[T float32 | float64](data T) string {
	return strconv.FormatFloat(float64(data), 'f', 2, 64)
}
func ChangeIntToString[T int | int32 | int64](data T) string {
	return strconv.FormatInt(int64(data), 10)
}
func ChangeInt64(data string) int64 {
	s, _ := strconv.ParseInt(data, 10, 64)
	return s
}
func Marshal(data interface{}) (s string) {
	s, _ = json.MarshalToString(data)
	return s
}
func Unmarshal(s string, data interface{}) error {
	return json.UnmarshalFromString(s, data)
}
func GetOnlyId() string {
	id := uuid.New().String()
	return strings.ReplaceAll(id, "-", "")
}
func GetKey(key string, field ...string) (s string) {
	data := strings.Builder{}
	data.WriteString(key)
	for _, v := range field {
		data.WriteString("_")
		data.WriteString(v)
	}

	return data.String()
}

// day 偏移的天数
func GetTimesSame[T int | int32 | int64](times T, day int) bool {
	now := time.Now().AddDate(0, 0, day).Format("2006-01-02")
	sign := time.Unix(int64(times), 0).Format("2006-01-02")
	return now == sign
}
func GetExpTime(day int) int64 {
	timeStr := time.Now().Format("2006-01-02")
	t2, _ := time.ParseInLocation("2006-01-02", timeStr, time.Local)
	return t2.AddDate(0, 0, day).Unix() - time.Now().Unix()
}

// GetRandom 随机数区间
func GetRandom[T int | int32 | int64](min, max T) T {
	base := int32(min) + rand.Int31n(int32(max-min+1))
	return T(base)
}

const (
	Add = iota
	Sub
	Multiply
	Divide
)

func Decimal[T float32 | float64](data1, data2 T, types int) float64 {
	var value float64
	switch types {
	case Multiply:
		value, _ = decimal.NewFromFloat(float64(data1)).Mul(decimal.NewFromFloat(float64(data2))).Float64()
	case Add:
		value, _ = decimal.NewFromFloat(float64(data1)).Add(decimal.NewFromFloat(float64(data2))).Float64()
	case Sub:
		value, _ = decimal.NewFromFloat(float64(data1)).Sub(decimal.NewFromFloat(float64(data2))).Float64()
	case Divide:
		value, _ = decimal.NewFromFloat(float64(data1)).Div(decimal.NewFromFloat(float64(data2))).Float64()

	}

	return DecimalData(value, 4)
}
func DecimalData(value float64, num int32) float64 {
	d, _ := decimal.NewFromFloat(value).RoundFloor(num).Float64()

	return d
}

// BinarySearch 两分法
func BinarySearch(data []int64, target int64) int {
	left, right := 0, len(data)
	for left <= right {
		mid := left + (right-left)/2
		if data[mid] > target {
			right = mid - 1
		} else if data[mid] < target {
			left = mid + 1
		} else {
			if mid == 0 || data[mid-1] != target {
				return mid
			} else {
				right = mid - 1
			}
		}
	}

	return left
}