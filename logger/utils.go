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

// sensitiveFields 敏感字段关键字切片
var sensitiveFields = []string{"password", "token", "secret"}

// GetSensitiveFields 获取敏感字段关键字切片的副本, 避免外部修改原切片
func GetSensitiveFields() []string {
	fieldsCopy := make([]string, len(sensitiveFields))
	copy(fieldsCopy, sensitiveFields)
	return fieldsCopy
}

// SetSensitiveFields 设置敏感字段关键字切片, 替换原切片内容
func SetSensitiveFields(fields []string) {
	sensitiveFields = make([]string, len(fields))
	copy(sensitiveFields, fields)
}

// MaskSensitiveFields 将传入 data 包含敏感字段关键字(包含即可,大小写不敏感)的字段值替换为 "******"
func MaskSensitiveFields(data any) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	recursiveMaskSensitiveFields(v, GetSensitiveFields())
}

// recursiveMaskSensitiveFields 递归处理敏感字段加上掩码
func recursiveMaskSensitiveFields(v reflect.Value, fields []string) {
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
		handleStructFields(v, fields)
	case reflect.Map:
		handleMapValues(v, fields)
	case reflect.Slice:
		handleSliceElements(v, fields)
	}
}

// isFieldSensitive 判断字段名是否包含任意敏感关键字(不区分大小写, 忽略下划线差异)
//
// Go 结构体字段名为 PascalCase(如 AppKey), 转为小写后为 appkey;
// 而用户配置的敏感关键字可能为 snake_case(如 app_key), 需要去掉下划线后再做包含匹配.
func isFieldSensitive(lowerFieldName string, fields []string) bool {
	// 去掉字段名中的下划线(防御性处理)
	normalizedField := strings.ReplaceAll(lowerFieldName, "_", "")
	for _, sensitiveField := range fields {
		normalizedKeyword := strings.ReplaceAll(strings.ToLower(sensitiveField), "_", "")
		if strings.Contains(normalizedField, normalizedKeyword) {
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
func handleStructFields(v reflect.Value, fields []string) {
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)

		// 将字段名转换为小写以进行大小写不敏感的匹配
		lowerFieldName := strings.ToLower(fieldType.Name)

		// 检查字段名是否包含任意敏感字段(不区分大小写)
		if isFieldSensitive(lowerFieldName, fields) && field.CanSet() {
			maskFieldValue(field)
		}

		// 递归处理嵌套结构体、切片、Map 字段
		switch field.Kind() {
		case reflect.Struct:
			recursiveMaskSensitiveFields(field, fields)
		case reflect.Pointer:
			if !field.IsNil() {
				recursiveMaskSensitiveFields(field, fields)
			}
		case reflect.Slice:
			handleSliceElements(field, fields)
		case reflect.Map:
			handleMapValues(field, fields)
		}
	}
}

// handleMapValues 递归处理 Map 类型的值
func handleMapValues(v reflect.Value, fields []string) {
	for _, key := range v.MapKeys() {
		val := v.MapIndex(key)
		recursiveMaskSensitiveFields(val, fields)
	}
}

// handleSliceElements 递归处理 Slice/数组 的每个元素
func handleSliceElements(v reflect.Value, fields []string) {
	for i := 0; i < v.Len(); i++ {
		recursiveMaskSensitiveFields(v.Index(i), fields)
	}
}
