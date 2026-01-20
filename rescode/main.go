//
// FilePath    : go-utils\rescode\main.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 响应码注册
//

// Package rescode 响应状态码
package rescode

import "maps"

// StatusCodeType 状态码类型
type StatusCodeType int

// CodeMsgMap 状态码信息映射
type CodeMsgMap map[StatusCodeType]string

// CodeMsgMapDoc 状态码信息文档结构体
type CodeMsgMapDoc struct {
	Title string         // 标题
	Start StatusCodeType // 起始状态码
	Map   CodeMsgMap     // 状态码映射
}

// StatusCodeMsgMap 全局状态码信息映射
var StatusCodeMsgMap = make(CodeMsgMap)

// StatusCodeMsgMapDoc 文档使用的状态码信息
var StatusCodeMsgMapDoc = make(map[StatusCodeType]CodeMsgMapDoc)

// MsgCode 状态码信息
type MsgCode struct {
	StatusCode StatusCodeType
}

// Msg 返回状态码信息中的消息
func (m *MsgCode) Msg() string {
	msg, ok := StatusCodeMsgMap[m.StatusCode]
	if !ok {
		return "未知状态码"
	}

	return msg
}

// RegisterCodes 注册状态码信息
func RegisterCodes(codeMap map[StatusCodeType]string) {
	maps.Copy(StatusCodeMsgMap, codeMap)
}

// RegisterDocCodes 注册状态码文档信息, 用于生成文档
func RegisterDocCodes(start StatusCodeType, title string, codeMap map[StatusCodeType]string) {
	StatusCodeMsgMapDoc[start] = CodeMsgMapDoc{
		Title: title,
		Start: start,
		Map:   codeMap,
	}
}

// SortStatusCodeTypeSlice 对 StatusCodeType 切片进行排序, isAsc 为 true 则升序排序, 否则降序排序
func SortStatusCodeTypeSlice(codes []StatusCodeType, isAsc bool) {
	if isAsc {
		sortAsc(codes)
	} else {
		sortDesc(codes)
	}
}

// sortAsc 对 StatusCodeType 切片进行升序冒泡排序
func sortAsc(codes []StatusCodeType) {
	for i := 0; i < len(codes)-1; i++ {
		for j := 0; j < len(codes)-i-1; j++ {
			if codes[j] > codes[j+1] {
				codes[j], codes[j+1] = codes[j+1], codes[j]
			}
		}
	}
}

// sortDesc 对 StatusCodeType 切片进行降序冒泡排序
func sortDesc(codes []StatusCodeType) {
	for i := 0; i < len(codes)-1; i++ {
		for j := 0; j < len(codes)-i-1; j++ {
			if codes[j] < codes[j+1] {
				codes[j], codes[j+1] = codes[j+1], codes[j]
			}
		}
	}
}
