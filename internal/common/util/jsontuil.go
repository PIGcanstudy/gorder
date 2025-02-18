package util

import "encoding/json"

// 用来序列化成json字符串
func MarshalString(v any) (string, error) {
	bytes, err := json.Marshal(v)
	return string(bytes), err
}
