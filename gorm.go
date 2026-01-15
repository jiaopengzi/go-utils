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

// CheckUpdateWithOptimisticLock 检查使用乐观锁更新记录时是否成功更新
func CheckUpdateWithOptimisticLock(result *gorm.DB) error {
	// 检查是否成功更新了行
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		// 影响行数为0, 表示乐观锁更新失败
		return ErrUpdateWithOptimisticLockFailed
	}

	return nil
}
