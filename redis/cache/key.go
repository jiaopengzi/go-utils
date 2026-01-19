//
// FilePath    : go-utils\redis\cache\key.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : Redis 缓存键生成器
//

// Package cache 缓存层，负责缓存数据
package cache

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jiaopengzi/go-utils/model"
	"go.uber.org/zap"
)

// Purpose 缓存用途分类
type Purpose string

// 缓存相关的变量
var (
	prefix    = "cache" // 缓存前缀
	Delimiter = ":"     // 缓存键的分隔符
)

// SetPrefix 设置缓存前缀
func SetPrefix(p string) {
	prefix = p
}

// SetDelimiter 设置缓存键的分隔符
func SetDelimiter(d string) {
	Delimiter = d
}

// GenerateKey 生成缓存格式化的 redis 键, args 为不定长参数的是 key 的组成部分, 根据传入的顺序, 拼接在一起形成最后的 key
func GenerateKey(args ...any) string {
	// key 初始化为应用的名称
	var key strings.Builder

	key.WriteString(prefix)

	// 通过循环，遍历所有的 key 的组成部分（kids）
	for _, kid := range args {
		// 在每个 key 的部分之间添加一个分隔符
		key.WriteString(Delimiter)

		// 使用switch case根据部分的类型进行处理
		switch v := kid.(type) {
		case uint64:
			// 如果kid的类型是uint64，则将其转换为字符串并添加到key中
			key.WriteString(strconv.FormatUint(v, 10))

		case int:
			// 如果kid的类型是int，则将其转换为字符串并添加到key中
			key.WriteString(strconv.Itoa(v))

		case string:
			// 如果kid的类型是字符串，则直接添加到key中
			key.WriteString(v)

		case Purpose:
			// 如果kid的类型是dto.Purpose，则将其转换为字符串并添加到key中
			key.WriteString(string(v))

		case model.Currency:
			// 如果kid的类型是m.Currency，则将其转换为字符串并添加到key中
			fmt.Fprint(&key, v)

		default:
			// 除了以上所列的类型，其他类型都不会被添加到key中
			key.WriteString("")

			zap.L().Error("缓存的key类型不包含", zap.Any("key", kid))
		}
	}

	// 返回完整的 key
	return key.String()
}
