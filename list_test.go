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
	"reflect"
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

func TestReverseSlice_Empty(t *testing.T) {
	var s []int
	r := ReverseSlice(s)
	if len(r) != 0 {
		t.Fatalf("expected empty slice, got %v", r)
	}
}

func TestReverseSlice_Single(t *testing.T) {
	s := []string{"a"}
	r := ReverseSlice(s)
	if !reflect.DeepEqual(r, []string{"a"}) {
		t.Fatalf("expected %v, got %v", []string{"a"}, r)
	}
	if !reflect.DeepEqual(s, []string{"a"}) {
		t.Fatalf("original modified: %v", s)
	}
}

func TestReverseSlice_Multiple(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	r := ReverseSlice(s)
	expected := []int{5, 4, 3, 2, 1}
	if !reflect.DeepEqual(r, expected) {
		t.Fatalf("expected %v, got %v", expected, r)
	}
	if !reflect.DeepEqual(s, []int{1, 2, 3, 4, 5}) {
		t.Fatalf("original modified: %v", s)
	}
}
