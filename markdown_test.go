//
// FilePath    : go-utils\markdown_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : markdown 付费标签处理测试
//

package utils

import (
	"testing"
)

func TestReplaceMarkdownPayTagToEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		payType  MarkdownPayType
		expected string
	}{
		{
			name: "多行 pay-read 块",
			input: `<pay-read>
# 付费阅读 1
## 二级标题
</pay-read>`,
			payType:  MarkdownPayRead,
			expected: "<pay-read></pay-read>",
		},
		{
			name:     "单行 pay-read 块",
			input:    "<pay-read> 付费阅读2 </pay-read>",
			payType:  MarkdownPayRead,
			expected: "<pay-read></pay-read>",
		},
		{
			name:     "无付费标签",
			input:    "普通内容，无标签",
			payType:  MarkdownPayRead,
			expected: "普通内容，无标签",
		},
		{
			name: "混合内容含 pay-read",
			input: `前面内容
<pay-read>
中间内容
</pay-read>
后面内容`,
			payType:  MarkdownPayRead,
			expected: "前面内容\n<pay-read></pay-read>\n后面内容",
		},
		{
			name:     "多个 pay-read 标签",
			input:    `<pay-read>内容1</pay-read>普通中间内容<pay-read>内容2</pay-read>`,
			payType:  MarkdownPayRead,
			expected: "<pay-read></pay-read>普通中间内容<pay-read></pay-read>",
		},
		{
			name:     "包含标签的 pay-read 行内代码块 开始标签",
			input:    "<pay-read>`<pay-read>`</pay-read>",
			payType:  MarkdownPayRead,
			expected: "<pay-read></pay-read>",
		},
		{
			name:     "包含标签的 pay-read 行内代码块 结束标签",
			input:    "<pay-read>`</pay-read>`</pay-read>",
			payType:  MarkdownPayRead,
			expected: "<pay-read></pay-read>",
		},
		{
			name:     "包含标签的 pay-read 多行代码块 开始标签",
			input:    "<pay-read>\n```\n<pay-read>\n```\n</pay-read>",
			payType:  MarkdownPayRead,
			expected: "<pay-read></pay-read>",
		},
		{
			name:     "包含标签的 pay-read 多行代码块 结束标签",
			input:    "<pay-read>\n```\n</pay-read>\n```\n</pay-read>",
			payType:  MarkdownPayRead,
			expected: "<pay-read></pay-read>",
		},
		{
			name:     "包含代码块的 pay-read",
			input:    "<pay-read>\n```\n<pay-read>\n这是一个示例代码\n</pay-read>\n```\n</pay-read>",
			payType:  MarkdownPayRead,
			expected: "<pay-read></pay-read>",
		},
		{
			name:     "嵌套 pay-read",
			input:    "<pay-read>嵌套外层<pay-read>嵌套内层</pay-read>嵌套外层</pay-read>",
			payType:  MarkdownPayRead,
			expected: "<pay-read></pay-read>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceMarkdownPayTagToEmpty(tt.input, tt.payType)
			if got != tt.expected {
				t.Errorf("ReplaceMarkdownPayTagToEmpty() = %q, want %q", got, tt.expected)
			}
		})
	}
}
func TestReplaceMarkdownPayTagsToEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "多行 pay-read 块",
			input: `<pay-read>
# 付费阅读 1
## 二级标题
</pay-read>`,
			expected: "<pay-read></pay-read>",
		},
		{
			name:     "单行 pay-read 块",
			input:    "<pay-read> 付费阅读2 </pay-read>",
			expected: "<pay-read></pay-read>",
		},
		{
			name:     "无付费标签",
			input:    "普通内容，无标签",
			expected: "普通内容，无标签",
		},
		{
			name: "混合内容含 pay-read",
			input: `前面内容
<pay-read>
中间内容
</pay-read>
后面内容`,
			expected: "前面内容\n<pay-read></pay-read>\n后面内容",
		},
		{
			name:     "多个 pay-read 标签",
			input:    `<pay-read>内容1</pay-read>普通中间内容<pay-read>内容2</pay-read>`,
			expected: "<pay-read></pay-read>普通中间内容<pay-read></pay-read>",
		},
		{
			name:     "包含标签的 pay-read 行内代码块 开始标签",
			input:    "<pay-read>`<pay-read>`</pay-read>",
			expected: "<pay-read></pay-read>",
		},
		{
			name:     "包含标签的 pay-read 行内代码块 结束标签",
			input:    "<pay-read>`</pay-read>`</pay-read>",
			expected: "<pay-read></pay-read>",
		},
		{
			name:     "包含标签的 pay-read 多行代码块 开始标签",
			input:    "<pay-read>\n```\n<pay-read>\n```\n</pay-read>",
			expected: "<pay-read></pay-read>",
		},
		{
			name:     "包含标签的 pay-read 多行代码块 结束标签",
			input:    "<pay-read>\n```\n</pay-read>\n```\n</pay-read>",
			expected: "<pay-read></pay-read>",
		},
		{
			name:     "包含代码块的 pay-read",
			input:    "<pay-read>\n```\n<pay-read>\n这是一个示例代码\n</pay-read>\n```\n</pay-read>",
			expected: "<pay-read></pay-read>",
		},
		{
			name:     "嵌套 pay-read",
			input:    "<pay-read>嵌套外层<pay-read>嵌套内层</pay-read>嵌套外层</pay-read>",
			expected: "<pay-read></pay-read>",
		},

		{

			name: "多行 pay-download 块",
			input: `<pay-download>
# 付费下载 1
## 二级标题
</pay-download>`,
			expected: "<pay-download></pay-download>",
		},
		{
			name:     "单行 pay-download 块",
			input:    "<pay-download> 付费下载2 </pay-download>",
			expected: "<pay-download></pay-download>",
		},
		{
			name:     "无付费标签",
			input:    "普通内容，无标签",
			expected: "普通内容，无标签",
		},
		{
			name: "混合内容含 pay-download",
			input: `前面内容
<pay-download>
中间内容
</pay-download>
后面内容`,
			expected: "前面内容\n<pay-download></pay-download>\n后面内容",
		},
		{
			name:     "多个 pay-download 标签",
			input:    `<pay-download>内容1</pay-download>普通中间内容<pay-download>内容2</pay-download>`,
			expected: "<pay-download></pay-download>普通中间内容<pay-download></pay-download>",
		},
		{
			name:     "包含标签的 pay-download 行内代码块 开始标签",
			input:    "<pay-download>`<pay-download>`</pay-download>",
			expected: "<pay-download></pay-download>",
		},
		{
			name:     "包含标签的 pay-download 行内代码块 结束标签",
			input:    "<pay-download>`</pay-download>`</pay-download>",
			expected: "<pay-download></pay-download>",
		},
		{
			name:     "包含标签的 pay-download 多行代码块 开始标签",
			input:    "<pay-download>\n```\n<pay-download>\n```\n</pay-download>",
			expected: "<pay-download></pay-download>",
		},
		{
			name:     "包含标签的 pay-download 多行代码块 结束标签",
			input:    "<pay-download>\n```\n</pay-download>\n```\n</pay-download>",
			expected: "<pay-download></pay-download>",
		},
		{
			name:     "包含代码块的 pay-download",
			input:    "<pay-download>\n```\n<pay-download>\n这是一个示例代码\n</pay-download>\n```\n</pay-download>",
			expected: "<pay-download></pay-download>",
		},
		{
			name:     "嵌套 pay-download",
			input:    "<pay-download>嵌套外层<pay-download>嵌套内层</pay-download>嵌套外层</pay-download>",
			expected: "<pay-download></pay-download>",
		},
		{
			name:     "pay-read pay-download 混合",
			input:    `<pay-read>内容1</pay-read>普通中间内容<pay-download>内容2</pay-download>`,
			expected: "<pay-read></pay-read>普通中间内容<pay-download></pay-download>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceMarkdownPayTagsToEmpty(tt.input)
			if got != tt.expected {
				t.Errorf("ReplaceMarkdownPayTagsToEmpty() = %q, want %q", got, tt.expected)
			}
		})
	}
}
