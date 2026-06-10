package common

import (
	json "github.com/json-iterator/go"
)

// Marshal 结构体转json格式字符串
func Marshal(data any) (s string) {
	s, _ = json.MarshalToString(data)
	return s
}

// Unmarshal json格式字符串转结构体
func Unmarshal(s string, data any) error {
	return json.UnmarshalFromString(s, data)
}
