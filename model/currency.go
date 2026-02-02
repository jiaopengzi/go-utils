//
// FilePath    : go-utils\model\currency.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 货币类型
//

package model

import (
	"encoding/json"
	"fmt"
)

type Currency int // 订单货币类型

// 定义订单货币类型常量
const (
	CurrencyCNY Currency = iota + 1 // 1 人民币
	CurrencyUSD                     // 2 美元
	CurrencyEUR                     // 3 欧元
	CurrencyGBP                     // 4 英镑
	CurrencyHKD                     // 5 港币
	CurrencyTWD                     // 6 台币
	CurrencySGD                     // 7 新加坡元
	CurrencyRUB                     // 8 卢布
)

// String 实现 fmt.Stringer 接口, 返回货币的符号标签.
// 返回值 string, 货币符号标签.
func (c Currency) String() string {
	switch c {
	case CurrencyCNY:
		return "￥"
	case CurrencyUSD:
		return "$"
	case CurrencyEUR:
		return "€"
	case CurrencyGBP:
		return "£"
	case CurrencyHKD:
		return "HK$"
	case CurrencyTWD:
		return "NT$"
	case CurrencySGD:
		return "S$"
	case CurrencyRUB:
		return "₽"
	default:
		return "未知货币"
	}
}

// MarshalJSON 实现 json.Marshaler 接口, 将货币符号列化为包含 value 和 label 的 JSON 对象.
// 返回值 []byte, JSON 编码后的字节切片; error, 出错时非 nil.
func (c Currency) MarshalJSON() ([]byte, error) {
	return json.Marshal(IntEnumJSON{
		Value: int(c),
		Label: c.String(),
	})
}

// AmountFenToYuan 金额从分转换为元, 保留两位小数
func (c Currency) AmountFenToYuan(amountFen int64) string {
	amountYuan := float64(amountFen) / 100.0

	return fmt.Sprintf("%s%.2f", c.String(), amountYuan)
}
