//
// FilePath    : go-utils\logger\utils.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 日志工具
//

package logger

import (
	"reflect"
	"strings"
)

// SensitiveFields 全局变量敏感字段关键字切片
var SensitiveFields = []string{"password", "token", "secret"}

// MaskSensitiveFields 将传入 data 包含敏感字段关键字(包含即可,大小写不敏感)的字段值替换为 "******"
func MaskSensitiveFields(data any, sensitiveFields []string) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	recursiveMaskSensitiveFields(v, sensitiveFields)
}

// recursiveMaskSensitiveFields 递归处理敏感字段加上掩码
func recursiveMaskSensitiveFields(v reflect.Value, sensitiveFields []string) {
	// 如果是指针类型, 获取其指向的值
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	// 如果值不可设置, 直接返回
	if !v.CanSet() {
		return
	}

	// 分发不同类型的处理逻辑
	switch v.Kind() {
	case reflect.Struct:
		handleStructFields(v, sensitiveFields)
	case reflect.Map:
		handleMapValues(v, sensitiveFields)
	case reflect.Slice:
		handleSliceElements(v, sensitiveFields)
	}
}

// isFieldSensitive 判断字段名是否包含任意敏感关键字(不区分大小写)
func isFieldSensitive(lowerFieldName string, sensitiveFields []string) bool {
	for _, sensitiveField := range sensitiveFields {
		if strings.Contains(lowerFieldName, strings.ToLower(sensitiveField)) {
			return true
		}
	}

	return false
}

// maskFieldValue 对单个字段执行掩码操作, 支持 string 和 *string 两种情况; 其他类型触发 panic(保留原行为)
func maskFieldValue(field reflect.Value) {
	switch field.Kind() {
	case reflect.String:
		field.SetString("******")
	case reflect.Pointer:
		if field.IsNil() || field.Elem().Kind() != reflect.String {
			return
		}

		elem := field.Elem()
		if elem.CanSet() {
			elem.SetString("******")
		}
	default:
		panic("unhandled default case")
	}
}

// handleStructFields 处理结构体的每个字段：判断敏感字段并掩码, 遇到嵌套结构体时递归调用
func handleStructFields(v reflect.Value, sensitiveFields []string) {
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)

		// 将字段名转换为小写以进行大小写不敏感的匹配
		lowerFieldName := strings.ToLower(fieldType.Name)

		// 检查字段名是否包含任意敏感字段(不区分大小写)
		if isFieldSensitive(lowerFieldName, sensitiveFields) && field.CanSet() {
			maskFieldValue(field)
		}

		// 递归处理嵌套结构体
		if field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && !field.IsNil() && field.Elem().Kind() == reflect.Struct) {
			recursiveMaskSensitiveFields(field, sensitiveFields)
		}
	}
}

// handleMapValues 递归处理 Map 类型的值
func handleMapValues(v reflect.Value, sensitiveFields []string) {
	for _, key := range v.MapKeys() {
		val := v.MapIndex(key)
		recursiveMaskSensitiveFields(val, sensitiveFields)
	}
}

// handleSliceElements 递归处理 Slice/数组 的每个元素
func handleSliceElements(v reflect.Value, sensitiveFields []string) {
	for i := 0; i < v.Len(); i++ {
		recursiveMaskSensitiveFields(v.Index(i), sensitiveFields)
	}
}
