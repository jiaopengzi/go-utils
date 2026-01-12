//
// FilePath    : go-utils\shared_mutex.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 分片互斥锁, 用于高并发场景下按 key 加锁.
//

package utils

import (
	"fmt"
	"sync"
)

// ShardedMutex 是一个基于 key 的分片互斥锁, 用于减少锁竞争.
type ShardedMutex struct {
	shards []*mutexShard
	mask   uint64 // 用于快速取模(要求 shardCount 是 2 的幂)
}

// mutexShard 是单个分片的结构体, 包含一个互斥锁.
type mutexShard struct {
	mu sync.Mutex
}

// NewShardedMutex 创建一个新的分片互斥锁.
// shardCount 必须是 2 的幂(如 32, 64, 128), 否则会返回错误.
func NewShardedMutex(shardCount int) (*ShardedMutex, error) {
	// 验证 shardCount 是否为 2 的幂
	if shardCount <= 0 || (shardCount&(shardCount-1)) != 0 {
		return nil, fmt.Errorf("shardCount must be a power of 2, got %d", shardCount)
	}

	shards := make([]*mutexShard, shardCount)
	for i := range shards {
		shards[i] = &mutexShard{}
	}

	return &ShardedMutex{
		shards: shards,
		mask:   uint64(shardCount) - 1,
	}, nil
}

// Lock 根据 key 对应的分片加锁; key 通常为用户 ID、设备 ID 等唯一标识.
func (sm *ShardedMutex) Lock(key uint64) {
	shard := sm.getShard(key)
	shard.mu.Lock()
}

// Unlock 根据 key 对应的分片解锁; key 必须与 Lock 时使用的 key 相同.
func (sm *ShardedMutex) Unlock(key uint64) {
	shard := sm.getShard(key)
	shard.mu.Unlock()
}

// getShard 获取 key 对应的分片(利用位运算加速取模).
func (sm *ShardedMutex) getShard(key uint64) *mutexShard {
	// 因为 mask = shardCount - 1, 且 shardCount 是 2 的幂,
	// 所以 key & mask == key % shardCount, 但更快.
	idx := key & sm.mask
	return sm.shards[idx]
}
