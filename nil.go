//
// FilePath    : go-utils\nil.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : nil 判断
//

package utils

import "reflect"

// IsInterfaceNil 判断接口变量是否为 nil
//
// 对于接口变量 i，reflect.ValueOf(i) 返回的是 i 持有的具体值，而不是 i 本身。
// 可能的情况有两种：
// 1. 接口变量本身为 nil：这意味着接口变量的类型信息和数据指针都是 nil。
// 2. 接口变量的具体类型为 nil：这意味着接口变量的类型信息不为 nil，但数据指针为 nil。
func IsInterfaceNil(i any) bool {
	// 当 i 为 nil 时，直接返回 true。
	if i == nil {
		return true
	}

	// 否则，使用 reflect.ValueOf(i) 获取 i 持有的具体值，并检查这个值是否是一个 nil 指针。
	v := reflect.ValueOf(i)

	return v.Kind() == reflect.Pointer && v.IsNil()
}
