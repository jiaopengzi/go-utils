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
var sensitiveFields = []string{"password", "token", "secret", "captcha"}

// GetSensitiveFields 获取敏感字段关键字切片的副本, 避免外部修改原切片
func GetSensitiveFields() []string {
	fieldsCopy := make([]string, len(sensitiveFields))
	copy(fieldsCopy, sensitiveFields)
	return fieldsCopy
}

// SetSensitiveFields 设置敏感字段关键字切片, 替换原切片内容.
func SetSensitiveFields(fields []string) {
	sensitiveFields = make([]string, len(fields))
	copy(sensitiveFields, fields)
}

// MaskSensitiveFields 将传入 data 包含敏感字段关键字(包含即可,大小写不敏感)的字段值替换为 "******"
func MaskSensitiveFields(data any) {
	v := reflect.ValueOf(data)
	if !v.IsValid() {
		return
	}

	recursiveMaskSensitiveFields(v, GetSensitiveFields())
}

// recursiveMaskSensitiveFields 递归处理敏感字段加上掩码
func recursiveMaskSensitiveFields(v reflect.Value, fields []string) {
	if !v.IsValid() {
		return
	}

	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}

		v = v.Elem()
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

// isFieldSensitive 判断标识名是否包含任意敏感关键字(不区分大小写, 忽略下划线差异)
//
// 结构体字段名、JSON tag、Map key 都可能作为稳定标识使用, 因此统一做规范化后匹配.
func isFieldSensitive(identifier string, fields []string) bool {
	normalizedField := normalizeSensitiveIdentifier(identifier)
	for _, sensitiveField := range fields {
		normalizedKeyword := normalizeSensitiveIdentifier(sensitiveField)
		if strings.Contains(normalizedField, normalizedKeyword) {
			return true
		}
	}

	return false
}

// maskFieldValue 对单个字段执行掩码操作, 支持 string 和 *string 两种情况; 其他类型直接跳过.
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
		return
	}
}

// normalizeSensitiveIdentifier 统一规范化字段名, 便于忽略大小写和下划线差异.
func normalizeSensitiveIdentifier(identifier string) string {
	return strings.ReplaceAll(strings.ToLower(identifier), "_", "")
}

// getSensitiveIdentifiers 返回结构体字段可用于脱敏匹配的稳定标识, 优先使用 JSON tag.
func getSensitiveIdentifiers(fieldType reflect.StructField) []string {
	identifiers := []string{fieldType.Name}
	jsonTag := fieldType.Tag.Get("json")
	if jsonTag == "" {
		return identifiers
	}

	jsonName := strings.Split(jsonTag, ",")[0]
	if jsonName == "" || jsonName == "-" {
		return identifiers
	}

	return append(identifiers, jsonName)
}

// handleStructFields 处理结构体的每个字段：判断敏感字段并掩码, 遇到嵌套结构体时递归调用
func handleStructFields(v reflect.Value, fields []string) {
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)

		isSensitive := false
		for _, identifier := range getSensitiveIdentifiers(fieldType) {
			if isFieldSensitive(identifier, fields) {
				isSensitive = true
				break
			}
		}

		// 检查字段名是否包含任意敏感字段(不区分大小写)
		if isSensitive && field.CanSet() {
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
		if key.Kind() == reflect.String && isFieldSensitive(key.String(), fields) {
			maskedValue, ok := buildMaskedMapValue(val, v.Type().Elem())
			if ok {
				v.SetMapIndex(key, maskedValue)
				continue
			}
		}

		recursiveMaskSensitiveFields(val, fields)
	}
}

// buildMaskedMapValue 为 Map 敏感键构造掩码值, 支持 string、interface 和 *string 等常见类型.
func buildMaskedMapValue(val reflect.Value, elemType reflect.Type) (reflect.Value, bool) {
	const maskedValue = "******"

	for val.IsValid() && (val.Kind() == reflect.Pointer || val.Kind() == reflect.Interface) {
		if val.IsNil() {
			break
		}

		val = val.Elem()
	}

	switch elemType.Kind() {
	case reflect.Interface:
		return reflect.ValueOf(maskedValue), true
	case reflect.String:
		return reflect.ValueOf(maskedValue).Convert(elemType), true
	case reflect.Pointer:
		if elemType.Elem().Kind() != reflect.String {
			return reflect.Value{}, false
		}

		masked := reflect.New(elemType.Elem())
		masked.Elem().SetString(maskedValue)
		return masked, true
	default:
		if val.IsValid() && val.Kind() == reflect.String && reflect.TypeOf(maskedValue).AssignableTo(elemType) {
			return reflect.ValueOf(maskedValue), true
		}
		return reflect.Value{}, false
	}
}

// handleSliceElements 递归处理 Slice/数组 的每个元素
func handleSliceElements(v reflect.Value, fields []string) {
	for i := 0; i < v.Len(); i++ {
		recursiveMaskSensitiveFields(v.Index(i), fields)
	}
}
