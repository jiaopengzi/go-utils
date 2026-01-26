//
// FilePath    : go-utils\model\currency.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 货币类型
//

package model

import "fmt"

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

// 货币符号映射
var CurrencySymbols = map[Currency]string{
	CurrencyCNY: "¥",
	CurrencyUSD: "$",
	CurrencyEUR: "€",
	CurrencyGBP: "£",
	CurrencyHKD: "HK$",
	CurrencyTWD: "NT$",
	CurrencySGD: "S$",
	CurrencyRUB: "₽",
}

// AmountFenToYuan 金额从分转换为元, 保留两位小数
func (c Currency) AmountFenToYuan(amountFen int64) string {
	amountYuan := float64(amountFen) / 100.0

	symbol, exists := CurrencySymbols[c]
	if !exists {
		symbol = ""
	}

	return fmt.Sprintf("%s%.2f", symbol, amountYuan)
}
