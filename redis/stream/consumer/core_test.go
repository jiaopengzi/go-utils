//
// FilePath    : go-utils\redis\stream\consumer\core_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 消费者签收日志脱敏回归测试.
//

package consumer

import (
	"context"
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// ackTestRedisClient 用于测试 AckMessage 的最小 Redis 客户端实现.
type ackTestRedisClient struct {
	redis.UniversalClient
	ackErr error
}

// XAck 模拟 Redis 签收消息.
func (m *ackTestRedisClient) XAck(ctx context.Context, stream, group string, ids ...string) *redis.IntCmd {
	return redis.NewIntResult(1, m.ackErr)
}

// ackMaskPayload 用于验证敏感非字符串字段不会导致日志脱敏 panic.
type ackMaskPayload struct {
	DeleteToken bool   `json:"delete_token"`
	AccessToken string `json:"access_token"`
}

// TestBaseConsumerAckMessage_DebugLogDoesNotPanicOnBoolSensitiveField 测试 Debug 成功日志中, bool 类型敏感字段不会触发 panic.
func TestBaseConsumerAckMessage_DebugLogDoesNotPanicOnBoolSensitiveField(t *testing.T) {
	core, observedLogs := observer.New(zapcore.DebugLevel)
	undo := zap.ReplaceGlobals(zap.New(core))
	defer undo()

	consumer := &BaseConsumer[ackMaskPayload]{
		StreamName:   "stream:User",
		GroupName:    "group:User",
		ConsumerName: "consumer:User0001",
		MsgKey:       "User",
		Ctx:          context.Background(),
		Rdb:          &ackTestRedisClient{},
	}

	payload := &ackMaskPayload{
		DeleteToken: true,
		AccessToken: "raw-token",
	}

	require.NotPanics(t, func() {
		err := consumer.AckMessage("1776940227100-0", payload, true)
		require.NoError(t, err)
	})
	require.True(t, payload.DeleteToken)
	require.Equal(t, "******", payload.AccessToken)
	require.Len(t, observedLogs.All(), 1)
}

// TestBaseConsumerAckMessage_ErrorDoesNotLeakSensitiveValue 测试签收失败时, 错误文本中的敏感值会被脱敏且不会 panic.
func TestBaseConsumerAckMessage_ErrorDoesNotLeakSensitiveValue(t *testing.T) {
	undo := zap.ReplaceGlobals(zap.NewNop())
	defer undo()

	consumer := &BaseConsumer[ackMaskPayload]{
		StreamName:   "stream:User",
		GroupName:    "group:User",
		ConsumerName: "consumer:User0001",
		MsgKey:       "User",
		Ctx:          context.Background(),
		Rdb: &ackTestRedisClient{
			ackErr: errors.New("ack failed"),
		},
	}

	payload := &ackMaskPayload{
		DeleteToken: true,
		AccessToken: "raw-token",
	}

	var err error
	require.NotPanics(t, func() {
		err = consumer.AckMessage("1776940227100-0", payload, false)
	})
	require.Error(t, err)
	require.NotContains(t, err.Error(), "raw-token")
	require.Contains(t, err.Error(), "******")
	require.True(t, payload.DeleteToken)
	require.Equal(t, "******", payload.AccessToken)
}
