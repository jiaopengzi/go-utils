//
// FilePath    : go-utils\model\main.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 业务模型入口
//

package model

import (
	"sync"
)

// Tabler 模型接口, 所有模型都需要实现这个接口
type Tabler interface {

	// TableName 模型对应的数据库表名称
	TableName() string // 表名
}

// 定义模型层相关变量
var (
	models []any      // 注册的模型切片
	mu     sync.Mutex // 互斥锁 (保证并发安全)
)

// GetModels 获取所有注册的模型
func GetModels() []any {
	return models
}

// RegisterModel 将模型注册到模型切片中, 用于后续的初始化.
func RegisterModel(model Tabler) {
	mu.Lock()         // 加锁
	defer mu.Unlock() // 解锁

	models = append(models, model) // 注册模型
}
