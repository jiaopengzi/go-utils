//
// FilePath    : go-utils\list.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 列表相关工具
//

package utils

import (
	"runtime"
	"sync"
)

// Difference 并行计算两个字符串切片的差集，返回 listA 中所有不在 listB 中的元素。
func Difference(listA, listB []string) []string {
	// 如果 listB 为空，直接返回 listA 的副本
	if len(listB) == 0 {
		return listA
	}

	if len(listB) < 2000 {
		return differenceSmall(listA, listB)
	}

	return differenceBig(listA, listB)
}

// differenceSmall 计算两个字符串切片的差集，返回 listA 中所有不在 listB 中的元素(适用于 listB 较小的情况)。
func differenceSmall(listA, listB []string) []string {
	// 将listB转换为map提高查找效率
	bMap := make(map[string]struct{}, len(listB))
	for _, email := range listB {
		bMap[email] = struct{}{}
	}

	var diff []string

	for _, email := range listA {
		if _, exists := bMap[email]; !exists {
			diff = append(diff, email)
		}
	}

	return diff
}

// differenceBig 并行计算两个字符串切片的差集，返回 listA 中所有不在 listB 中的元素。
// 该函数通过多协程并发构建 listB 的查找表（sync.Map），保证并发安全，适用于 listB 较大时的高性能场景。
func differenceBig(listA, listB []string) []string {
	var (
		bMap sync.Map
		wg   sync.WaitGroup
		diff []string
	)

	numWorkers := min(runtime.NumCPU(), len(listB))
	chunkSize := max(len(listB)/numWorkers, 1)

	wg.Add(numWorkers)

	for i := range numWorkers {
		go func(workerID int) {
			defer wg.Done()

			start := workerID * chunkSize

			end := start + chunkSize
			if workerID == numWorkers-1 {
				end = len(listB)
			}

			for _, email := range listB[start:end] {
				bMap.Store(email, struct{}{})
			}
		}(i)
	}

	wg.Wait()

	for _, email := range listA {
		if _, exists := bMap.Load(email); !exists {
			diff = append(diff, email)
		}
	}

	return diff
}

// RemoveDuplicateElement 移除字符串切片中的重复元素,返回一个只包含唯一元素的新切片。
func RemoveDuplicateElement(list []string) []string {
	seen := make(map[string]struct{})      // 用于记录已见过的元素
	result := make([]string, 0, len(list)) // 结果切片，预分配内存

	for _, item := range list {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}

			result = append(result, item)
		}
	}

	return result
}
