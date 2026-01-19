//
// FilePath    : go-utils\types\json_int64.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 字符串转int64
//

package types

import (
	"encoding/json"
	"strconv"
)

// JSONInt64 自定义类型，以 int64 形式解析字符串。
type JSONInt64 int64

// UnmarshalJSON 实现 JSONInt64 的 json.Unmarshaler 接口。
func (u *JSONInt64) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	if s == "" {
		*u = 0
		return nil
	}

	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}

	*u = JSONInt64(v)

	return nil
}

// MarshalJSON 实现 JSONInt64 的 json.Marshaler 接口。
func (u *JSONInt64) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(*u), 10))
}

// 将 JSONInt64 转换为 string
func (u *JSONInt64) ToString() string {
	return strconv.FormatInt(int64(*u), 10)
}

// JSONInt64Slice 自定义类型，以 int64 切片形式解析字符串切片。
type JSONInt64Slice []int64

// UnmarshalJSON 实现 JSONInt64Slice 的 json.UnmarshalJSON 接口。
func (u *JSONInt64Slice) UnmarshalJSON(b []byte) error {
	var strSlice []string
	if err := json.Unmarshal(b, &strSlice); err != nil {
		return err
	}

	int64Slice := make([]int64, len(strSlice))

	for i, s := range strSlice {
		if s == "" {
			int64Slice[i] = 0
			continue
		}

		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}

		int64Slice[i] = v
	}

	*u = int64Slice

	return nil
}

// MarshalJSON 实现 JSONInt64Slice 的 json.Marshaler 接口。
func (u *JSONInt64Slice) MarshalJSON() ([]byte, error) {
	strSlice := make([]string, len(*u))
	for i, v := range *u {
		strSlice[i] = strconv.FormatInt(v, 10)
	}

	return json.Marshal(strSlice)
}
