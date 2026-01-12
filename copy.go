//
// FilePath    : go-utils\copy.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 拷贝相关操作
//

package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
)

// DeepCopy 将 data 递归地深拷贝任意类型的值,深拷贝到一个新的变量中
func DeepCopy[T any](data T) (T, error) {
	src := reflect.ValueOf(data)
	copyValue := reflect.New(src.Type()).Elem()

	err := copyRecursive(src, copyValue)
	if err != nil {
		var zero T
		return zero, err
	}

	result, ok := copyValue.Interface().(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("DeepCopy failed: %v", copyValue.Interface())
	}

	return result, nil
}

// copyRecursive 使用反射递归地拷贝值
func copyRecursive(src, cpy reflect.Value) error {
	// 检查 src 是否有效
	if !src.IsValid() {
		return errors.New("copyRecursive invalid value")
	}

	// 根据 src 的类型进行不同的处理
	switch src.Kind() {
	case reflect.Pointer:
		// 处理指针类型
		return copyPointer(src, cpy)
	case reflect.Struct:
		// 处理结构体类型
		return copyStruct(src, cpy)
	case reflect.Slice:
		// 处理切片类型
		return copySlice(src, cpy)
	case reflect.Map:
		// 处理 map 类型
		return copyMap(src, cpy)
	case reflect.Array:
		// 处理数组类型
		return copyArray(src, cpy)
	default:
		// 处理基本类型和其他类型，直接复制值
		return copyDefault(src, cpy)
	}
}

// copyPointer 处理指针类型的深拷贝
func copyPointer(src, cpy reflect.Value) error {
	if src.IsNil() {
		// 如果原指针为空，设置目标为该类型的零值
		cpy.Set(reflect.Zero(cpy.Type()))
		return nil
	}

	// 获取指针指向的值
	srcValue := src.Elem()

	// 创建一个新的指针并递归复制其值
	cpy.Set(reflect.New(srcValue.Type()))

	return copyRecursive(srcValue, cpy.Elem())
}

// copyStruct 处理结构体类型的深拷贝
func copyStruct(src, cpy reflect.Value) error {
	if src.Type() == reflect.TypeFor[sql.NullTime]() {
		// 特殊处理 sql.NullTime 类型
		cpy.Set(src)
		return nil
	}

	// 遍历结构体的每个字段并递归复制
	for i := 0; i < src.NumField(); i++ {
		if src.Field(i).CanSet() {
			err := copyRecursive(src.Field(i), cpy.Field(i))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// copySlice 处理切片类型的深拷贝
func copySlice(src, cpy reflect.Value) error {
	if src.IsNil() {
		// 如果原切片为空，设置目标为该类型的零值
		cpy.Set(reflect.Zero(cpy.Type()))
		return nil
	}

	// 创建一个新的切片并递归复制其元素
	cpy.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))

	for i := 0; i < src.Len(); i++ {
		err := copyRecursive(src.Index(i), cpy.Index(i))
		if err != nil {
			return err
		}
	}

	return nil
}

// copyMap 处理 map 类型的深拷贝
func copyMap(src, cpy reflect.Value) error {
	if src.IsNil() {
		// 如果原 map 为空，设置目标为该类型的零值
		cpy.Set(reflect.Zero(cpy.Type()))
		return nil
	}

	// 创建一个新的 map 并递归复制其键值对
	cpy.Set(reflect.MakeMapWithSize(src.Type(), src.Len()))

	for _, key := range src.MapKeys() {
		newValue := reflect.New(src.MapIndex(key).Type()).Elem()

		err := copyRecursive(src.MapIndex(key), newValue)
		if err != nil {
			return err
		}

		cpy.SetMapIndex(key, newValue)
	}

	return nil
}

// copyArray 处理数组类型的深拷贝
func copyArray(src, cpy reflect.Value) error {
	// 遍历数组的每个元素并递归复制
	for i := 0; i < src.Len(); i++ {
		err := copyRecursive(src.Index(i), cpy.Index(i))
		if err != nil {
			return err
		}
	}

	return nil
}

// copyDefault 处理基本类型和其他类型，直接复制值
func copyDefault(src, cpy reflect.Value) error {
	cpy.Set(src)
	return nil
}
