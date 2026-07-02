package common

import (
	"math/rand"
	"time"

	"github.com/juhaozero/util/model"
)

var (
	Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// RandomString 随机字符串
func RandomString[T model.Integer](n T) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[RandomNum(0, len(letters))]
	}
	return string(b)
}
func RandomNum[T model.Number](min, max T) T {
	return min + T(Rand.Intn(int(max-min)))
}

// RandomFloat 随机浮点数
// min 最小值
// max 最大值
func RandomFloat[T model.Float](min, max T) T {
	return min + T(Rand.Float64()*(float64(max)-float64(min)))
}

