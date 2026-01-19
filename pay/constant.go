//
// FilePath    : go-utils\pay\constant.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 支付常量
//

package pay

// PayType 支付类型
type PayType string

// 支付类型常量
const (
	PayTypeZero   PayType = "zero"       // 零元支付
	PayTypeAlipay PayType = "alipay"     // 支付宝
	PayTypeWechat PayType = "wechat_pay" // 微信支付
)

// TradeState 支付状态
type TradeState string

// 支付状态常量
const (
	TradeStateUnpaid   TradeState = "unpaid"   // 1 未支付
	TradeStatePaid     TradeState = "paid"     // 2 已支付
	TradeStateRefunded TradeState = "refunded" // 3 转入退款
	TradeStateClosed   TradeState = "closed"   // 4 已关闭
)

// RefundStatus 退款状态
type RefundStatus string

// 退款状态常量
const (
	RefundStatusPending    RefundStatus = "pending"    // 待处理
	RefundStatusProcessing RefundStatus = "processing" // 退款处理中
	RefundStatusSuccess    RefundStatus = "success"    // 退款成功
	RefundStatusClosed     RefundStatus = "closed"     // 退款关闭
	RefundStatusFailed     RefundStatus = "failed"     // 退款失败
)
