//
// FilePath    : go-utils\model\migrate.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 数据库迁移
//

package model

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CustomAutoMigrate 自定义迁移
func CustomAutoMigrate(db *gorm.DB) (err error) {
	// 获取所有注册的模型
	models := GetModels()

	// 执行迁移
	err = db.AutoMigrate(models...)
	if err != nil {
		return fmt.Errorf("custom migrate failed: %w", err)
	}

	zap.L().Info("custom migrate success")

	return nil
}
