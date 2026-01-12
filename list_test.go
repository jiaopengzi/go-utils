//
// FilePath    : go-utils\list_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : Difference 函数测试
//

package utils

import (
	"fmt"
	"testing"
)

func TestDifference_Basic(t *testing.T) {
	listA := []string{"a", "b", "c", "d"}
	listB := []string{"b", "d"}
	expected := []string{"a", "c"}

	result := Difference(listA, listB)
	if !IsSlicesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestDifference_EmptyListA(t *testing.T) {
	listA := []string{}
	listB := []string{"a", "b"}
	expected := []string{}

	result := Difference(listA, listB)
	if !IsSlicesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestDifference_EmptyListB(t *testing.T) {
	listA := []string{"a", "b"}
	listB := []string{}
	expected := []string{"a", "b"}

	result := Difference(listA, listB)
	if !IsSlicesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestDifference_AllOverlap(t *testing.T) {
	listA := []string{"a", "b"}
	listB := []string{"a", "b"}
	expected := []string{}

	result := Difference(listA, listB)
	if !IsSlicesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestDifference_NoOverlap(t *testing.T) {
	listA := []string{"a", "b"}
	listB := []string{"c", "d"}
	expected := []string{"a", "b"}

	result := Difference(listA, listB)
	if !IsSlicesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestDifference_LargeInput(t *testing.T) {
	n := 10000
	listA := make([]string, n)
	listB := make([]string, n/2)
	// 确保 listA 和 listB 的元素完全不同
	for i := range n {
		listA[i] = fmt.Sprintf("test%d", i)
		if i < n/2 {
			listB[i] = fmt.Sprintf("test%d", i)
		}
	}
	result := Difference(listA, listB)
	if len(result) == n {
		t.Errorf("Expected length %d, got %d", n, len(result))
	}
}

func TestDifference_ListBShorterThanCPUs(t *testing.T) {
	listA := []string{"a", "b", "c"}
	listB := []string{"b"}
	expected := []string{"a", "c"}

	result := Difference(listA, listB)
	if !IsSlicesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
