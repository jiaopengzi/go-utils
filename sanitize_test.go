//
// FilePath    : go-utils\sanitize_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : HTML 消毒单元测试
//

package utils

import (
	"strings"
	"testing"
)

// TestSanitizeHTMLContent 覆盖常见 XSS 向量的剥离与正常内容的保留.
func TestSanitizeHTMLContent(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string // 期望消毒后包含(不含)的内容
		notWant string
	}{
		{
			name:  "script标签块_剥离",
			input: "前面<script>alert(1)</script>后面",
			want:  "前面后面",
		},
		{
			name:    "img_onerror事件_剥离属性",
			input:   `<img src=x onerror=alert(1)>`,
			notWant: "onerror",
		},
		{
			name:    "a_javascript协议_剥离",
			input:   `<a href="javascript:alert(2)">点我</a>`,
			notWant: "javascript:",
		},
		{
			name:  "iframe_整块剥离",
			input: `文字<iframe src="evil"></iframe>结尾`,
			want:  "文字结尾",
		},
		{
			name:  "未闭合script_剥离到结尾",
			input: `正常<script>alert(1)`,
			want:  "正常",
		},
		{
			name:    "data伪协议_剥离",
			input:   `<a href="data:text/html;base64,xxx">x</a>`,
			notWant: "data:",
		},
		{
			name:    "vbscript_剥离",
			input:   `<a href="vbscript:msgbox(1)">x</a>`,
			notWant: "vbscript:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeHTMLContent(tt.input)

			if tt.want != "" && !strings.Contains(got, tt.want) {
				t.Fatalf("SanitizeHTMLContent(%q) = %q, 期望包含 %q", tt.input, got, tt.want)
			}

			if tt.notWant != "" && strings.Contains(strings.ToLower(got), strings.ToLower(tt.notWant)) {
				t.Fatalf("SanitizeHTMLContent(%q) = %q, 不应包含 %q", tt.input, got, tt.notWant)
			}
		})
	}
}

// TestSanitizeHTMLContentKeepsNormal 普通文本与无害内容不受影响.
func TestSanitizeHTMLContentKeepsNormal(t *testing.T) {
	cases := []string{
		"这篇文章写的真好",
		"1 < 2 且 3 > 2 的数学题",
		"代码示例: if (a < b) { return }",
	}

	for _, c := range cases {
		if got := SanitizeHTMLContent(c); got != strings.TrimSpace(c) {
			t.Fatalf("SanitizeHTMLContent(%q) = %q, 普通文本不应被修改", c, got)
		}
	}

	if got := SanitizeHTMLContent(""); got != "" {
		t.Fatalf("空字符串应原样返回, got %q", got)
	}
}
