//
// FilePath    : go-utils\sanitize.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 用户产生内容的 HTML 消毒(纵深防御, 剥离常见 XSS 向量)
//

package utils

import (
	"regexp"
	"strings"
)

// dangerousTags 需要(连同内容)整块剥离的标签
var dangerousTags = []string{"script", "iframe", "object", "embed", "style", "form"}

var (
	// 各危险标签的成对块(开标签+内容+闭标签)与未闭合开标签(到文本结尾)
	reDangerousBlocks   []*regexp.Regexp
	reDangerousOpenTags []*regexp.Regexp
	reEventAttr         = regexp.MustCompile(`(?i)\bon[a-z]+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	reDangerousProto    = regexp.MustCompile(`(?i)(javascript|vbscript)\s*:`)
	reDataURL           = regexp.MustCompile(`(?i)data\s*:\s*text\s*/\s*html`)
)

func init() {
	for _, tag := range dangerousTags {
		reDangerousBlocks = append(reDangerousBlocks,
			regexp.MustCompile(`(?is)<\s*`+tag+`\b[^>]*>.*?<\s*/\s*`+tag+`\s*>`))
		reDangerousOpenTags = append(reDangerousOpenTags,
			regexp.MustCompile(`(?is)<\s*`+tag+`\b[^>]*>.*`))
	}
}

// SanitizeHTMLContent 对用户产生内容做 HTML 消毒, 作为前端转义之外的纵深防御:
// 剥离 script/iframe 等危险标签块、标签内 on* 事件属性以及 javascript:/vbscript:/data:text/html 协议.
// 普通文本与无害标签不受影响.
func SanitizeHTMLContent(content string) string {
	if content == "" {
		return content
	}

	result := content
	for _, re := range reDangerousBlocks {
		result = re.ReplaceAllString(result, "")
	}

	for _, re := range reDangerousOpenTags {
		result = re.ReplaceAllString(result, "")
	}

	result = reEventAttr.ReplaceAllString(result, "")
	result = reDangerousProto.ReplaceAllString(result, "")
	result = reDataURL.ReplaceAllString(result, "")

	return strings.TrimSpace(result)
}
