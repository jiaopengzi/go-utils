//
// FilePath    : go-utils\email\helper_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 测试辅助工具
//

package email

import "os"

// writeTestFile 写入测试用临时文件
func writeTestFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
