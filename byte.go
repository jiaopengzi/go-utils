//
// FilePath    : go-utils\byte.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : byte 相关工具
//

package utils

import "fmt"

// PrintByte 打印字符串的字节切片表示形式
// 例如: "abc" -> []byte{99,98,97}
// 第二个参数 perLine 指定每行打印的 byte 个数，可选，默认 16
func PrintByte(s string, perLine ...int) {
	n := 16
	if len(perLine) > 0 && perLine[0] > 0 {
		n = perLine[0]
	}

	b := []byte(s)

	fmt.Println("[]byte{")

	for i, v := range b {
		fmt.Printf("%d, ", v)

		if (i+1)%n == 0 {
			fmt.Print("\n")
		}
	}

	if len(b)%n != 0 {
		fmt.Println()
	}

	fmt.Println("}")
}
