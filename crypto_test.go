//
// FilePath    : go-utils\crypto_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 加密解密 单元测试
//

package utils

import (
	"strings"
	"testing"
)

var password = "123456"

func TestComparePasswords(t *testing.T) {
	hashedPassword, err := GenerateHashedPassword(password, 10)
	if err != nil {
		t.Fatalf("GenerateHashedPassword() error = %v", err)
	}

	if !strings.HasPrefix(hashedPassword, pbkdf2HashScheme+"$") {
		t.Fatalf("GenerateHashedPassword() returned unexpected format: %s", hashedPassword)
	}

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
