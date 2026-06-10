package common

import (
	"math/rand"

	"github.com/juhaozero/util/model"
)

func RandomInt[T model.Integer](min, max T) T {
	return min + T(rand.Intn(int(max-min)))
}

// RandomString 随机字符串
func RandomString[T model.Integer](n T) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[RandomInt(0, len(letters))]
	}
	return string(b)
}
