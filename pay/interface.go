//
// FilePath    : go-utils\pay\interface.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 支付接口
//

package pay

import (
	"net/http"
	"time"
)

type Payer interface {
	// Prepay 支付接口
	//   - orderID: 订单ID
	//   - amount: 金额，单位为分
	//   - description: 商品描述
	//   - returnURL: 支付完成后跳转的页面
	//   - timeExpire: 订单失效时间, 格式为 ISO 8601
	// 返回值为支付链接或二维码的URL
	Prepay(orderID uint64, amount int64, description, returnURL string, timeExpire time.Time) (string, error)

	// GetNotifyPayment 获取支付结果通知接口, 包含验签和获取支付结果
	//  - request: HTTP请求对象
	// 返回值为是否成功处理通知，支付结果和错误信息
	GetNotifyPayment(request *http.Request) (bool, *PaymentResult, error)

	// ValidateNotifyPayment 验证支付结果通知接口
	//  - payment: 支付结果
	//  - orderID: 订单ID
	//  - amount: 金额，单位为分
	ValidateNotifyPayment(payment *PaymentResult, orderID uint64, amount int64) (bool, *PaymentResult, error)

	// QueryPayment 查询支付结果接口
	//  - orderID: 订单ID
	// 返回值为支付结果和错误信息
	QueryPayment(orderID uint64) (*PaymentResult, error)

	// CloseOrder 关闭订单接口
	// - orderID: 订单ID
	// 返回值为错误信息
	CloseOrder(orderID uint64) error

	// Refund 退款接口
	// - orderID: 订单ID
	// - RefundID: 退款ID
	// - amount: 订单总金额，单位为分
	// - refundAmount: 退款金额，单位为分(不能超过订单总金额)
	Refund(orderID, refundID uint64, amount, refundAmount int64, reason string) (*RefundResult, error)

	// GetNotifyRefund 应答退款结果通知接口, 包含验签和获取退款结果
	//  - request: HTTP请求对象
	// 返回值为是否成功处理通知，退款结果和错误信息
	GetNotifyRefund(request *http.Request) (bool, *RefundResult, error)

	// QueryRefund 查询退款结果接口
	//  - orderID: 订单ID
	//  - refundID: 退款ID
	// 返回值为退款结果和错误信息
	QueryRefund(orderID, refundID uint64) (*RefundResult, error)
}
