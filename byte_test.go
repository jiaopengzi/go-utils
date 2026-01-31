//
// FilePath    : go-utils\byte_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : PrintByte 函数测试
//

package utils

import (
	"testing"
)

func TestPrintByte_WithPerLine(t *testing.T) {
	t.Log("测试自定义 perLine=8")
	PrintByte("Hello, World!", 8)
}

func TestPrintByte_DefaultPerLine(t *testing.T) {
	t.Log("测试默认 perLine=16")
	PrintByte("Hello, World!")
}

func TestPrintByte_EmptyString(t *testing.T) {
	t.Log("测试空字符串")
	PrintByte("")
}

func TestPrintByte_SingleChar(t *testing.T) {
	t.Log("测试单个字符")
	PrintByte("A")
}

func TestPrintByte_ExactMultiple(t *testing.T) {
	t.Log("测试正好是 perLine 倍数的字符串 (8个字符, perLine=4)")
	PrintByte("ABCDEFGH", 4)
}

func TestPrintByte_ChineseString(t *testing.T) {
	t.Log("测试中文字符串 (UTF-8 多字节)")
	PrintByte("你好世界", 8)
}

func TestPrintByte_PerLineZero(t *testing.T) {
	t.Log("测试 perLine=0 时使用默认值")
	PrintByte("Test", 0)
}

func TestPrintByte_PerLineNegative(t *testing.T) {
	t.Log("测试 perLine=-1 时使用默认值")
	PrintByte("Test", -1)
}
