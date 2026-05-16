//
// FilePath    : go-utils\query.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 查询工具
//

package utils

import (
	"bytes"
	"sync"
	"text/template"
)

// UnsafeSQLTemplate 使用 text/template 将 varMap 中的变量替换到 SQL 模板中并生成 SQL 语句。
//
// 警告：本函数直接将变量值拼入 SQL 字符串，不提供任何转义或参数化保护。
// 严禁传入任何外部用户输入，仅限内部硬编码的列名/表名等可信标识符使用。
// 新增查询请优先使用参数化 SQL 或 ORM 条件构造。
func UnsafeSQLTemplate(templateStr string, varMap map[string]string) (string, error) {
	// 使用text/template包替换占位符
	t, err := template.New("sql").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var queryBuffer bytes.Buffer

	err = t.Execute(&queryBuffer, varMap)
	if err != nil {
		return "", err
	}

	return queryBuffer.String(), nil
}

// ConcurrentQuery 通用并发查询
//   - wg: *sync.WaitGroup 等待组
//   - queryFunc: func() (T, error) 查询函数
//   - resultChan: chan<- T 结果通道
//   - errorChan: chan<- error 错误通道
func ConcurrentQuery[T any](wg *sync.WaitGroup, queryFunc func() (T, error), resultChan chan<- T, errorChan chan<- error) {
	defer wg.Done()

	result, err := queryFunc()
	if err != nil {
		errorChan <- err
		return
	}
	resultChan <- result
}
