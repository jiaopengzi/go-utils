//
// FilePath    : go-utils\err.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 自定义错误
//

package utils

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
)

// Error 实现 error 接口 Error 方法
func (e JpzError) Error() string { return string(e) }
