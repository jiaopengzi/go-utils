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
	"sort"
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

// buildMarkdownLineOffsets 构建 markdown 每一行的起始偏移.
//   - input, 原始 markdown 内容.
//
// 返回值 []int, 每一行在原始字符串中的起始位置.
func buildMarkdownLineOffsets(input string) []int {
	lineOffsets := []int{0}

	for index := 0; index < len(input); index++ {
		if input[index] == '\n' && index+1 < len(input) {
			lineOffsets = append(lineOffsets, index+1)
		}
	}

	return lineOffsets
}

// findMarkdownLineIndex 根据字符偏移定位所在行.
//   - offset, 当前字符偏移.
//   - lineOffsets, 每一行的起始偏移集合.
//
// 返回值 int, 当前字符所在的行下标.
func findMarkdownLineIndex(offset int, lineOffsets []int) int {
	if len(lineOffsets) == 0 {
		return 0
	}

	lineIndex := sort.Search(len(lineOffsets), func(index int) bool {
		return lineOffsets[index] > offset
	})

	if lineIndex == 0 {
		return 0
	}

	return lineIndex - 1
}

// countMarkdownBacktickRun 统计当前位置连续反引号的数量.
//   - input, 原始 markdown 内容.
//   - start, 开始统计的位置.
//
// 返回值 int, 连续反引号数量.
func countMarkdownBacktickRun(input string, start int) int {
	backtickCount := 0

	for start+backtickCount < len(input) && input[start+backtickCount] == '`' {
		backtickCount++
	}

	return backtickCount
}

// getMarkdownLineByIndex 获取指定行的原始内容, 不包含换行符.
//   - input, 原始 markdown 内容.
//   - lineOffsets, 每一行的起始偏移集合.
//   - lineIndex, 目标行下标.
//
// 返回值 string, 指定行内容.
func getMarkdownLineByIndex(input string, lineOffsets []int, lineIndex int) string {
	lineStart := lineOffsets[lineIndex]
	lineEnd := len(input)
	if lineIndex+1 < len(lineOffsets) {
		lineEnd = lineOffsets[lineIndex+1] - 1
	}

	return input[lineStart:lineEnd]
}

// findMarkdownFenceTrimStart 计算 fenced code block 候选行的有效起始位置.
//   - line, 单行 markdown 内容.
//
// 返回值 int, 去除最多 3 个前导空格后的起始位置; bool, 是否满足 fenced code block 的缩进要求.
func findMarkdownFenceTrimStart(line string) (int, bool) {
	trimStart := 0
	for trimStart < len(line) && trimStart < 3 && line[trimStart] == ' ' {
		trimStart++
	}

	if trimStart < len(line) && line[trimStart] == ' ' {
		return 0, false
	}

	return trimStart, true
}

// parseMarkdownFenceLine 识别当前行是否为 fenced code block 的开始或结束标记.
//   - line, 单行 markdown 内容.
//   - inFencedCodeBlock, 当前是否位于 fenced code block 内.
//   - fenceBacktickCount, 当前 fenced code block 使用的反引号数量.
//
// 返回值 bool, 是否识别到 fence 行; bool, 识别后是否位于 fenced code block 内; int, 更新后的 fence 反引号数量.
func parseMarkdownFenceLine(line string, inFencedCodeBlock bool, fenceBacktickCount int) (bool, bool, int) {
	trimStart, ok := findMarkdownFenceTrimStart(line)
	if !ok {
		return false, inFencedCodeBlock, fenceBacktickCount
	}

	backtickCount := countMarkdownBacktickRun(line, trimStart)
	if backtickCount < 3 {
		return false, inFencedCodeBlock, fenceBacktickCount
	}

	if !inFencedCodeBlock {
		return true, true, backtickCount
	}

	if backtickCount >= fenceBacktickCount && strings.TrimSpace(line[trimStart+backtickCount:]) == "" {
		return true, false, 0
	}

	return false, inFencedCodeBlock, fenceBacktickCount
}

// isMarkdownFencedLine 判断字符偏移是否位于 fenced code block 行内.
//   - offset, 当前字符偏移.
//   - lineOffsets, 每一行的起始偏移集合.
//   - fencedLineFlags, 每一行是否位于 fenced code block 内的标记.
//
// 返回值 bool, true 表示当前字符位于 fenced code block 行内.
func isMarkdownFencedLine(offset int, lineOffsets []int, fencedLineFlags []bool) bool {
	lineIndex := findMarkdownLineIndex(offset, lineOffsets)
	return lineIndex < len(fencedLineFlags) && fencedLineFlags[lineIndex]
}

// consumeMarkdownInlineCodeBackticks 处理 pay block 扫描中的行内代码反引号状态.
//   - input, 原始 markdown 内容.
//   - i, 当前扫描位置.
//   - inlineCodeBacktickCount, 当前行内代码使用的反引号数量.
//
// 返回值 int, 更新后的扫描位置; int, 更新后的行内代码反引号数量; bool, 是否消费了反引号序列.
func consumeMarkdownInlineCodeBackticks(input string, i int, inlineCodeBacktickCount int) (int, int, bool) {
	backtickCount := countMarkdownBacktickRun(input, i)
	if backtickCount == 0 {
		return i, inlineCodeBacktickCount, false
	}

	if inlineCodeBacktickCount == 0 {
		inlineCodeBacktickCount = backtickCount
	} else if backtickCount == inlineCodeBacktickCount {
		inlineCodeBacktickCount = 0
	}

	return i + backtickCount, inlineCodeBacktickCount, true
}

// consumeMarkdownPayBlockTag 处理 pay block 扫描中的起止标签.
//   - input, 原始 markdown 内容.
//   - i, 当前扫描位置.
//   - startTag, 当前付费标签开始标记.
//   - endTag, 当前付费标签结束标记.
//   - depth, 当前嵌套深度.
//
// 返回值 int, 更新后的扫描位置; int, 更新后的嵌套深度; bool, 是否消费了标签.
func consumeMarkdownPayBlockTag(input string, i int, startTag, endTag string, depth int) (int, int, bool) {
	switch {
	case strings.HasPrefix(input[i:], startTag):
		return i + len(startTag), depth + 1, true
	case strings.HasPrefix(input[i:], endTag):
		return i + len(endTag), depth - 1, true
	default:
		return i, depth, false
	}
}

// collectMarkdownFencedCodeLineFlags 标记每一行是否位于 fenced code block 内.
//   - input, 原始 markdown 内容.
//   - lineOffsets, 每一行的起始偏移集合.
//
// 返回值 []bool, 与行数等长的标记切片, true 表示当前行位于 fenced code block 内.
func collectMarkdownFencedCodeLineFlags(input string, lineOffsets []int) []bool {
	fencedLineFlags := make([]bool, len(lineOffsets))
	inFencedCodeBlock := false
	fenceBacktickCount := 0

	for lineIndex := range lineOffsets {
		line := getMarkdownLineByIndex(input, lineOffsets, lineIndex)
		if inFencedCodeBlock {
			fencedLineFlags[lineIndex] = true
		}

		matched, nextInFencedCodeBlock, nextFenceBacktickCount := parseMarkdownFenceLine(line, inFencedCodeBlock, fenceBacktickCount)
		if !matched {
			continue
		}

		if !inFencedCodeBlock && nextInFencedCodeBlock {
			fencedLineFlags[lineIndex] = true
		}

		inFencedCodeBlock = nextInFencedCodeBlock
		fenceBacktickCount = nextFenceBacktickCount
	}

	return fencedLineFlags
}

// ReplaceMarkdownPayTagToEmpty 替换 markdown 付费标签为空标签.
//   - input, 原始 markdown 内容.
//   - payType, 需要替换的付费标签类型.
//
// 返回值 string, 替换后的 markdown 内容.
func ReplaceMarkdownPayTagToEmpty(input string, payType MarkdownPayType) string {
	startTag := GetMarkdownPayTag(payType, TagStart) // 开始标签
	endTag := GetMarkdownPayTag(payType, TagEnd)     // 结束标签
	emptyTag := GetMarkdownEmptyPayTag(payType)      // 空标签
	lineOffsets := buildMarkdownLineOffsets(input)
	fencedLineFlags := collectMarkdownFencedCodeLineFlags(input, lineOffsets)

	var result strings.Builder // 结果使用 strings.Builder 构造, 避免频繁字符串拼接

	i := 0          // 当前处理位置
	n := len(input) // 输入字符串长度

	for i < n {
		lineIndex := findMarkdownLineIndex(i, lineOffsets)
		if lineIndex < len(fencedLineFlags) && fencedLineFlags[lineIndex] {
			result.WriteByte(input[i])
			i++
			continue
		}

		if strings.HasPrefix(input[i:], startTag) {
			result.WriteString(emptyTag)
			i += len(startTag)
			i = skipPayBlock(input, i, n, startTag, endTag, lineOffsets, fencedLineFlags)
			continue
		}

		result.WriteByte(input[i])
		i++
	}

	return result.String()
}

// skipPayBlock 跳过当前位置开始的付费块内容, 同时忽略 fenced code block 与行内代码中的标签.
//   - input, 原始 markdown 内容.
//   - i, 当前扫描位置.
//   - n, 原始 markdown 长度.
//   - startTag, 当前付费标签开始标记.
//   - endTag, 当前付费标签结束标记.
//   - lineOffsets, 每一行的起始偏移集合.
//   - fencedLineFlags, 每一行是否位于 fenced code block 内的标记.
//
// 返回值 int, 跳过当前付费块后的扫描位置.
func skipPayBlock(input string, i int, n int, startTag, endTag string, lineOffsets []int, fencedLineFlags []bool) int {
	depth := 1 // 嵌套深度, 初始为 1
	inlineCodeBacktickCount := 0

	for i < n && depth > 0 {
		if isMarkdownFencedLine(i, lineOffsets, fencedLineFlags) {
			i++
			continue
		}

		nextIndex, nextInlineCodeBacktickCount, consumed := consumeMarkdownInlineCodeBackticks(input, i, inlineCodeBacktickCount)
		if consumed {
			i = nextIndex
			inlineCodeBacktickCount = nextInlineCodeBacktickCount
			continue
		}

		if inlineCodeBacktickCount != 0 {
			i++
			continue
		}

		nextIndex, nextDepth, consumed := consumeMarkdownPayBlockTag(input, i, startTag, endTag, depth)
		if consumed {
			i = nextIndex
			depth = nextDepth
			continue
		}

		i++
	}

	return i
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
