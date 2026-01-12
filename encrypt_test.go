//
// FilePath    : go-utils\encrypt_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 加密解密 单元测试
//

package utils

import (
	"testing"
)

var password = "123456"
var hashedPassword = "$2a$10$2PgrS4Kr.Jc2N1bpUPgh2u78kWJ9cG7NCCYQ4wHC7gKQWq.2S9ige"

func TestComparePasswords(t *testing.T) {

	// 测试匹配的情况
	isValid := ComparePasswords(hashedPassword, password)
	if !isValid {
		t.Errorf("Expected password to be valid, but it is invalid")
	}

	// 测试不匹配的情况
	invalidPassword := "wrong password"
	isValid = ComparePasswords(hashedPassword, invalidPassword)
	if isValid {
		t.Errorf("Expected password to be invalid, but it is valid")
	}
}
