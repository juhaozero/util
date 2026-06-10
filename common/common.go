package common

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"

	"github.com/juhaozero/util/model"

	"github.com/google/uuid"
	json "github.com/json-iterator/go"
	"github.com/shopspring/decimal"
)

// DoubleToString 小数类型转 string类型
func DoubleToString[T model.Float](data T) string {
	return strconv.FormatFloat(float64(data), 'f', 2, 64)
}

// NumberToString 数字类型转 string类型
func NumberToString[T model.Number](data T) string {
	return strconv.FormatInt(int64(data), 10)
}

// StringToNumber string类型转 数字类型
func StringToNumber[T model.Number](data string) T {
	s, _ := strconv.ParseInt(data, 10, 64)
	return T(s)
}

// Marshal 结构体转json格式字符串
func Marshal(data any) (s string) {
	s, _ = json.MarshalToString(data)
	return s
}

// Unmarshal json格式字符串转结构体
func Unmarshal(s string, data any) error {
	return json.UnmarshalFromString(s, data)
}

// GetOnlyId 获取唯一id
func GetOnlyId() string {
	id := uuid.New().String()
	return strings.ReplaceAll(id, "-", "")
}

// GetKey 拼接key
func GetKey(key string, separator string, field ...string) (s string) {
	data := strings.Builder{}
	data.WriteString(key)
	for _, v := range field {
		data.WriteString(separator)
		data.WriteString(v)
	}
	return data.String()
}

// GetRandom 随机数区间
func GetRandom[T model.Number](min, max T) T {
	base := int64(min) + rand.Int63n(int64(max-min+1))
	return T(base)
}

// Decimal 精确浮点加减
// num 保留小数点后几位
func Decimal[T model.Float, A model.Number](data1, data2 T, types, num A) T {
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
	return T(DecimalData(value, int32(num)))
}

// DecimalData 保留小数点后几位
func DecimalData(value float64, num int32) float64 {
	d, _ := decimal.NewFromFloat(value).RoundFloor(num).Float64()
	return d
}

// BinarySearch 两分法
func BinarySearch[T model.Number](data []T, target T) int {
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

// 数组转 string字符串
func ArrayToString(array []any) string {
	return strings.Replace(strings.Trim(fmt.Sprint(array), "[]"), " ", ",", -1)
}

// 复制指针
func CopyPoint(m any) any {
	vt := reflect.TypeOf(m).Elem()
	newoby := reflect.New(vt)
	newoby.Elem().Set(reflect.ValueOf(m).Elem())
	return newoby.Interface()
}

// StructToMapString 结构体转map
func StructToMapString(obj any) map[string]string {
	mapping := make(map[string]string)
	var valueOf = reflect.ValueOf(obj)
	if valueOf.Kind() == reflect.Pointer {
		valueOf = reflect.ValueOf(obj).Elem()
	}
	for i := 0; i < valueOf.NumField(); i++ {
		field := valueOf.Field(i)
		jTag := valueOf.Type().Field(i).Tag.Get("json")
		index := strings.Index(jTag, ",")
		if index > 0 {
			jTag = jTag[:index]
		}
		if field.IsZero() {
			continue
		}
		mapping[jTag] = fmt.Sprint(field.Interface())
	}
	return mapping
}
func RandomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

// RandomString 随机字符串
func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[RandomInt(0, len(letters))]
	}
	return string(b)
}
