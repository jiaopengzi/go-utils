//
// FilePath    : go-utils\redis\bloom\core.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 基于 RedisBloom 模块的布隆过滤器封装
//

// Package bloom 布隆过滤器.
package bloom

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Bloom 过滤器结构体
type Bloom struct {
	Client    redis.UniversalClient // redis
	RedisKey  string                // Redis key
	Expansion int64                 // 每次扩展倍数
	N         int64                 // 预计元素数量
	FP        float64               // 误判率
	Ctx       context.Context       // 上下文
}

// NewBloom 创建布隆过滤器实例 用于判断元素是否存在
//   - ctx: 上下文
//   - client: 用来保存布隆过滤器数据的 redis 客户端
//   - redisKey: 用来保存布隆过滤器数据的 redis key
//   - expansion: 每次扩展倍数
//   - n: 预计元素数量个数
//   - fp: 误判率 0-1 之间, 越小误判率越低, 但是占用内存越大 一般取 0.01
func NewBloom(ctx context.Context, client redis.UniversalClient, redisKey string, expansion, n int64, fp float64) (*Bloom, error) {
	// 检查布隆过滤器是否存在
	exists, err := client.Exists(ctx, redisKey).Result()
	if err != nil {
		zap.L().Error("检查布隆过滤器是否存在失败", zap.String("key", redisKey), zap.Error(err))
		return nil, fmt.Errorf("bloom filter exists check failed: %w", err)
	}

	// 已存在则直接返回
	if exists > 0 {
		zap.L().Info("布隆过滤器已存在, 无需创建", zap.String("key", redisKey))

		return &Bloom{
			Client:    client,
			RedisKey:  redisKey,
			Expansion: expansion,
			N:         n,
			FP:        fp,
			Ctx:       ctx,
		}, nil
	}

	// 不存在则创建
	_, err = client.BFReserveExpansion(ctx, redisKey, fp, n, expansion).Result()
	if err != nil {
		// 可能已被其他实例创建(竞态), 忽略 "item exists" 错误
		if err.Error() != "ERR item exists" {
			zap.L().Error("创建布隆过滤器失败", zap.String("key", redisKey), zap.Error(err))
			return nil, fmt.Errorf("bloom filter create failed: %w", err)
		}

		zap.L().Warn("布隆过滤器已被其他实例创建, 无需创建", zap.String("key", redisKey))
	} else {
		zap.L().Info("成功创建布隆过滤器", zap.String("key", redisKey))
	}

	return &Bloom{
		Client:    client,
		RedisKey:  redisKey,
		Expansion: expansion,
		N:         n,
		FP:        fp,
		Ctx:       ctx,
	}, nil
}

// Add 添加元素到布隆过滤器
func (b *Bloom) Add(item string) error {
	// 判断 item 是否为空
	if item == "" {
		return errors.New("item cannot be empty")
	}

	_, err := b.Client.BFAdd(b.Ctx, b.RedisKey, item).Result()
	if err != nil {
		zap.L().Error("向布隆过滤器添加元素失败", zap.String("key", b.RedisKey), zap.String("item", item), zap.Error(err))
		return fmt.Errorf("bloom filter add item failed: %w", err)
	}

	return nil
}

// MAdd 批量添加元素到布隆过滤器
func (b *Bloom) MAdd(items []string) error {
	// 判断 items 是否为空
	if len(items) == 0 {
		return errors.New("items cannot be empty")
	}

	_, err := b.Client.BFMAdd(b.Ctx, b.RedisKey, items).Result()
	if err != nil {
		zap.L().Error("向布隆过滤器批量添加元素失败", zap.String("key", b.RedisKey), zap.Error(err))
		return fmt.Errorf("bloom filter madd items failed: %w", err)
	}

	return nil
}

// Test 判断元素是否存在(可能会误判, 但如果不存在则一定不存在)
func (b *Bloom) Test(testStr string) (bool, error) {
	// 判断 testStr 是否为空
	if testStr == "" {
		return false, errors.New("testStr cannot be empty")
	}

	exists, err := b.Client.BFExists(b.Ctx, b.RedisKey, testStr).Result()
	if err != nil {
		zap.L().Error("检查元素是否存在于布隆过滤器失败", zap.String("key", b.RedisKey), zap.String("item", testStr), zap.Error(err))
		return exists, fmt.Errorf("bloom filter test item failed: %w", err)
	}

	return exists, nil
}
