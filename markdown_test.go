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
		{
			name:     "fenced code block 内的 pay-video 保留, 外部 pay-video 清空",
			input:    "```\n<pay-video>\n除视频外的其他隐藏内容, 如附件下载等; 若没有则将标签设置为一行\n<pay-membership></pay-membership>\n</pay-video>\n```\n\n<pay-video>\n真实付费视频内容\n</pay-video>",
			payType:  MarkdownPayVideo,
			expected: "```\n<pay-video>\n除视频外的其他隐藏内容, 如附件下载等; 若没有则将标签设置为一行\n<pay-membership></pay-membership>\n</pay-video>\n```\n\n<pay-video has-material></pay-video>",
		},
		{
			name:     "pay-video 无内容时不携带 has-material",
			input:    "<pay-video></pay-video>",
			payType:  MarkdownPayVideo,
			expected: "<pay-video></pay-video>",
		},
		{
			name:     "pay-video 仅空白内容时不携带 has-material",
			input:    "<pay-video>   \n  \n  </pay-video>",
			payType:  MarkdownPayVideo,
			expected: "<pay-video></pay-video>",
		},
		{
			name:     "pay-video 有内容时携带 has-material",
			input:    "<pay-video>素材链接</pay-video>",
			payType:  MarkdownPayVideo,
			expected: "<pay-video has-material></pay-video>",
		},
		{
			name:     "pay-read 有内容时不携带 has-material",
			input:    "<pay-read>付费阅读内容</pay-read>",
			payType:  MarkdownPayRead,
			expected: "<pay-read></pay-read>",
		},
		{
			name:     "行内代码包裹的 <pay-read> 不会被剩离 (多个并列)",
			input:    "- 项目自定义标签（例如 `<pay-read>`、`<power-bi>`、`<wechat-captcha>`）\n\n后面的内容\n",
			payType:  MarkdownPayRead,
			expected: "- 项目自定义标签（例如 `<pay-read>`、`<power-bi>`、`<wechat-captcha>`）\n\n后面的内容\n",
		},
		{
			name:     "行内代码包裹的 <pay-read> 与后续真实 <pay-read> 共存",
			input:    "上一句: `<pay-read>` 是示例\n\n<pay-read>隐藏内容</pay-read>\n\n结尾",
			payType:  MarkdownPayRead,
			expected: "上一句: `<pay-read>` 是示例\n\n<pay-read></pay-read>\n\n结尾",
		},
		{
			name:     "未闭合的 <pay-read> 当作字面量保留, 不吞掉后续内容",
			input:    "前文\n<pay-read>这里不是付费区块, 只是个忘写尾标签的示例\n\n后文保留\n",
			payType:  MarkdownPayRead,
			expected: "前文\n<pay-read>这里不是付费区块, 只是个忘写尾标签的示例\n\n后文保留\n",
		},
		{
			name:     "行内代码不跨行: 未闭合的反引号不影响下一行的真实 <pay-read>",
			input:    "未闭合反引号 `<pay-read 不闭合\n<pay-read>真实内容</pay-read>\n尾\n",
			payType:  MarkdownPayRead,
			expected: "未闭合反引号 `<pay-read 不闭合\n<pay-read></pay-read>\n尾\n",
		},
		{
			name:     "fenced code block 内的 <pay-read> 不会被剥离",
			input:    "```html\n<pay-read>示例内容</pay-read>\n```\n\n<pay-read>真实</pay-read>\n尾",
			payType:  MarkdownPayRead,
			expected: "```html\n<pay-read>示例内容</pay-read>\n```\n\n<pay-read></pay-read>\n尾",
		},

		// .bug/260426-03 同源场景: pay-download
		{
			name:     "行内代码包裹的 <pay-download> 不会被剥离 (多个并列)",
			input:    "- 项目自定义标签（例如 `<pay-download>`、`<power-bi>`、`<wechat-captcha>`）\n\n后面的内容\n",
			payType:  MarkdownPayDownload,
			expected: "- 项目自定义标签（例如 `<pay-download>`、`<power-bi>`、`<wechat-captcha>`）\n\n后面的内容\n",
		},
		{
			name:     "行内代码包裹的 <pay-download> 与后续真实 <pay-download> 共存",
			input:    "上一句: `<pay-download>` 是示例\n\n<pay-download>隐藏内容</pay-download>\n\n结尾",
			payType:  MarkdownPayDownload,
			expected: "上一句: `<pay-download>` 是示例\n\n<pay-download></pay-download>\n\n结尾",
		},
		{
			name:     "未闭合的 <pay-download> 当作字面量保留, 不吞掉后续内容",
			input:    "前文\n<pay-download>使用者忘写了尾标签\n\n后文保留\n",
			payType:  MarkdownPayDownload,
			expected: "前文\n<pay-download>使用者忘写了尾标签\n\n后文保留\n",
		},
		{
			name:     "行内代码不跨行: 未闭合的反引号不影响下一行的真实 <pay-download>",
			input:    "未闭合反引号 `<pay-download 不闭合\n<pay-download>真实内容</pay-download>\n尾\n",
			payType:  MarkdownPayDownload,
			expected: "未闭合反引号 `<pay-download 不闭合\n<pay-download></pay-download>\n尾\n",
		},
		{
			name:     "fenced code block 内的 <pay-download> 不会被剥离",
			input:    "```html\n<pay-download>示例内容</pay-download>\n```\n\n<pay-download>真实</pay-download>\n尾",
			payType:  MarkdownPayDownload,
			expected: "```html\n<pay-download>示例内容</pay-download>\n```\n\n<pay-download></pay-download>\n尾",
		},

		// .bug/260426-03 同源场景: pay-video
		{
			name:     "行内代码包裹的 <pay-video> 不会被剥离 (多个并列)",
			input:    "- 项目自定义标签（例如 `<pay-video>`、`<power-bi>`、`<wechat-captcha>`）\n\n后面的内容\n",
			payType:  MarkdownPayVideo,
			expected: "- 项目自定义标签（例如 `<pay-video>`、`<power-bi>`、`<wechat-captcha>`）\n\n后面的内容\n",
		},
		{
			name:     "行内代码包裹的 <pay-video> 与后续真实 <pay-video> 共存",
			input:    "上一句: `<pay-video>` 是示例\n\n<pay-video>素材链接</pay-video>\n\n结尾",
			payType:  MarkdownPayVideo,
			expected: "上一句: `<pay-video>` 是示例\n\n<pay-video has-material></pay-video>\n\n结尾",
		},
		{
			name:     "未闭合的 <pay-video> 当作字面量保留, 不吞掉后续内容",
			input:    "前文\n<pay-video>使用者忘写了尾标签\n\n后文保留\n",
			payType:  MarkdownPayVideo,
			expected: "前文\n<pay-video>使用者忘写了尾标签\n\n后文保留\n",
		},
		{
			name:     "行内代码不跨行: 未闭合的反引号不影响下一行的真实 <pay-video>",
			input:    "未闭合反引号 `<pay-video 不闭合\n<pay-video>真实内容</pay-video>\n尾\n",
			payType:  MarkdownPayVideo,
			expected: "未闭合反引号 `<pay-video 不闭合\n<pay-video has-material></pay-video>\n尾\n",
		},
		{
			name:     "fenced code block 内的 <pay-video> 不会被剥离",
			input:    "```html\n<pay-video>示例内容</pay-video>\n```\n\n<pay-video>真实</pay-video>\n尾",
			payType:  MarkdownPayVideo,
			expected: "```html\n<pay-video>示例内容</pay-video>\n```\n\n<pay-video has-material></pay-video>\n尾",
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
		{
			name:     "fenced code block 内的 pay-video 保留, 其他付费标签继续替换",
			input:    "```\n<pay-video>\n除视频外的其他隐藏内容, 如附件下载等; 若没有则将标签设置为一行\n<pay-membership></pay-membership>\n</pay-video>\n```\n\n<pay-membership></pay-membership>\n\n<pay-read>\n您付费阅读的内容\n</pay-read>\n\n<pay-download>\n您的付费下载内容\n</pay-download>\n\n<pay-key id=\"您的key\" title=\"您的标题\" description=\"您的说明\"></pay-key>\n\n<video-player video-type=\"hls\" id=\"m-2-7f9d0d9c\"></video-player>",
			expected: "```\n<pay-video>\n除视频外的其他隐藏内容, 如附件下载等; 若没有则将标签设置为一行\n<pay-membership></pay-membership>\n</pay-video>\n```\n\n<pay-membership></pay-membership>\n\n<pay-read></pay-read>\n\n<pay-download></pay-download>\n\n<pay-key id=\"您的key\" title=\"您的标题\" description=\"您的说明\"></pay-key>\n\n<video-player video-type=\"hls\" id=\"m-2-7f9d0d9c\"></video-player>",
		},
		{
			name:     "复现行内代码包裹的 <pay-read>/<power-bi>/<wechat-captcha> 不会被剥离",
			input:    "\n### 2.2 支持这些增强语法\n\n\n- 项目自定义标签（例如 `<pay-read>`、`<power-bi>`、`<wechat-captcha>`）\n\n后面的内容\n",
			expected: "\n### 2.2 支持这些增强语法\n\n\n- 项目自定义标签（例如 `<pay-read>`、`<power-bi>`、`<wechat-captcha>`）\n\n后面的内容\n",
		},
		{
			name:     "行内代码包裹的 pay-* 与后续真实付费块共存",
			input:    "示例: `<pay-read>`、`<pay-download>`。\n\n<pay-read>阅读隐藏</pay-read>\n\n<pay-download>下载隐藏</pay-download>\n\n<pay-video>视频隐藏</pay-video>\n\n尾部\n",
			expected: "示例: `<pay-read>`、`<pay-download>`。\n\n<pay-read></pay-read>\n\n<pay-download></pay-download>\n\n<pay-video has-material></pay-video>\n\n尾部\n",
		},
		{
			name:     "未闭合的付费起标签作为字面量保留, 不会吞掉后续内容",
			input:    "前文\n<pay-read>使用者忘写了尾标签\n\n后文保留\n",
			expected: "前文\n<pay-read>使用者忘写了尾标签\n\n后文保留\n",
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
