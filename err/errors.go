package err

import (
	"errors"

	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

var codes = make(map[error]int32)

var (
	// 请求成功
	Success = reg(200, "")
)

func reg(code int32, msg string) *Error {
	err := newError(code, msg)
	codes[err] = code
	return err
}

type Error struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

func (e *Error) Error() string {

	return e.Msg
}

func newError(code int32, detail string) *Error {
	return &Error{
		Code: code,
		Msg:  detail,
	}
}

// mysql 未找到
func IsSqlNoFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// redis 为空
func IsRedisNil(err error) bool {
	return errors.Is(err, redis.Nil)
}
