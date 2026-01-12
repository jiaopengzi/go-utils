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

// SQLTemplate 通过 sql 模板 templateStr, varMap 模板变量, 生成 sql 查询语句
func SQLTemplate(templateStr string, varMap map[string]string) (string, error) {
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
