//
// FilePath    : go-utils\dtovalidator\main.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 请求参数校验器
//

// Package validators 请求参数校验器
package dtovalidator

import (
	"github.com/go-playground/validator/v10"
)

const (
	FieldTypeSlice = "slice" // 切片类型, 用于验证器函数中
)

// ValidatorFunc 定义一个验证器函数
type ValidatorFunc func(fl validator.FieldLevel) bool

// ValidatorEntry 验证器明细
type ValidatorEntry struct {
	ValidatorFunc ValidatorFunc // 验证器函数
	ErrMsg        string        // 错误信息
}

// EntryMap 是一个映射, 其中键是验证器的名称, 值是 ValidatorEntry 结构体
var EntryMap = make(map[string]ValidatorEntry)

// RegisterValidator 添加新的验证器到 EntryMap 中
func RegisterValidator(name string, entry ValidatorEntry) {
	EntryMap[name] = entry
}
