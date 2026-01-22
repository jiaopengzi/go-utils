//
// FilePath    : go-utils\err.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 自定义错误
//

package utils

import (
	"errors"
	"net"
	"os"
)

type JpzError string

const (
	ErrNotEmpty               = JpzError("not_empty.")                      // 不能为空
	ErrSlugTooLong            = JpzError("slug_too_long.")                  // slug 过长
	ErrRedisNoAuth            = JpzError("NOAUTH Authentication required.") // redis 未授权
	ErrOrderNotOwn            = JpzError("order is not own.")               // 订单不属于当前用户
	ErrOrderCheckoutExpired   = JpzError("order checkout is expired.")      // 订单结算信息已过期
	ErrRefundWeChatNotEnough  = JpzError("refund wechat not enough.")       // 微信退款余额不足
	ErrTokenInvalidClaims     = JpzError("invalid_token_claims.")           // token 声明无效
	ErrTokenInvalid           = JpzError("token_is_invalid.")               // token 无效
	ErrTokenInvalidType       = JpzError("invalid_token_type.")             // token 类型无效
	ErrTokenMissingUserID     = JpzError("token_missing_user_id.")          // token 缺少用户ID
	ErrTokenMissingJwi        = JpzError("token_missing_jwi.")              // token 缺少 jwi
	ErrUpdateRowsAffectedZero = JpzError("update_rows_affected_zero.")      // 更新影响行数为0
	ErrTimeout                = JpzError("timeout.")                        // 超时
	ErrInvalidSignature       = JpzError("invalid_signature.")              // 签名无效
	ErrTimestampDiffExceeded  = JpzError("timestamp_difference_exceeded.")  // 时间戳差异超出允许范围
	ErrRequestIDNotFound      = JpzError("request_id_not_found.")           // 请求ID未找到
	ErrDistributedLockFailed  = JpzError("distributed_lock_failed.")        // 分布式锁获取失败
)

// Error 实现 error 接口 Error 方法
func (e JpzError) Error() string { return string(e) }

// IsTimeoutError 判断是否是超时错误
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	// 检查 os.ErrDeadlineExceeded
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}

	// 检查 net.Error 的 Timeout() 方法
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	// 是否包含 JpzError 超时错误
	if errors.Is(err, ErrTimeout) {
		return true
	}

	return false
}
