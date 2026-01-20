//
// FilePath    : go-utils\dtovalidator\translator.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 翻译器初始化
//

// Package req 请求参数及其校验
package dtovalidator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

// Trans 定义一个全局翻译器T
var Trans ut.Translator

// GlobalValidator 定义全局验证器
var GlobalValidator *validator.Validate

// InitTrans 初始化翻译器, 传入语言环境 locale (zh 或 en)
func InitTrans(locale string) (err error) {
	// 修改 gin 框架中的 Validator 引擎属性，实现定制
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 赋值给全局验证器
		GlobalValidator = v
		// 注册一个获取 json tag 的自定义方法
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0] // 以逗号分隔，忽略后面的内容
			if name == "-" {
				return ""
			}

			return name
		})

		zhT := zh.New() // 中文翻译器
		enT := en.New() // 英文翻译器

		// 第一个参数是备用（fallback）的语言环境
		// 后面的参数是应该支持的语言环境（支持多个）
		// uni := ut.New(zhT, zhT) 也是可以的
		uni := ut.New(enT, zhT)

		// locale 通常取决于 http 请求头的 'Accept-Language'
		// var ok bool
		// 也可以使用 uni.FindTranslator(...) 传入多个locale进行查找
		Trans, ok = uni.GetTranslator(locale)
		if !ok {
			return fmt.Errorf("uni.GetTranslator(%s) failed", locale)
		}

		// 注册翻译器
		switch locale {
		case "en":
			err = enTranslations.RegisterDefaultTranslations(v, Trans)
		case "zh":
			err = zhTranslations.RegisterDefaultTranslations(v, Trans)
		default:
			err = enTranslations.RegisterDefaultTranslations(v, Trans)
		}

		// 注册自定义验证器
		for tag, entry := range EntryMap {
			if err = registerValidatorFunc(v, tag, entry.ErrMsg, entry.ValidatorFunc); err != nil {
				return err
			}
		}

		return err
	}

	return err
}

// registerValidatorFunc 根据 v 验证器注册 tag 标签的自定义校验器, fn 为校验函数
func registerValidatorFunc(v *validator.Validate, tag string, msgStr string, fn ValidatorFunc) error {
	// 注册tag自定义校验
	if err := v.RegisterValidation(tag, validator.Func(fn)); err != nil {
		return err
	}

	// 自定义错误内容
	err := v.RegisterTranslation(tag, Trans, func(ut ut.Translator) error {
		// return ut.Add(tag, "{0}"+msgStr, true) // 参考 universal-translator 的 API 文档 {0} 会自动填充字段名
		return ut.Add(tag, msgStr, true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, err := ut.T(tag, fe.Field())
		if err != nil {
			errFinal, ok := fe.(error)
			if !ok {
				return fe.Translate(ut)
			}

			if errFinal != nil {
				return errFinal.Error()
			}
		}

		return t
	})

	return err
}
