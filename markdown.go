//
// FilePath    : go-utils\markdown.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : markdown 工具
//

package utils

import (
	"fmt"
	"strings"
)

// MarkdownPayType markdown 付费类型
type MarkdownPayType string

const (
	MarkdownPayRead     MarkdownPayType = "pay-read"     // 付费阅读
	MarkdownPayDownload MarkdownPayType = "pay-download" // 付费下载
	MarkdownPayVideo    MarkdownPayType = "pay-video"    // 付费视频
)

const (
	TagStart = "start" // 标签开始
	TagEnd   = "end"   // 标签结束
	Backtick = '`'     // markdown 代码块使用反引号 ` 包裹
)

// AllMarkdownPayType 所有的 markdown 付费类型
var AllMarkdownPayType = []MarkdownPayType{
	MarkdownPayRead,
	MarkdownPayDownload,
	MarkdownPayVideo,
}

// GetMarkdownPayTag 获取 markdown 付费标签
func GetMarkdownPayTag(payType MarkdownPayType, flag string) string {
	switch flag {
	case TagStart:
		return fmt.Sprintf("<%s>", payType)
	case TagEnd:
		return fmt.Sprintf("</%s>", payType)
	}

	return ""
}

// GetMarkdownEmptyPayTag 获取 markdown 空的付费标签
func GetMarkdownEmptyPayTag(payType MarkdownPayType) string {
	return fmt.Sprintf("<%s></%s>", payType, payType)
}

// ReplaceMarkdownPayTagToEmpty 替换 markdown 付费标签为空标签
func ReplaceMarkdownPayTagToEmpty(input string, payType MarkdownPayType) string {
	startTag := GetMarkdownPayTag(payType, TagStart) // 开始标签
	endTag := GetMarkdownPayTag(payType, TagEnd)     // 结束标签
	emptyTag := GetMarkdownEmptyPayTag(payType)      // 空标签

	var result strings.Builder // 结果使用 strings.Builder 构造, 避免频繁字符串拼接

	i := 0             // 当前处理位置
	n := len(input)    // 输入字符串长度
	backtickCount := 0 // 在 markdown 中反引号出现次数, >0 表示在代码内

	for i < n {
		if backtickCount == 0 {
			// 不在代码块内, 检测标签
			if strings.HasPrefix(input[i:], startTag) {
				// 找到开始标签, 替换整个付费块为空标签
				result.WriteString(emptyTag)

				i += len(startTag) // 移动到开始标签后面

				// 寻找并跳过整个付费块（包含嵌套与代码块内部的反引号处理）
				i, backtickCount = skipPayBlock(input, i, n, startTag, endTag, backtickCount)
			} else {
				// 普通内容, 直接写入
				result.WriteByte(input[i])

				i++
			}
		} else {
			// 在代码内, 直接写入字符, 不处理标签
			if input[i] == Backtick {
				backtickCount--
			}

			result.WriteByte(input[i])

			i++
		}
	}

	return result.String()
}

// skipPayBlock 跳过从当前位置开始的付费块内容，考虑嵌套标签和代码块中的反引号。
// 返回跳过后的索引和更新后的 backtickCount。
func skipPayBlock(input string, i int, n int, startTag, endTag string, backtickCount int) (int, int) {
	depth := 1 // 嵌套深度, 初始为 1

	// 寻找对应的结束标签, 考虑嵌套情况
	for i < n && depth > 0 {
		if input[i] == Backtick {
			backtickCount++
		} else if input[i] == Backtick && backtickCount > 0 {
			backtickCount--
		}

		if backtickCount == 0 {
			// 只有不在代码内, 才检测标签
			switch {
			case strings.HasPrefix(input[i:], startTag):
				depth++
				i += len(startTag)
			case strings.HasPrefix(input[i:], endTag):
				depth--
				i += len(endTag)
			default:
				i++
			}
		} else {
			// 仍在代码内, 直接跳过
			i++
		}
	}

	return i, backtickCount
}

// ReplaceMarkdownPayTagsToEmpty 替换多个 markdown 付费标签为空标签, 一次遍历处理
func ReplaceMarkdownPayTagsToEmpty(input string) string {
	// 输入为空, 直接返回
	if input == "" {
		return ""
	}

	// 依次替换所有的付费标签
	for _, payType := range AllMarkdownPayType {
		input = ReplaceMarkdownPayTagToEmpty(input, payType)
	}

	return input
}

// GenerateMarkdownDetail 生成 markdown 详情内容
func GenerateMarkdownDetail(summary, content string) string {
	// 生成折叠详情模板
	const detailsTemplate = `<details><summary>%s</summary>
<p>

%s

</p>
</details>

`

	// 返回生成的详情内容
	return fmt.Sprintf(detailsTemplate, summary, content)
}

// GenerateMarkdownTable 生成 markdown 表格
func GenerateMarkdownTable(headers []string, rows [][]string) string {
	// |column1|column2|column3|
	// |:---:|:---:|:---:|
	// |content1|content2|content3|
	// |content1|content2|content3|
	// |content1|content2|content3|
	var builder strings.Builder

	// 生成表头
	builder.WriteString("|")

	for _, header := range headers {
		builder.WriteString(header)
		builder.WriteString("|")
	}

	builder.WriteString("\n")

	// 生成分隔行
	builder.WriteString("|")

	for range headers {
		builder.WriteString(":---:")
		builder.WriteString("|")
	}

	builder.WriteString("\n")

	// 生成数据行
	for _, row := range rows {
		builder.WriteString("|")

		for _, cell := range row {
			builder.WriteString(cell)
			builder.WriteString("|")
		}

		builder.WriteString("\n")
	}

	return builder.String()
}
