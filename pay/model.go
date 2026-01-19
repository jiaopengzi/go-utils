//
// FilePath    : go-utils\pay\model.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 支付相关数据结构
//

package pay

// PaymentResult 支付结果
type PaymentResult struct {
	PayType       PayType    `json:"pay_type"`
	OrderID       uint64     `json:"order_id"`
	TotalAmount   int64      `json:"total_amount"`
	TransactionID string     `json:"transaction_id"`
	TradeState    TradeState `json:"trade_state"`
	TradeType     string     `json:"trade_type"`
	AppID         string     `json:"app_id,omitempty"`    // 仅在通知中返回
	MchID         string     `json:"mch_id,omitempty"`    // 仅微信支付需要
	SellerID      string     `json:"seller_id,omitempty"` // 仅支付宝需要
}

// RefundResult 退款结果
type RefundResult struct {
	PayType             PayType      `json:"pay_type"`
	RefundID            uint64       `json:"refund_id"`
	OrderID             uint64       `json:"order_id"`
	TransactionID       string       `json:"transaction_id"`
	RefundTransactionID string       `json:"refund_transaction_id"`
	TotalAmount         int64        `json:"total_amount"`
	RefundAmount        int64        `json:"refund_amount"`
	Reason              string       `json:"reason"`
	Status              RefundStatus `json:"status"`
	AppID               string       `json:"app_id,omitempty"`    // 仅在通知中返回
	MchID               string       `json:"mch_id,omitempty"`    // 仅微信支付需要
	SellerID            string       `json:"seller_id,omitempty"` // 仅支付宝需要
}
