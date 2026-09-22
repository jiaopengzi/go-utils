//
// FilePath    : go-utils\safe_math.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 有符号整数安全运算, 防止余额/累计值加减溢出回绕
//

package utils

import (
	"errors"
	"math"
)

// ErrIntOverflow 整数运算溢出错误
var ErrIntOverflow = errors.New("int64 arithmetic overflow")

// SafeAddInt64 有符号整数安全加法, 溢出时返回 ErrIntOverflow 而不是回绕.
func SafeAddInt64(a, b int64) (int64, error) {
	if b > 0 && a > math.MaxInt64-b {
		return 0, ErrIntOverflow
	}

	if b < 0 && a < math.MinInt64-b {
		return 0, ErrIntOverflow
	}

	return a + b, nil
}

// SafeMulInt64 有符号整数安全乘法, 溢出时返回 ErrIntOverflow 而不是回绕.
func SafeMulInt64(a, b int64) (int64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}

	// MinInt64 * -1 会回绕为 MinInt64 且除法校验同样回绕, 需要特判
	if a == math.MinInt64 || b == math.MinInt64 {
		if a == -1 || b == -1 {
			return 0, ErrIntOverflow
		}

		if a != 1 && b != 1 {
			return 0, ErrIntOverflow
		}
	}

	result := a * b

	if result/b != a {
		return 0, ErrIntOverflow
	}

	return result, nil
}
