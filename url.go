//
// FilePath    : go-utils\url.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : URL 工具
//

package utils

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// ParseNumber 根据字符串 s 解析数字，如果解析失败则返回 fallback
func ParseNumber(s string, fallback int) int {
	v, err := strconv.Atoi(s)
	if err == nil {
		return v
	}

	return fallback
}

// ParseBool 根据字符串 s 解析布尔值，如果解析失败则返回 fallback
func ParseBool(s string, fallback bool) bool {
	v, err := strconv.ParseBool(s)
	if err == nil {
		return v
	}

	return fallback
}

// ParsePath 解析 URL 的路径部分，如果为空则返回 fallback
func ParsePath(u *url.URL, fallback string) string {
	p := u.Path
	if p == "" {
		return fallback
	}

	return p
}

// JoinURL 拼接 baseURL 和路径元素，返回完整的 URL, paths 可以是多个
func JoinURL(baseURL string, paths ...string) (string, error) {
	// 检查 baseURL 是否为空
	if baseURL == "" {
		return "", errors.New("baseURL cannot be empty")
	}

	// 解析 baseURL
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid baseURL: %v", err)
	}

	// 检查路径部分是否包含非法字符
	for _, p := range paths {
		if strings.ContainsAny(p, "?#") {
			return "", errors.New("paths cannot contain '?' or '#'")
		}
	}

	// 使用 path.Join 拼接路径部分
	fullPath := path.Join(paths...)
	u.Path = path.Join(u.Path, fullPath)

	return u.String(), nil
}

// AddQueryParams 向 baseURL 添加查询参数，返回完整的 URL
func AddQueryParams(baseURL string, params map[string]string) (string, error) {
	// 解析 baseURL
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid baseURL: %v", err)
	}

	// 获取查询参数对象
	q := u.Query()

	// 添加新的查询参数
	for key, value := range params {
		q.Add(key, value)
	}

	// 设置查询参数
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// IsImageURL 判断字符串 u 是否为图片 URL
func IsImageURL(u string) bool {
	// 判断是否为空
	if u == "" {
		return false
	}

	// 使用正则表达式匹配图片文件扩展名和 URL 格式
	imageURLPattern := `(?i)^(https?:\/\/.*\.(jpg|jpeg|png|gif|bmp|tiff))$`

	matched, err := regexp.MatchString(imageURLPattern, u)
	if err != nil {
		return false
	}

	return matched
}

// EncodeURL 使用 URL 编码对字符串 s 进行编码
func EncodeURL(s string) string {
	return url.QueryEscape(s)
}

// isURLComponentUnreserved 判断字节 c 是否为 encodeURIComponent 的不转义字符.
// JS encodeURIComponent 仅保留字母/数字与 -_.!~*'() 字面量, 其余全部转义.
func isURLComponentUnreserved(c byte) bool {
	if strings.ContainsRune("-_.!~*'()", rune(c)) {
		return true
	}
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')
}

// EncodeURLComponent 按 JS encodeURIComponent 语义对字符串 s 做百分号编码.
// EncodeURL(url.QueryEscape) 的关键差异: 空格编码为 %20 而非 "+",
// "!'()*~" 保留字面量. 该形态可直接用于 URL 路径段, 且与前端 encodeURIComponent 的
// 编解码往返恒等, 保证按 slug 精确匹配时两侧形态一致 (QueryEscape 形态中的字面 "+"
// 经前端 decodeURIComponent/encodeURIComponent 往返会变为 %2B 导致未命中).
func EncodeURLComponent(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if isURLComponentUnreserved(c) {
			b.WriteByte(c)
			continue
		}
		fmt.Fprintf(&b, "%%%02X", c)
	}
	return b.String()
}

// GenerateSlugByName 根据名称生成 slug
func GenerateSlugByName(s string) (string, error) {
	//	判断 s 是否为空
	if s == "" {
		return "", ErrNotEmpty
	}

	//	生成 slug (改用 JS encodeURIComponent 兼容编码, 空格为 %20 而非 "+")
	slug := EncodeURLComponent(s)

	//	判断 slug 长度是否超过 255
	if len(slug) > 255 {
		return "", ErrSlugTooLong
	}

	return slug, nil
}
