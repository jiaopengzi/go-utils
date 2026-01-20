//
// FilePath    : go-utils\redis\stream\producer\core.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 生产者核心.
//

// Package core redis stream 生产者核心
package producer

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

// StreamInfo 包含流名称和流ID
type StreamInfo struct {
	Name string `json:"stream_name" binding:"required,ValidateStreamName" example:"stream:XXX"`  // 流名称
	ID   string `json:"stream_id" binding:"required,ValidateStreamID" example:"1747532798796-0"` // 消息ID
}

// Producer
type Producer[T any] interface {
	// AddMessageToStream 添加消息到 stream
	AddMessageToStream(value T) (*StreamInfo, error)
}

// MessageStateInitializer 消息状态初始化接口
// 用于在消息生产后初始化其状态(如缓存中的签收状态跟踪)
type MessageStateInitializer interface {
	// InitMessageStatus 初始化消息状态
	// streamName: 流名称, msgID: 消息ID
	InitMessageStatus(streamName, msgID string) error
}

// BaseProducer 生产者基类
type BaseProducer[T any] struct {
	StreamName       string                  // stream 名称 相同的 stream 名称的消费者将会共享消息
	MsgKey           string                  // 消息的 key 用于解析消息.
	MaxLength        int64                   // 最大消息数量,零值为不进行修剪。如果Redis 的 Stream中的条目数超过了这个限制，Redis 会删除(修剪)最旧的条目，以保持流的条目数在 stream_max_length 以内, 防止内存占用过多。
	Ctx              context.Context         // context 上下文
	Rdb              redis.UniversalClient   // Redis 客户端
	StateInitializer MessageStateInitializer // 状态初始化器
}

// AddMessageToStream 实现 Producer 接口方法, 添加消息到 stream, 并返回消息 ID
func (p *BaseProducer[T]) AddMessageToStream(value T) (*StreamInfo, error) {
	// 将 value 转换为 json 字符串
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	// jsonString := string(jsonBytes)
	// fmt.Printf("==>Producer jsonString:%v\n", jsonString)

	msgID, err := p.Rdb.XAdd(p.Ctx, &redis.XAddArgs{
		Stream: p.StreamName,                        // stream 名称
		ID:     "*",                                 // 自动创建 ID
		Values: map[string]any{p.MsgKey: jsonBytes}, // 消息内容
	}).Result()
	if err != nil {
		return nil, err
	}

	// 如果有状态初始化器, 初始化消息状态
	if p.StateInitializer != nil {
		if err = p.StateInitializer.InitMessageStatus(p.StreamName, msgID); err != nil {
			return nil, err
		}
	}

	// 如果设置了最大消息长度,则进行修剪
	if p.MaxLength > 0 {
		if err = p.Rdb.XTrimMaxLen(p.Ctx, p.StreamName, p.MaxLength).Err(); err != nil {
			return nil, err
		}
	}

	return &StreamInfo{
		Name: p.StreamName,
		ID:   msgID,
	}, nil
}
