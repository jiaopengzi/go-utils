//
// FilePath    : go-utils\types\json_uint64.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 字符串转uint64
//

package types

import (
	"encoding/json"
	"strconv"
)

// JSONUint64 自定义类型，以 uint64 形式解析字符串。
type JSONUint64 uint64

// UnmarshalJSON 实现 JSONUint64 的 json.Unmarshaler 接口。
func (u *JSONUint64) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	if s == "" {
		*u = 0
		return nil
	}

	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}

	*u = JSONUint64(v)

	return nil
}

// MarshalJSON 实现 JSONUint64 的 json.Marshaler 接口。
func (u JSONUint64) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatUint(uint64(u), 10))
}

// 将 JSONUint64 转换为 string
func (u JSONUint64) ToString() string {
	return strconv.FormatUint(uint64(u), 10)
}

// JSONUint64Slice 自定义类型，以 uint64 切片形式解析字符串切片。
type JSONUint64Slice []uint64

// UnmarshalJSON 实现 JSONUint64Slice 的 json.UnmarshalJSON 接口。
func (u *JSONUint64Slice) UnmarshalJSON(b []byte) error {
	var strSlice []string
	if err := json.Unmarshal(b, &strSlice); err != nil {
		return err
	}

	uint64Slice := make([]uint64, len(strSlice))

	for i, s := range strSlice {
		if s == "" {
			uint64Slice[i] = 0
			continue
		}

		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}

		uint64Slice[i] = v
	}

	*u = uint64Slice

	return nil
}

// MarshalJSON 实现 JSONUint64Slice 的 json.Marshaler 接口。
func (u JSONUint64Slice) MarshalJSON() ([]byte, error) {
	strSlice := make([]string, len(u))
	for i, v := range u {
		strSlice[i] = strconv.FormatUint(v, 10)
	}

	return json.Marshal(strSlice)
}

// ConvertToUint64Slice 将 JSONUint64Slice 转换为 any 切片。
func ConvertToAnySlice(slice JSONUint64Slice) []any {
	// 创建一个 any 切片
	uint64Any := make([]any, len(slice))
	// 遍历 JSONUint64Slice，将每个元素转换为 any 类型
	for i, v := range slice {
		uint64Any[i] = v
	}

	return uint64Any
}
