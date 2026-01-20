//
// FilePath    : go-utils\redis\stream\constant.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : stream 相关常量.
//

package stream

// stream、group、consumer 的命名约定
const (
	NamePrefix         = "stream:"   // 消息队列名称前缀
	GroupNamePrefix    = "group:"    // 消费者组名称前缀
	ConsumerNamePrefix = "consumer:" // 消费者名称前缀
	CreateStreamStart  = "$"         // 创建消息队列开始位置
)

// Consumer 相关常量
const (
	GenerateUniqueConsumerNameMaxCount = 200   // 同组生成唯一消费者名称最大校验次数
	RemoveConsumerForMaxCount          = 10000 // 删除消费者最大校验次数
	ConsumerMaxCount                   = 100   // 同组消费者最大数量
	ConsumerMinCount                   = 1     // 同组消费者最小数量
)
