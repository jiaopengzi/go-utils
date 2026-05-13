//
// FilePath    : go-utils\sonyflake_generator_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 通用雪花 ID 生成器测试
//

package utils

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/sony/sonyflake"
)

// newSonyflakeGeneratorForTest 创建测试使用的雪花生成器.
func newSonyflakeGeneratorForTest() *SonyflakeGenerator {
	return NewSonyflakeGenerator(func() (sonyflake.Settings, error) {
		return sonyflake.Settings{
			StartTime: time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC),
			MachineID: func() (uint16, error) {
				return 1, nil
			},
		}, nil
	})
}

// TestSonyflakeGeneratorConcurrentUnique 验证并发生成雪花 ID 时不会出现重复值.
func TestSonyflakeGeneratorConcurrentUnique(t *testing.T) {
	generator := newSonyflakeGeneratorForTest()

	const total = 256
	start := make(chan struct{})
	ids := make(chan uint64, total)
	errCh := make(chan error, total)

	var wg sync.WaitGroup
	for index := 0; index < total; index++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start

			id, err := generator.NextID()
			if err != nil {
				errCh <- err
				return
			}

			ids <- id
		}()
	}

	close(start)
	wg.Wait()
	close(ids)
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("期望并发生成 ID 成功, 实际报错: %v", err)
		}
	}

	seen := make(map[uint64]struct{}, total)
	for id := range ids {
		if _, exists := seen[id]; exists {
			t.Fatalf("期望并发生成的雪花 ID 全部唯一, 实际发现重复 ID: %d", id)
		}
		seen[id] = struct{}{}
	}

	if len(seen) != total {
		t.Fatalf("期望生成 %d 个唯一 ID, 实际为 %d", total, len(seen))
	}
}

// TestSonyflakeGeneratorProviderError 验证初始化失败时会返回提供方错误.
func TestSonyflakeGeneratorProviderError(t *testing.T) {
	expectedErr := errors.New("provider failed")
	generator := NewSonyflakeGenerator(func() (sonyflake.Settings, error) {
		return sonyflake.Settings{}, expectedErr
	})

	_, err := generator.NextID()
	if !errors.Is(err, expectedErr) {
		t.Fatalf("期望返回 provider 错误 %v, 实际为 %v", expectedErr, err)
	}
}
