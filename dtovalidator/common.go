//
// FilePath    : go-utils\dtovalidator\common.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 通用的 DTO 校验器
//

package dtovalidator

import (
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/jiaopengzi/go-utils/model"
	"github.com/jiaopengzi/go-utils/types"
)

// init 初始化注册校验器
func init() {
	RegisterValidator("ValidatePaginate", ValidatorEntry{
		ValidatorFunc: ValidatePaginate,
		ErrMsg:        "分页参数错误,参数需要正整数.",
	})

	RegisterValidator("ValidateInt", ValidatorEntry{
		ValidatorFunc: ValidateInt,
		ErrMsg:        "参数需要正整数.",
	})

	RegisterValidator("ValidateIntYear", ValidatorEntry{
		ValidatorFunc: ValidateIntYear,
		ErrMsg:        "请输入正确的年份:1000-9999",
	})

	RegisterValidator("ValidateIntMonth", ValidatorEntry{
		ValidatorFunc: ValidateIntMonth,
		ErrMsg:        "请输入正确的月份:1-12",
	})

	RegisterValidator("ValidateJSONUint64", ValidatorEntry{
		ValidatorFunc: ValidateJSONUint64,
		ErrMsg:        "参数需要正整数.",
	})

	RegisterValidator("ValidateJSONUint64Slice", ValidatorEntry{
		ValidatorFunc: ValidateJSONUint64Slice,
		ErrMsg:        "参数需要正整数列表.",
	})

	RegisterValidator("ValidateJSONInt64", ValidatorEntry{
		ValidatorFunc: ValidateJSONInt64,
		ErrMsg:        "参数需要正整数.",
	})

	RegisterValidator("ValidateJSONInt64Slice", ValidatorEntry{
		ValidatorFunc: ValidateJSONInt64Slice,
		ErrMsg:        "参数需要正整数列表.",
	})

	RegisterValidator("ValidateTrimContent", ValidatorEntry{
		ValidatorFunc: ValidateTrimContent,
		ErrMsg:        "请输入正确内容,首尾不包含空格",
	})
	RegisterValidator("ValidateCurrency", ValidatorEntry{
		ValidatorFunc: ValidateCurrency,
		ErrMsg:        "请输入正确的货币类型.",
	})
	RegisterValidator("ValidateCSR", ValidatorEntry{
		ValidatorFunc: ValidateCSR,
		ErrMsg:        "请输入正确的证书请求(CSR).",
	})
}

// ValidatePaginate 分页参数校验
func ValidatePaginate(fl validator.FieldLevel) bool {
	// 校验分页参数 判断是否为正整数
	page := fl.Field().Int()
	return page >= 1
}

// ValidateInt 校验正整数
func ValidateInt(fl validator.FieldLevel) bool {
	// 校验正整数
	_, ok := ValidateAndGetJSONInt(fl)

	return ok
}

func ValidateAndGetJSONInt(fl validator.FieldLevel) (int64, bool) {
	// 校验正整数
	value := fl.Field().Int()

	return value, value > 0
}

// ValidateIntYear 校验年份
func ValidateIntYear(fl validator.FieldLevel) bool {
	v, ok := ValidateAndGetJSONInt(fl)
	if !ok {
		return false
	}

	return v >= 1000 && v <= 9999
}

// ValidateIntMonth 校验月份
func ValidateIntMonth(fl validator.FieldLevel) bool {
	v, ok := ValidateAndGetJSONInt(fl)
	if !ok {
		return false
	}

	return v >= 1 && v <= 12
}

// ValidateJSONUint64 校验正整数
func ValidateJSONUint64(fl validator.FieldLevel) bool {
	// 校验正整数
	_, ok := ValidateAndGetJSONUint64(fl)
	return ok
}

// ValidateJSONInt64 校验正整数
func ValidateJSONInt64(fl validator.FieldLevel) bool {
	// 校验正整数
	_, ok := ValidateAndGetJSONInt(fl)
	return ok
}

// ValidateAndGetJSONUint64 校验Uint64
func ValidateAndGetJSONUint64(fl validator.FieldLevel) (uint64, bool) {
	// 校验正整数
	value, ok := fl.Field().Interface().(types.JSONUint64)
	if !ok {
		return 0, false
	}

	if uint64(value) > 0 {
		return uint64(value), true
	}

	return 0, false
}

// ValidateJSONUint64Slice 校验正整数列表
func ValidateJSONUint64Slice(fl validator.FieldLevel) bool {
	_, ok := ValidateAndGetJSONUint64Slice(fl)
	return ok
}

// ValidateAndGetJSONUint64Slice 校验并获取 []uint64
func ValidateAndGetJSONUint64Slice(fl validator.FieldLevel) ([]any, bool) {
	// 判断是否为空
	if fl.Field().String() == "" {
		return nil, false
	}
	// 判断是否为切片
	if fl.Field().Kind().String() != FieldTypeSlice {
		return nil, false
	}
	// 判断切片长度
	if fl.Field().Len() == 0 {
		return nil, false
	}

	var uint64Slice []any

	values, ok := fl.Field().Interface().(types.JSONUint64Slice)
	if !ok {
		return nil, false
	}

	for _, value := range values {
		// 是否能解析为正整数
		if value == 0 {
			return nil, false
		}

		uint64Slice = append(uint64Slice, value)
	}

	return uint64Slice, true
}

// ValidateJSONInt64Slice 校验正整数列表
func ValidateJSONInt64Slice(fl validator.FieldLevel) bool {
	_, ok := ValidateAndGetJSONInt64Slice(fl)
	return ok
}

// ValidateAndGetJSONInt64Slice 校验并获取 []int64
func ValidateAndGetJSONInt64Slice(fl validator.FieldLevel) ([]any, bool) {
	// 判断是否为空
	if fl.Field().String() == "" {
		return nil, false
	}
	// 判断是否为切片
	if fl.Field().Kind().String() != FieldTypeSlice {
		return nil, false
	}
	// 判断切片长度
	if fl.Field().Len() == 0 {
		return nil, false
	}

	var int64Slice []any

	values, ok := fl.Field().Interface().(types.JSONInt64Slice)
	if !ok {
		return nil, false
	}

	for _, value := range values {
		// 是否能解析为正整数
		if value == 0 {
			return nil, false
		}

		int64Slice = append(int64Slice, value)
	}

	return int64Slice, true
}

// ValidateTrimContent 校验内容是否为空，首位是否包含空格
func ValidateTrimContent(fl validator.FieldLevel) bool {
	content := fl.Field().String()
	// 判断content是否为空
	if content == "" {
		return false
	}

	// 判断content是否首位是否包含空格，包含则返回false
	if content[0] == ' ' || content[len(content)-1] == ' ' {
		return false
	}

	return true
}

// ValidateEnumInt64 通用的枚举校验函数 int64
func ValidateEnumInt64(fl validator.FieldLevel, validValues ...int64) bool {
	v, ok := ValidateAndGetJSONInt(fl)
	if !ok {
		return false
	}

	return slices.Contains(validValues, v)
}

// ValidateEnumString 通用的枚举校验函数 string
func ValidateEnumString(fl validator.FieldLevel, validValues ...string) bool {
	v := fl.Field().String()
	if v == "" {
		return false
	}

	return slices.Contains(validValues, v)
}

// ValidateCurrency 校验货币类型
func ValidateCurrency(fl validator.FieldLevel) bool {
	return ValidateEnumInt64(fl,
		int64(model.CurrencyCNY),
		int64(model.CurrencyUSD),
		int64(model.CurrencyEUR),
		int64(model.CurrencyGBP),
		int64(model.CurrencyHKD),
		int64(model.CurrencyTWD),
		int64(model.CurrencySGD),
		int64(model.CurrencyRUB),
	)
}

// ValidateCSR 校验证书请求(CSR)
// -----BEGIN CERTIFICATE REQUEST-----
// MIG6MG4CAQAwFDESMBAGA1UEAxMJbG9jYWxob3N0MCowBQYDK2VwAyEAr2h/kLhK
// -----END CERTIFICATE REQUEST-----
func ValidateCSR(fl validator.FieldLevel) bool {
	csr := strings.TrimSpace(fl.Field().String())
	if csr == "" {
		return false
	}

	const (
		csrBegin = "-----BEGIN CERTIFICATE REQUEST-----"
		csrEnd   = "-----END CERTIFICATE REQUEST-----"
	)

	// 判断是否 以 -----BEGIN CERTIFICATE REQUEST----- 开头
	if !stringHasPrefixIgnoreCase(csr, csrBegin) {
		return false
	}

	// 判断是否 以 -----END CERTIFICATE REQUEST----- 结尾
	if !stringHasSuffixIgnoreCase(csr, csrEnd) {
		return false
	}

	return true
}

// stringHasPrefixIgnoreCase 判断 s 是否以 prefix 开头，忽略大小写
func stringHasPrefixIgnoreCase(s, prefix string) bool {
	if prefix == "" {
		return true
	}

	if len(prefix) > len(s) {
		return false
	}

	return strings.EqualFold(s[:len(prefix)], prefix)
}

// stringHasSuffixIgnoreCase 判断 s 是否以 suffix 结尾，忽略大小写
func stringHasSuffixIgnoreCase(s, suffix string) bool {
	if suffix == "" {
		return true
	}

	if len(suffix) > len(s) {
		return false
	}

	return strings.EqualFold(s[len(s)-len(suffix):], suffix)
}
