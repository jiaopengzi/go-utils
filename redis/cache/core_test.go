//
// FilePath    : go-utils\redis\cache\core_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 缓存计数器并发回归测试
//

package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// newCacheClientForTest 创建测试使用的缓存客户端.
func newCacheClientForTest(t *testing.T) (*Client, *miniredis.Miniredis, func()) {
	t.Helper()

	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("启动 miniredis 失败: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: server.Addr()})
	client := NewClient(redisClient)

	cleanup := func() {
		_ = redisClient.Close()
		server.Close()
	}

	return client, server, cleanup
}

// TestIncrementCounterConcurrentNoTxFailure 验证高并发计数不会再因事务冲突失败.
func TestIncrementCounterConcurrentNoTxFailure(t *testing.T) {
	client, _, cleanup := newCacheClientForTest(t)
	defer cleanup()

	const total = 128
	ctx := context.Background()
	start := make(chan struct{})
	errCh := make(chan error, total)
	valueCh := make(chan int64, total)

	var wg sync.WaitGroup
	for range total {
		wg.Go(func() {
			<-start

			value, err := client.IncrementCounter(ctx, "counter:test", 5*time.Second, false)
			if err != nil {
				errCh <- err
				return
			}

			valueCh <- value
		})
	}

	close(start)
	wg.Wait()
	close(errCh)
	close(valueCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("期望并发递增成功, 实际报错: %v", err)
		}
	}

	seen := make(map[int64]struct{}, total)
	for value := range valueCh {
		seen[value] = struct{}{}
	}

	if len(seen) != total {
		t.Fatalf("期望拿到 %d 个唯一计数结果, 实际为 %d", total, len(seen))
	}

	finalValue, err := client.GetCounterValue(ctx, "counter:test")
	if err != nil {
		t.Fatalf("获取最终计数值失败: %v", err)
	}
	if finalValue != total {
		t.Fatalf("期望最终计数值为 %d, 实际为 %d", total, finalValue)
	}

	ttl, err := client.GetKeyTll(ctx, "counter:test")
	if err != nil {
		t.Fatalf("获取 TTL 失败: %v", err)
	}
	if ttl <= 0 {
		t.Fatalf("期望计数器 TTL 已设置, 实际为 %v", ttl)
	}
}

// TestIncrementCounterPreserveTTLWindow 验证 overrideTTL=false 时不会在后续请求中重置 TTL 窗口.
func TestIncrementCounterPreserveTTLWindow(t *testing.T) {
	client, server, cleanup := newCacheClientForTest(t)
	defer cleanup()

	ctx := context.Background()
	_, err := client.IncrementCounter(ctx, "counter:ttl", 5*time.Second, false)
	if err != nil {
		t.Fatalf("首次递增失败: %v", err)
	}

	firstTTL, err := client.GetKeyTll(ctx, "counter:ttl")
	if err != nil {
		t.Fatalf("获取首次 TTL 失败: %v", err)
	}

	server.FastForward(2 * time.Second)
	_, err = client.IncrementCounter(ctx, "counter:ttl", 5*time.Second, false)
	if err != nil {
		t.Fatalf("第二次递增失败: %v", err)
	}

	secondTTL, err := client.GetKeyTll(ctx, "counter:ttl")
	if err != nil {
		t.Fatalf("获取第二次 TTL 失败: %v", err)
	}

	if secondTTL >= firstTTL {
		t.Fatalf("期望 TTL 窗口不被重置, 实际 firstTTL=%v, secondTTL=%v", firstTTL, secondTTL)
	}
}

// TestSetStringNXConcurrentOnlyOneWinner 验证并发写入同一个 key 时仅首个请求成功.
func TestSetStringNXConcurrentOnlyOneWinner(t *testing.T) {
	client, _, cleanup := newCacheClientForTest(t)
	defer cleanup()

	const total = 64
	ctx := context.Background()
	start := make(chan struct{})
	resultCh := make(chan bool, total)
	errCh := make(chan error, total)

	var wg sync.WaitGroup
	for i := range total {
		value := fmt.Sprintf("value-%d", i)

		wg.Go(func() {
			<-start

			ok, err := client.SetStringNX(ctx, "setnx:test", value, 5*time.Second)
			if err != nil {
				errCh <- err
				return
			}

			resultCh <- ok
		})
	}

	close(start)
	wg.Wait()
	close(resultCh)
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("期望 SetStringNX 并发成功, 实际报错: %v", err)
		}
	}

	var successCount int
	for ok := range resultCh {
		if ok {
			successCount++
		}
	}

	if successCount != 1 {
		t.Fatalf("期望仅 1 个并发请求原子写入成功, 实际为 %d", successCount)
	}

	value, err := client.GetString(ctx, "setnx:test")
	if err != nil {
		t.Fatalf("获取 SetStringNX 最终值失败: %v", err)
	}
	if value == "" {
		t.Fatal("期望 SetStringNX 已写入值, 实际为空")
	}

	ttl, err := client.GetKeyTll(ctx, "setnx:test")
	if err != nil {
		t.Fatalf("获取 SetStringNX TTL 失败: %v", err)
	}
	if ttl <= 0 {
		t.Fatalf("期望 SetStringNX 已设置 TTL, 实际为 %v", ttl)
	}
	if ttl > 5*time.Second {
		t.Fatalf("期望 SetStringNX TTL 不超过写入值, 实际为 %v", ttl)
	}
}

// TestSetCounterNXConcurrentOnlyOneWinner 验证并发初始化同一个计数器时仅首个请求成功.
func TestSetCounterNXConcurrentOnlyOneWinner(t *testing.T) {
	client, _, cleanup := newCacheClientForTest(t)
	defer cleanup()

	const total = 64
	ctx := context.Background()
	start := make(chan struct{})
	resultCh := make(chan bool, total)
	errCh := make(chan error, total)

	var wg sync.WaitGroup
	for i := range total {
		value := int64(i)

		wg.Go(func() {
			<-start

			ok, err := client.SetCounterNX(ctx, "counter:setnx", value, 5*time.Second)
			if err != nil {
				errCh <- err
				return
			}

			resultCh <- ok
		})
	}

	close(start)
	wg.Wait()
	close(resultCh)
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("期望 SetCounterNX 并发成功, 实际报错: %v", err)
		}
	}

	var successCount int
	for ok := range resultCh {
		if ok {
			successCount++
		}
	}

	if successCount != 1 {
		t.Fatalf("期望仅 1 个并发请求原子初始化计数器成功, 实际为 %d", successCount)
	}

	value, err := client.GetCounterValue(ctx, "counter:setnx")
	if err != nil {
		t.Fatalf("获取 SetCounterNX 最终值失败: %v", err)
	}
	if value < 0 || value >= total {
		t.Fatalf("期望 SetCounterNX 写入的计数值落在 [0, %d) 区间, 实际为 %d", total, value)
	}

	ttl, err := client.GetKeyTll(ctx, "counter:setnx")
	if err != nil {
		t.Fatalf("获取 SetCounterNX TTL 失败: %v", err)
	}
	if ttl <= 0 {
		t.Fatalf("期望 SetCounterNX 已设置 TTL, 实际为 %v", ttl)
	}
	if ttl > 5*time.Second {
		t.Fatalf("期望 SetCounterNX TTL 不超过写入值, 实际为 %v", ttl)
	}
}
