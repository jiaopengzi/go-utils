//
// FilePath    : go-utils\convert.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 转换工具
//

package utils

import (
	"regexp"
	"strconv"
	"strings"
)

// MapKeysToSlice 将 map 的 key 转换为切片, 保留 map 键的类型
func MapKeysToSlice[K comparable, V any](m map[K]V) []K {
	var keys []K
	for key := range m {
		keys = append(keys, key)
	}

	return keys
}

// Int64Ptr 将 int64 转换为指针
func Int64Ptr(i int64) *int64 { return &i }

// StrToUint64 将字符串转换为 uint64
func StrToUint64(s string) uint64 {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}

	return v
}

// StrToInt64 将字符串转换为 int64
func StrToInt64(s string) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}

	return v
}

// StrYuanToInt64Fen 将字符串包含小数的元转换为 int64 分
// 例如 "12.34" 转换为 1234
// 如果字符串无法转换为数字，则返回 0
func StrYuanToInt64Fen(s string) int64 {
	if s == "" {
		return 0
	}

	// 将字符串转换为 float64
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}

	// 将 float64 转换为 int64 分
	return int64(f * 100)
}

// Int64FenToStrYuan 将 int64 转换为字符串包含小数的元
// 例如 1234 转换为 "12.34"
// 如果 int64 为 0，则返回 "0.00"
func Int64FenToStrYuan(i int64) string {
	if i == 0 {
		return "0.00"
	}

	// 将 int64 转换为 float64
	f := float64(i) / 100.0

	// 将 float64 转换为字符串，保留两位小数
	return strconv.FormatFloat(f, 'f', 2, 64)
}

// StrToInt 将字符串转换为 int
func StrToInt(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}

	return v
}

// SliceStrToUint64 字符串切片转 uint64 切片
func SliceStrToUint64(list []string) []uint64 {
	var result []uint64
	for _, item := range list {
		result = append(result, StrToUint64(item))
	}

	return result
}

// Uint64ToStr 将 uint64 转字符串
func Uint64ToStr(i uint64) string {
	return strconv.FormatUint(i, 10)
}

// Int64ToStr 将 int64 转字符串
func Int64ToStr(i int64) string {
	return strconv.FormatInt(i, 10)
}

// IntToStr 将 int 转字符串
func IntToStr(i int) string {
	return strconv.Itoa(i)
}

// SliceUint64ToStr uint64 切片转字符串切片
func SliceUint64ToStr(list []uint64) []string {
	var result []string
	for _, item := range list {
		result = append(result, Uint64ToStr(item))
	}

	return result
}

// StrIsUint64 将字符串转换为 uint64
func StrIsUint64(s any) bool {
	// 判断是否为空
	if s == "" {
		return false
	}

	// 判断是否为数字
	if _, ok := s.(uint64); ok {
		return true
	}

	// 如果是字符串
	if str, ok := s.(string); ok {
		_, err := strconv.ParseUint(str, 10, 64)
		return err == nil
	}

	return false
}

// SplitCommand 按照空格拆分字符串, 但是引号内的内容不拆分; keepQuotes 为 true 保留引号, 为 false 不保留, 例如 "ls -l" 拆分为 ["ls", "-l"]
func SplitCommand(input string, keepQuotes bool) []string {
	// 正则表达式匹配双引号、单引号包裹的部分和非空格部分
	re := regexp.MustCompile(`"[^"]*"|'[^']*'|\S+`)
	matches := re.FindAllString(input, -1)

	// 去除双引号、单引号
	if !keepQuotes {
		for i, match := range matches {
			if strings.HasPrefix(match, "\"") && strings.HasSuffix(match, "\"") {
				matches[i] = match[1 : len(match)-1]
			} else if strings.HasPrefix(match, "'") && strings.HasSuffix(match, "'") {
				matches[i] = match[1 : len(match)-1]
			}
		}
	}

	return matches
}

// MakeSet 将字符串切片转换为 map
func MakeSet(slice []string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, item := range slice {
		set[item] = struct{}{}
	}

	return set
}

// 预编译正则表达式
var (
	// 匹配 NOTE 和 STYLE 块
	reNoteStyle = regexp.MustCompile(`(?s)(NOTE.*?(\r?\n){2}|STYLE.*?(\r?\n){2})`)

	// 匹配时间表达式,中间符号 --> 在后续判断 支持的时间格式包括 hh:mm:ss.mmm, mm:ss.mmm, ss.mmm
	reTimeExpression = regexp.MustCompile(`\b(?:\d{2}:)?(?:\d{2}:)?\d{2}\.\d{3}.*(?:\d{2}:)?(?:\d{2}:)?\d{2}\.\d{3}\b|.*-->.*`)

	// 匹配时间格式 hh:mm:ss.mmm, mm:ss.mmm, ss.mmm
	reTimeFormat = regexp.MustCompile(`^(?:\d{2}:)?(?:\d{2}:)?\d{2}\.\d{3}$`)
)

// IsWebvtt 判断是否为 Webvtt 字幕
func IsWebvtt(content string) (bool, string) {
	if strings.TrimSpace(content) == "" {
		return false, "字幕内容不能为空"
	}

	// 单行特殊处理
	if strings.Count(content, "\n") == 0 {
		if strings.HasPrefix(content, "WEBVTT") {
			return true, ""
		}
		return false, "字幕需要以 WEBVTT 开头"
	}

	if !hasWebvttHeader(content) {
		return false, "字幕需要以 WEBVTT 开头"
	}

	// 移除 NOTE/STYLE 块后进行时间表达式与内容校验
	filtered := removeNoteStyle(content)

	matches := findTimeExpressions(filtered)
	if len(matches) == 0 {
		return false, "需要包含时间表达式 hh:mm:ss.mmm, mm:ss.mmm, ss.mmm"
	}

	if ok, msg := validateTimeExpressions(matches); !ok {
		return false, msg
	}

	if ok, msg := validateSubtitleContentAfterTiming(filtered); !ok {
		return false, msg
	}

	return true, ""
}

func hasWebvttHeader(content string) bool {
	return strings.HasPrefix(content, "WEBVTT")
}

func removeNoteStyle(content string) string {
	return reNoteStyle.ReplaceAllString(content, "\n")
}

func findTimeExpressions(content string) []string {
	return reTimeExpression.FindAllString(content, -1)
}

func validateTimeExpressions(matches []string) (bool, string) {
	for _, match := range matches {
		parts := strings.Split(match, " --> ")
		if len(parts) != 2 {
			return false, "时间表达式中需要 --> 分隔符"
		}

		startTime := strings.TrimSpace(parts[0])
		endTime := strings.TrimSpace(parts[1])

		if !reTimeFormat.MatchString(startTime) || !reTimeFormat.MatchString(endTime) {
			return false, "时间格式错误，支持 hh:mm:ss.mmm, mm:ss.mmm, ss.mmm"
		}
	}
	return true, ""
}

func validateSubtitleContentAfterTiming(content string) (bool, string) {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if reTimeExpression.MatchString(line) {
			if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) == "" {
				return false, "时间表达式后需要有字幕内容,不能为空"
			}
		}
	}
	return true, ""
}

// StrToPtr 将字符串返回为指针
func StrToPtr(s string) *string {
	return &s
}
