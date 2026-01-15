//
// FilePath    : go-utils\gorm.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : gorm 相关工具
//

package utils

import (
	"gorm.io/gorm"
)

// CheckGormUpdate 检查 gorm 更新操作的结果
func CheckGormUpdate(result *gorm.DB) error {
	// 检查是否成功更新了行
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		// 没有更新任何行
		return ErrUpdateRowsAffectedZero
	}

	return nil
}
