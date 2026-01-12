//
// FilePath    : go-utils\compare.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 通用比较工具
//

package utils

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
)

// IsSlicesEqual  判断两个切片是否相等
func IsSlicesEqual[T any](sliceSrc, sliceTar []T) bool {
	// 处理空切片的情况
	if len(sliceSrc) == 0 && len(sliceTar) == 0 {
		return true
	}

	// 判断长度是否相等
	if len(sliceSrc) != len(sliceTar) {
		return false
	}

	// 判断元素是否相等
	for i, v := range sliceSrc {
		if !reflect.DeepEqual(v, sliceTar[i]) {
			return false
		}
	}

	return true
}

// FieldMismatchError 自定义错误类型，用于描述字段不匹配的错误
type FieldMismatchError struct {
	FieldName string
	Index     int
}

// Error 实现 error 接口 Error 方法
func (e *FieldMismatchError) Error() string {
	return fmt.Sprintf("field %s at index %d does not match", e.FieldName, e.Index)
}

// IsSlicesEqualByField 判断两个切片是否相等，指定结构体的某些字段需要相等,如果指定字段外其他的字段不相等就返回错误.
func IsSlicesEqualByField[T any](sliceSrc, sliceTar []T, fieldNames []string) (bool, error) {
	// 处理空切片
	if (sliceSrc == nil) != (sliceTar == nil) {
		return false, errors.New("one of the slices is nil")
	}

	// 判断长度是否相等
	if len(sliceSrc) != len(sliceTar) {
		return false, errors.New("slices have different lengths")
	}

	// 判断元素是否相等
	for i := range sliceSrc {
		srcValue := reflect.Indirect(reflect.ValueOf(sliceSrc[i]))
		tarValue := reflect.Indirect(reflect.ValueOf(sliceTar[i]))

		// 优先比较剩余字段的值
		for j := range srcValue.NumField() {
			fieldName := srcValue.Type().Field(j).Name
			if contains(fieldNames, fieldName) {
				continue
			}

			if !reflect.DeepEqual(srcValue.Field(j).Interface(), tarValue.Field(j).Interface()) {
				return false, &FieldMismatchError{
					FieldName: fieldName,
					Index:     i,
				}
			}
		}

		// 比较指定字段的值
		for _, fieldName := range fieldNames {
			srcField := srcValue.FieldByName(fieldName)
			tarField := tarValue.FieldByName(fieldName)

			if !srcField.IsValid() || !tarField.IsValid() {
				return false, fmt.Errorf("field %s does not exist", fieldName)
			}

			if !reflect.DeepEqual(srcField.Interface(), tarField.Interface()) {
				return false, nil
			}
		}
	}

	return true, nil
}

// contains 判断切片是否包含某个元素
func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
