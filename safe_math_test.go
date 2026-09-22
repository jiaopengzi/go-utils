//
// FilePath    : go-utils\safe_math_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 安全运算单元测试
//

package utils

import (
	"errors"
	"math"
	"testing"
)

// TestSafeAddInt64 覆盖正负溢出边界与正常值.
func TestSafeAddInt64(t *testing.T) {
	cases := []struct {
		name    string
		a, b    int64
		want    int64
		wantErr bool
	}{
		{"正常相加", 500, 100, 600, false},
		{"负数相加", 500, -600, -100, false},
		{"零", 0, 0, 0, false},
		{"正溢出", math.MaxInt64, 1, 0, true},
		{"正溢出_大数", 500, math.MaxInt64, 0, true},
		{"负溢出", math.MinInt64, -1, 0, true},
		{"边界不溢出", math.MaxInt64 - 1, 1, math.MaxInt64, false},
		{"边界不溢出_负", math.MinInt64 + 1, -1, math.MinInt64, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := SafeAddInt64(c.a, c.b)
			if c.wantErr {
				if !errors.Is(err, ErrIntOverflow) {
					t.Fatalf("SafeAddInt64(%d, %d) err = %v, want ErrIntOverflow", c.a, c.b, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("SafeAddInt64(%d, %d) unexpected err: %v", c.a, c.b, err)
			}

			if got != c.want {
				t.Fatalf("SafeAddInt64(%d, %d) = %d, want %d", c.a, c.b, got, c.want)
			}
		})
	}
}

// TestSafeMulInt64 覆盖乘法溢出边界与正常值.
func TestSafeMulInt64(t *testing.T) {
	cases := []struct {
		name    string
		a, b    int64
		want    int64
		wantErr bool
	}{
		{"正常相乘", 100, 3, 300, false},
		{"含零", 123, 0, 0, false},
		{"负数", -5, 7, -35, false},
		{"溢出", math.MaxInt64, 2, 0, true},
		{"溢出_回绕为正风险值", 4294967296, 4294967296, 0, true},
		{"溢出_MinInt64乘-1", math.MinInt64, -1, 0, true},
		{"边界不溢出", math.MaxInt64 / 2, 2, math.MaxInt64 - 1, false},
		{"边界不溢出_MinInt64乘1", math.MinInt64, 1, math.MinInt64, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := SafeMulInt64(c.a, c.b)
			if c.wantErr {
				if !errors.Is(err, ErrIntOverflow) {
					t.Fatalf("SafeMulInt64(%d, %d) err = %v, want ErrIntOverflow", c.a, c.b, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("SafeMulInt64(%d, %d) unexpected err: %v", c.a, c.b, err)
			}

			if got != c.want {
				t.Fatalf("SafeMulInt64(%d, %d) = %d, want %d", c.a, c.b, got, c.want)
			}
		})
	}
}
