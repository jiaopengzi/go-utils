//
// FilePath    : go-utils\type.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 类型相关工具
//

package utils

import "reflect"

// IsPointer 检查传入的 v 是否是指针
func IsPointer(v any) bool {
	return reflect.TypeOf(v).Kind() == reflect.Pointer
}
