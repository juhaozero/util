package common

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/juhaozero/util/model"

	"github.com/google/uuid"
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

// StringToFloat string类型转 float64类型
func StringToFloat[T model.Float](data string) T {
	s, _ := strconv.ParseFloat(data, 64)
	return T(s)
}

// GetOnlyId 获取唯一id
func GetOnlyId[T model.Number](num T) string {
	id := uuid.New().String()
	id = strings.ReplaceAll(id, "-", "")
	if int(num) > len(id) {
		num = T(len(id))
	}
	return id[:num]
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
	return strings.Trim(strings.ReplaceAll(fmt.Sprint(array), " ", ","), "[]")
}

// 复制指针
func CopyPoint(m any) any {
	vt := reflect.TypeOf(m).Elem()
	newoby := reflect.New(vt)
	newoby.Elem().Set(reflect.ValueOf(m).Elem())
	return newoby.Interface()
}

// StructToMapString 结构体转map
func StructToMapString(obj any) (map[string]string, error) {
	if obj == nil {
		return nil, fmt.Errorf("obj must not be nil")
	}

	valueOf := reflect.ValueOf(obj)
	if valueOf.Kind() == reflect.Pointer {
		if valueOf.IsNil() {
			return nil, fmt.Errorf("obj must not be nil")
		}
		valueOf = valueOf.Elem()
	}
	if valueOf.Kind() != reflect.Struct {
		return nil, fmt.Errorf("obj must be a struct or pointer to struct")
	}

	mapping := make(map[string]string)
	typ := valueOf.Type()
	for i := 0; i < valueOf.NumField(); i++ {
		structField := typ.Field(i)
		field := valueOf.Field(i)
		if !structField.IsExported() || !field.CanInterface() {
			continue
		}

		jTag, ok := jsonFieldName(structField.Tag.Get("json"))
		if !ok {
			continue
		}
		if field.IsZero() {
			continue
		}
		mapping[jTag] = fmt.Sprint(field.Interface())
	}
	return mapping, nil
}

// MapToStruct map转结构体
func MapToStruct(data map[string]string, obj any) error {
	if obj == nil {
		return fmt.Errorf("obj must not be nil")
	}

	valueOf := reflect.ValueOf(obj)
	if valueOf.Kind() != reflect.Pointer {
		return fmt.Errorf("obj must be a pointer to struct")
	}
	if valueOf.IsNil() {
		return fmt.Errorf("obj must not be nil")
	}

	valueOf = valueOf.Elem()
	if valueOf.Kind() != reflect.Struct {
		return fmt.Errorf("obj must be a pointer to struct")
	}

	typ := valueOf.Type()
	for i := 0; i < valueOf.NumField(); i++ {
		structField := typ.Field(i)
		field := valueOf.Field(i)
		if !field.IsValid() || !field.CanSet() {
			continue
		}

		jTag, ok := jsonFieldName(structField.Tag.Get("json"))
		if !ok {
			continue
		}

		val, ok := data[jTag]
		if !ok {
			continue
		}

		if err := setFieldFromString(field, val); err != nil {
			return fmt.Errorf("set field %s: %w", structField.Name, err)
		}
	}
	return nil
}

func setFieldFromString(field reflect.Value, val string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(val)
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		field.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(val, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(val, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(val, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetFloat(n)
	default:
		return fmt.Errorf("unsupported type %s", field.Type())
	}
	return nil
}

func jsonFieldName(tag string) (string, bool) {
	index := strings.Index(tag, ",")
	if index > 0 {
		tag = tag[:index]
	}
	if tag == "" || tag == "-" {
		return "", false
	}
	return tag, true
}
