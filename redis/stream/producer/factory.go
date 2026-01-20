//
// FilePath    : go-utils\redis\stream\producer\factory.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 生产者工厂.
//

package producer

import (
	"context"

	_stream "github.com/jiaopengzi/go-utils/redis/stream"
	"github.com/redis/go-redis/v9"
)

// ManageProducers 通过配置初始化并管理生产者
//   - msgKey: 消息键
//   - maxLength: 最大消息数量
//   - rdb: Redis 客户端
//   - initializer: 消息状态初始化器
//
// 返回 Producer 接口
func ManageProducers[T any](msgKey string, maxLength int64, rdb redis.UniversalClient, initializer MessageStateInitializer) Producer[T] {
	return &BaseProducer[T]{
		StreamName:       _stream.NamePrefix + msgKey, // 消息队列名称
		MsgKey:           msgKey,                      // 消息的 key 用于解析消息.
		MaxLength:        maxLength,                   // 最大消息数量
		Ctx:              context.Background(),        // 默认使用背景上下文
		Rdb:              rdb,                         // Redis 客户端
		StateInitializer: initializer,                 // 状态初始化器
	}
}
