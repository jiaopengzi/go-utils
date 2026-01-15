//
// FilePath    : go-utils\hash_signature_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 单元测试 - 签名函数
//

package utils

import (
	"testing"
)

func TestSignData(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		secret string
		opts   []SignOptionFunc
	}{
		{
			name:   "default SHA256 algorithm",
			data:   "hello world",
			secret: "mysecret",
			opts:   nil,
		},
		{
			name:   "empty data",
			data:   "",
			secret: "mysecret",
			opts:   nil,
		},
		{
			name:   "empty secret",
			data:   "hello world",
			secret: "",
			opts:   nil,
		},
		{
			name:   "with SHA256 option",
			data:   "test data",
			secret: "testsecret",
			opts:   []SignOptionFunc{WithAlgorithm(SHA256)},
		},
		{
			name:   "unicode data",
			data:   "你好世界",
			secret: "密钥",
			opts:   nil,
		},
		{
			name:   "long data",
			data:   "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
			secret: "longsecretkey123456789",
			opts:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SignData(tt.data, tt.secret, tt.opts...)

			// Verify signature is not empty
			if result == "" {
				t.Error("SignData returned empty signature")
			}

			// Verify signature is deterministic (same input produces same output)
			result2 := SignData(tt.data, tt.secret, tt.opts...)
			if result != result2 {
				t.Errorf("SignData is not deterministic: got %s and %s for same input", result, result2)
			}
		})
	}
}

func TestSignData_DifferentInputsProduceDifferentSignatures(t *testing.T) {
	secret := "mysecret"

	sig1 := SignData("data1", secret)
	sig2 := SignData("data2", secret)

	if sig1 == sig2 {
		t.Error("Different data should produce different signatures")
	}
}

func TestSignData_DifferentSecretsProduceDifferentSignatures(t *testing.T) {
	data := "same data"

	sig1 := SignData(data, "secret1")
	sig2 := SignData(data, "secret2")

	if sig1 == sig2 {
		t.Error("Different secrets should produce different signatures")
	}
}

func TestSignData_URLSafeOutput(t *testing.T) {
	result := SignData("test data", "secret")

	// Check that output doesn't contain URL-unsafe characters or padding
	for _, c := range result {
		if c == '+' || c == '/' || c == '=' {
			t.Errorf("Signature contains non-URL-safe character: %c", c)
		}
	}
}
