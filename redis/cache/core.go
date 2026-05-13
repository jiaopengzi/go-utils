//
// FilePath    : go-utils\redis\cache\core.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 缓存核心工具。
//

// Package cache 缓存核心工具
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cacher 缓存通用工具接口
type Cacher interface {
	HMSet(ctx context.Context, key string, fields map[string]any) error                                        // 增 缓存数据 hash
	HMGet(ctx context.Context, key string, fields ...string) ([]any, error)                                    // 获取缓存数据 hash
	HSet(ctx context.Context, key, field string, value any) error                                              // 增 缓存数据 hash
	HGet(ctx context.Context, key, field string) (string, error)                                               // 获取缓存数据 hash
	HDel(ctx context.Context, key string, fields ...string) error                                              // 删 删除缓存数据 hash
	HGetAll(ctx context.Context, key string) (map[string]string, error)                                        // 获取所有缓存数据 hash
	SetBool(ctx context.Context, key string, value bool, duration time.Duration) error                         // 增 缓存数据 布尔
	SetString(ctx context.Context, key, value string, duration time.Duration) error                            // 增 缓存数据 纯字符串
	SetStringNX(ctx context.Context, key, value string, duration time.Duration) (bool, error)                  // 原子写入纯字符串，仅在 key 不存在时成功
	SetStringWithStruct(ctx context.Context, key string, value any, duration time.Duration) error              // 增 缓存数据 结构体
	GetBool(ctx context.Context, key string) (bool, error)                                                     // 获取缓存数据 纯字符串
	GetString(ctx context.Context, key string) (string, error)                                                 // 获取缓存数据 纯字符串
	GetStringWithStruct(ctx context.Context, key string, value any) error                                      // 获取缓存数据 结构体
	CheckString(ctx context.Context, key, str string) (bool, error)                                            // 检查key对应的字符串是否等于 str
	CheckWithStruct(ctx context.Context, key string, value any) (bool, error)                                  // 检查对应key的字符串是否等于 value
	SAdd(ctx context.Context, key string, member any) error                                                    // 添加字符串到 缓存 set中
	SRem(ctx context.Context, key string, members ...any) error                                                // 删除缓存 set 数据
	SIsMember(ctx context.Context, key, str string) (bool, error)                                              // 检查字符串是否在 缓存 set中
	GetSets(ctx context.Context, key string) ([]string, error)                                                 // 获取缓存 set 数据
	SetCounter(ctx context.Context, key string, value int64, duration time.Duration) error                     // 设置计数器的初始值
	SetCounterNX(ctx context.Context, key string, value int64, duration time.Duration) (bool, error)           // 原子设置计数器初始值，仅在 key 不存在时成功
	IncrementCounter(ctx context.Context, key string, duration time.Duration, overrideTTL bool) (int64, error) // 递增计数器
	DecrementCounter(ctx context.Context, key string, duration time.Duration, overrideTTL bool) (int64, error) // 递减计数器
	GetCounterValue(ctx context.Context, key string) (int64, error)                                            // 获取计数器的值
	GetKeyTll(ctx context.Context, key string) (time.Duration, error)                                          // 获取 key 的剩余有效期
	Del(ctx context.Context, key string) error                                                                 // 删 删除缓存数据
	DelKeysWithPrefix(ctx context.Context, prefix string) error                                                // 删除指定前缀的所有 key
	ZAdd(ctx context.Context, key string, members ...redis.Z) error                                            // 增加 zset 数据
	ZRem(ctx context.Context, key string, members ...any) error                                                // 删除 zset 数据
	ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error)                    // 获取 zset 数据(包含分数)
	ZCard(ctx context.Context, key string) (int64, error)                                                      // 获取 zset 数据个数
	XInfoGroups(ctx context.Context, key string) *redis.XInfoGroupsCmd                                         // 获取 stream 的所有组信息
}

// Client 缓存客户端
type Client struct {
	// Client redis 客户端
	Client redis.UniversalClient
}

// NewClient 创建缓存客户端实例
func NewClient(client redis.UniversalClient) *Client {
	return &Client{
		Client: client,
	}
}

// HMSet 实现 Cacher 接口 HMSet 方法
func (c *Client) HMSet(ctx context.Context, key string, fields map[string]any) error {
	return c.Client.HMSet(ctx, key, fields).Err()
}

// HMGet 实现 Cacher 接口 HMGet 方法 获取缓存数据 hash 多个字段
func (c *Client) HMGet(ctx context.Context, key string, fields ...string) ([]any, error) {
	return c.Client.HMGet(ctx, key, fields...).Result()
}

// HSet 实现 Cacher 接口 HSet 方法 增 缓存数据 hash 单个字段
func (c *Client) HSet(ctx context.Context, key, field string, value any) error {
	return c.Client.HSet(ctx, key, field, value).Err()
}

// HGet 实现 Cacher 接口 HGet 方法 获取缓存数据 hash 单个字段
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	return c.Client.HGet(ctx, key, field).Result()
}

// HDel 实现 Cacher 接口 HDel 方法 删除缓存数据 hash 多个字段
func (c *Client) HDel(ctx context.Context, key string, fields ...string) error {
	return c.Client.HDel(ctx, key, fields...).Err()
}

// HGetAll 实现 Cacher 接口 HGetAll 方法 获取所有缓存数据 hash
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.Client.HGetAll(ctx, key).Result()
}

// SetString 实现 Cacher 接口 SetString 方法 增 缓存数据 纯字符串
func (c *Client) SetString(ctx context.Context, key, value string, duration time.Duration) error {
	return c.Client.Set(ctx, key, value, duration).Err()
}

// SetStringNX 实现 Cacher 接口 SetStringNX 方法 原子写入纯字符串，仅在 key 不存在时成功.
func (c *Client) SetStringNX(ctx context.Context, key, value string, duration time.Duration) (bool, error) {
	return c.Client.SetNX(ctx, key, value, duration).Result()
}

// SetBool 实现 Cacher 接口 SetBool 方法 增 缓存数据 布尔
func (c *Client) SetBool(ctx context.Context, key string, value bool, duration time.Duration) error {
	boolStr := fmt.Sprintf("%v", value) // 将布尔值转成字符串
	// 将信息写入 cache
	return c.Client.Set(ctx, key, boolStr, duration).Err()
}

// SetStringWithStruct 实现 Cacher 接口 SetStringWithStruct 方法 增 缓存数据 结构体
func (c *Client) SetStringWithStruct(ctx context.Context, key string, value any, duration time.Duration) error {
	// 将 value 序列化为 JSON 格式
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return err
	}
	// 将信息写入 cache
	return c.Client.Set(ctx, key, string(valueJSON), duration).Err()
}

// GetBool 实现 Cacher 接口 GetBool 方法 获取缓存数据 纯字符串
func (c *Client) GetBool(ctx context.Context, key string) (bool, error) {
	// 从缓存中获取
	boolStr, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		return false, err
	}
	// 将字符串转换为布尔值
	return boolStr == "true", nil
}

// GetString 实现 Cacher 接口 GetString 方法 获取缓存数据 纯字符串
func (c *Client) GetString(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

// GetStringWithStruct 实现 Cacher 接口 GetStringWithStruct 方法 获取缓存数据 结构体
func (c *Client) GetStringWithStruct(ctx context.Context, key string, value any) error {
	// 从 Redis 中获取 Value 的 JSON 字符串
	valueJSON, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	// 将 JSON 字符串反序列化为 value 结构
	return json.Unmarshal([]byte(valueJSON), value)
}

// CheckString 实现 Cacher 接口 CheckString 方法 检查对应key的字符串是否等于 str
func (c *Client) CheckString(ctx context.Context, key, str string) (bool, error) {
	val, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return val == str, nil
}

// CheckWithStruct 实现 Cacher 接口 CheckWithStruct 方法 检查对应key的字符串是否等于 value
func (c *Client) CheckWithStruct(ctx context.Context, key string, value any) (bool, error) {
	// 从 Redis 中获取 Value 的 JSON 字符串
	valueJSONSrc, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		return false, err
	}
	// 将 value 序列化为 JSON 格式
	valueJSONTar, err := json.Marshal(value)
	if err != nil {
		return false, err
	}

	return valueJSONSrc == string(valueJSONTar), nil
}

// SAdd 实现 Cacher 接口 SAdd 方法 添加字符串到 缓存 set中
func (c *Client) SAdd(ctx context.Context, key string, member any) error {
	return c.Client.SAdd(ctx, key, member).Err()
}

// SRem 实现 Cacher 接口 SRem 方法 删除缓存 set 数据
func (c *Client) SRem(ctx context.Context, key string, members ...any) error {
	return c.Client.SRem(ctx, key, members...).Err()
}

// SIsMember 实现 Cacher 接口 SIsMember 方法 检查字符串是否在 缓存 set中
func (c *Client) SIsMember(ctx context.Context, key, str string) (bool, error) {
	return c.Client.SIsMember(ctx, key, str).Result()
}

// GetSets 实现 Cacher 接口 GetSets 方法 获取缓存 set 数据
func (c *Client) GetSets(ctx context.Context, key string) ([]string, error) {
	// set 类型
	return c.Client.SMembers(ctx, key).Result()
}

// SetCounter 实现 Cacher 接口 SetCounter 方法 设置计数器的初始值
func (c *Client) SetCounter(ctx context.Context, key string, value int64, duration time.Duration) error {
	return c.Client.Set(ctx, key, value, duration).Err()
}

// SetCounterNX 实现 Cacher 接口 SetCounterNX 方法, 仅在 key 不存在时原子设置计数器初始值.
func (c *Client) SetCounterNX(ctx context.Context, key string, value int64, duration time.Duration) (bool, error) {
	return c.Client.SetNX(ctx, key, value, duration).Result()
}

// IncrementCounter 实现 Cacher 接口 IncrementCounter 方法 计数器，每次调用加一, 根据 overrideTTL 判断是否覆盖原有 TTL
func (c *Client) IncrementCounter(ctx context.Context, key string, duration time.Duration, overrideTTL bool) (int64, error) {
	// 使用通用事务函数来降低圈复杂度
	return c.performCounterTx(ctx, key, duration, overrideTTL, 1)
}

// DecrementCounter 实现 Cacher 接口 DecrementCounter 方法 计数器，每次调用减一, 根据 overrideTTL 判断是否覆盖原有 TTL
func (c *Client) DecrementCounter(ctx context.Context, key string, duration time.Duration, overrideTTL bool) (int64, error) {
	// 使用通用事务函数来降低圈复杂度
	return c.performCounterTx(ctx, key, duration, overrideTTL, -1)
}

// performCounterTx 执行计数器增减并处理 TTL。
// 这里不再使用 WATCH 事务, 避免高并发下因乐观锁冲突把正常计数放大为服务器错误。
// delta: +1 表示递增, -1 表示递减。
func (c *Client) performCounterTx(ctx context.Context, key string, duration time.Duration, overrideTTL bool, delta int64) (int64, error) {
	var (
		result int64
		err    error
	)

	switch delta {
	case 1:
		result, err = c.Client.Incr(ctx, key).Result()
	case -1:
		result, err = c.Client.Decr(ctx, key).Result()
	default:
		return 0, fmt.Errorf("unsupported delta: %d", delta)
	}
	if err != nil {
		return 0, err
	}

	if duration <= 0 {
		return result, nil
	}

	if overrideTTL {
		if err = c.Client.Expire(ctx, key, duration).Err(); err != nil {
			return 0, err
		}

		return result, nil
	}

	// 仅在 key 当前没有 TTL 时设置窗口，避免每次请求把限流窗口向后推移。
	if err = c.Client.ExpireNX(ctx, key, duration).Err(); err != nil {
		return 0, err
	}

	return result, nil
}

// GetCounterValue 实现 Cacher 接口 GetCounterValue 方法 获取计数器的值
func (c *Client) GetCounterValue(ctx context.Context, key string) (int64, error) {
	val, err := c.Client.Get(ctx, key).Int64()
	if err != nil {
		return 0, err
	}

	return val, nil
}

// GetKeyTll 实现 Cacher 接口 GetKeyTll 方法 获取 key 的剩余有效期
func (c *Client) GetKeyTll(ctx context.Context, key string) (time.Duration, error) {
	return c.Client.TTL(ctx, key).Result()
}

// Del 实现 Cacher 接口 Del 方法 删除缓存数据
func (c *Client) Del(ctx context.Context, key string) error {
	// 如果已经有 key 就清除
	return c.Client.Del(ctx, key).Err()
}

// DelKeysWithPrefix 实现 Cacher 接口 DelKeysWithPrefix 方法 删除指定前缀的所有 key
func (c *Client) DelKeysWithPrefix(ctx context.Context, prefix string) error {
	var (
		cursor uint64   // 游标
		keys   []string // key 列表
		err    error    // 错误信息
	)

	for {
		// 扫描所有符合条件的 key
		keys, cursor, err = c.Client.Scan(ctx, cursor, prefix+"*", 0).Result()
		if err != nil {
			return err
		}

		// 如果有 key 就删除
		if len(keys) > 0 {
			// fmt.Printf("==>keys:%v\n", keys)
			if err := c.Client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		// 如果游标为 0，表示已经扫描完毕
		if cursor == 0 {
			break
		}
	}

	return nil
}

// ZAdd 实现 Cacher 接口 ZAdd 方法 增加 zset 数据
func (c *Client) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return c.Client.ZAdd(ctx, key, members...).Err()
}

// ZRem 实现 Cacher 接口 ZRem 方法 删除 zset 数据
func (c *Client) ZRem(ctx context.Context, key string, members ...any) error {
	return c.Client.ZRem(ctx, key, members...).Err()
}

// ZRange 实现 Cacher 接口 ZRange 方法 获取 zset 数据(包含分数)
func (c *Client) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return c.Client.ZRangeWithScores(ctx, key, start, stop).Result()
}

// ZCard 实现 Cacher 接口 ZCard 方法 获取 zset 数据个数
func (c *Client) ZCard(ctx context.Context, key string) (int64, error) {
	return c.Client.ZCard(ctx, key).Result()
}

// XInfoGroups 实现 Cacher 接口 XInfoGroups 方法 获取 stream 的所有组信息
func (c *Client) XInfoGroups(ctx context.Context, key string) *redis.XInfoGroupsCmd {
	// 返回一个 XInfoGroupsCmd 命令对象
	return c.Client.XInfoGroups(ctx, key)
}
