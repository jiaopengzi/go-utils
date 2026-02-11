//
// FilePath    : go-utils\model\transaction_flow.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 交易流水类型
//

package model

import (
	"encoding/json"
)

// 交易类型
type TransactionType int

// 定义交易类型常量
const (
	TransactionTypeReward   TransactionType = iota + 1 // 1 奖励
	TransactionTypeRecharge                            // 2 充值
	TransactionTypeConsume                             // 3 消费
	TransactionTypeRefund                              // 4 退款
	TransactionTypePenalty                             // 5 处罚
)

// String 实现 fmt.Stringer 接口, 返回流水表状态码的字符串表示.
// 返回值 string, 流水表状态码的标签.
func (t TransactionType) String() string {
	switch t {
	case TransactionTypeReward:
		return "奖励"
	case TransactionTypeRecharge:
		return "充值"
	case TransactionTypeConsume:
		return "消费"
	case TransactionTypeRefund:
		return "退款"
	case TransactionTypePenalty:
		return "处罚"
	default:
		return "未知类型"
	}
}

// MarshalJSON 实现 json.Marshaler 接口, 将流水表状态码序列化为包含 value 和 label 的 JSON 对象.
// 返回值 []byte, JSON 编码后的字节切片; error, 出错时非 nil.
func (t TransactionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(IntEnumJSON{
		Value: int(t),
		Label: t.String(),
	})
}

// UnmarshalJSON 实现 json.Unmarshaler 接口, 支持从整数或对象反序列化.
func (t *TransactionType) UnmarshalJSON(data []byte) error {
	var intVal int
	if err := json.Unmarshal(data, &intVal); err == nil {
		*t = TransactionType(intVal)
		return nil
	}

	var obj IntEnumJSON
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}

	*t = TransactionType(obj.Value)

	return nil
}

// AmountSign 返回交易类型对应的金额符号.
func (t TransactionType) AmountSign() int64 {
	switch t {
	case TransactionTypeReward:
		return 1
	case TransactionTypeRecharge:
		return 1
	case TransactionTypeConsume:
		return -1
	case TransactionTypeRefund:
		return -1
	case TransactionTypePenalty:
		return -1
	default:
		return 0
	}
}

// Amount 根据交易类型调整金额符号
func (t TransactionType) Amount(amount int64) int64 {
	// 先对 amount 取绝对值
	if amount < 0 {
		amount = -amount
	}

	// 根据交易类型调整金额符号
	amount *= t.AmountSign()

	return amount
}
