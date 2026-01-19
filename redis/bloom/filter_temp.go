//
// FilePath    : go-utils\redis\bloom\filter_temp.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 内存临时使用布隆过滤器
//

package bloom

import (
	"github.com/bits-and-blooms/bloom/v3"
)

// FilterTemp 临时布隆过滤器结构体
type FilterTemp struct {
	Filter *bloom.BloomFilter // 布隆过滤器
	N      uint               // 预计元素数量
	FP     float64            // 误判率
}

// NewFilter 创建布隆过滤器实例 用于判断元素是否存在
//   - n: 预计元素数量个数
//   - fp: 误判率 0-1 之间，越小误判率越低，但是占用内存越大 一般取 0.01
func NewFilterTemp(n uint, fp float64) *FilterTemp {
	bf := bloom.NewWithEstimates(n, fp)

	return &FilterTemp{
		Filter: bf,
		N:      n,
		FP:     fp,
	}
}

// Add 添加元素到布隆过滤器
func (b *FilterTemp) Add(item string) {
	// 判断 item 是否为空
	if item != "" {
		b.Filter.Add([]byte(item))
	}
}

// AddGroup 批量添加元素到布隆过滤器
func (b *FilterTemp) MAdd(items []string) {
	for _, item := range items {
		// 判断 item 是否为空
		if item != "" {
			b.Filter.Add([]byte(item))
		}
	}
}

// Test 判断元素是否存在(可能会误判,但如果不存在则一定不存在)
func (b *FilterTemp) Test(testStr string) bool {
	return b.Filter.Test([]byte(testStr))
}

// ClearAll 清空过滤器
func (b *FilterTemp) ClearAll() error {
	b.Filter.ClearAll()
	return nil
}
