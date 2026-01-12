//
// FilePath    : go-utils\regex.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 正则表达式校验
//

package utils

import "regexp"

// CheckRegex 通过 regex 正则表达式校验 str 字符串是否匹配
func CheckRegex(regex, str string) bool {
	ok, err := regexp.MatchString(regex, str)
	if err != nil {
		return false
	}

	return ok
}
