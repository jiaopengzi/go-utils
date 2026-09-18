//
// FilePath    : go-utils\url_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : URL 工具单测
//

package utils

import (
	"testing"
)

func TestIsImageURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"example.com/image.jpg", false},
		{"example.com/image.jpg", false},
		{"http://example.com/image.jpg", true},
		{"https://example.com/image.jpeg", true},
		{"http://example.com/image.png", true},
		{"https://example.com/image.gif", true},
		{"http://example.com/image.bmp", true},
		{"https://example.com/image.tiff", true},
		{"http://example.com/image.txt", false},
		{"https://example.com/image", false},
		{"", false},
		{"http://example.com/image.JPG", true}, // Test case-insensitivity
		{"https://example.com/image.JPEG", true},
	}

	for _, test := range tests {
		result := IsImageURL(test.url)
		if result != test.expected {
			t.Errorf("IsImageURL(%s) = %v; expected %v", test.url, result, test.expected)
		}
	}
}

// TestEncodeURL
// @Description: 测试 EncodeURL 函数
// @Parameter    t: testing.T
func TestEncodeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://example.com", "https%3A%2F%2Fexample.com"},
		{"https://example.com/path?query=1", "https%3A%2F%2Fexample.com%2Fpath%3Fquery%3D1"},
		{"https://example.com/中文", "https%3A%2F%2Fexample.com%2F%E4%B8%AD%E6%96%87"},
		{"", ""},
	}

	for _, test := range tests {
		result := EncodeURL(test.input)
		if result != test.expected {
			t.Errorf("EncodeURL(%s) = %v; expected %v", test.input, result, test.expected)
		}
	}
}

// TestEncodeURLComponent
// @Description: 测试 EncodeURLComponent 函数 (JS encodeURIComponent 兼容编码)
// @Parameter    t: testing.T
func TestEncodeURLComponent(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"abcXYZ012-_.!~*'()", "abcXYZ012-_.!~*'()"},
		{"中文 空格标签", "%E4%B8%AD%E6%96%87%20%E7%A9%BA%E6%A0%BC%E6%A0%87%E7%AD%BE"},
		{"konge biaoqian", "konge%20biaoqian"},
		{"C++", "C%2B%2B"},
		{"100%", "100%25"},
		{"a&b=c?d/e", "a%26b%3Dc%3Fd%2Fe"},
	}

	for _, test := range tests {
		result := EncodeURLComponent(test.input)
		if result != test.expected {
			t.Errorf("EncodeURLComponent(%s) = %v; expected %v", test.input, result, test.expected)
		}
	}
}

// TestGenerateSlugByName
// @Description: 测试 GenerateSlugByName 函数 (bf-260919-01: 空格生成 %20 而非 "+")
// @Parameter    t: testing.T
func TestGenerateSlugByName(t *testing.T) {
	//	空名称返回错误
	if _, err := GenerateSlugByName(""); err == nil {
		t.Errorf("GenerateSlugByName(empty) expected error, got nil")
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"中文 空格标签", "%E4%B8%AD%E6%96%87%20%E7%A9%BA%E6%A0%BC%E6%A0%87%E7%AD%BE"},
		{"konge biaoqian", "konge%20biaoqian"},
		{"Tabular Editor", "Tabular%20Editor"},
	}

	for _, test := range tests {
		result, err := GenerateSlugByName(test.input)
		if err != nil {
			t.Errorf("GenerateSlugByName(%s) unexpected error: %v", test.input, err)
			continue
		}
		if result != test.expected {
			t.Errorf("GenerateSlugByName(%s) = %v; expected %v", test.input, result, test.expected)
		}
	}
}
