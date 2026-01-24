//
// FilePath    : go-utils\gorm.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : gorm 相关工具
//

package utils

import (
	"fmt"
	"reflect"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ColCount 是一个常量, 用于 SQL 查询中统计行数
const ColCount = "COUNT(*) as count"

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

// WithTransaction 处理事务的函数
// 借鉴 GORM 原生 Transaction 实现, 使用 panicked 标志模式
// 优点:
//  1. 不使用 recover(), 让 panic 自然传播到上层(如 Gin Recovery)
//  2. 统一的 defer 回滚逻辑, 覆盖 panic、业务错误、commit 失败所有场景
//  3. 添加日志记录便于问题排查
//  4. 支持嵌套事务, 使用 SavePoint 实现
func Transaction(db *gorm.DB, txFunc func(tx *gorm.DB) error) (err error) {
	// panicked 标志: 初始为 true, 只有完全成功才设为 false
	panicked := true

	// 检测是否已在事务中(嵌套事务场景)
	if committer, ok := db.Statement.ConnPool.(gorm.TxCommitter); ok && committer != nil {
		// 嵌套事务: 使用 SavePoint
		spID := fmt.Sprintf("sp%d", time.Now().UnixNano())

		if err = db.SavePoint(spID).Error; err != nil {
			return fmt.Errorf("创建 SavePoint %s 失败: %w", spID, err)
		}

		// 统一的 defer 回滚逻辑
		// 触发回滚的情况: panic(panicked 仍为 true)、业务错误
		defer rollbackOnFailure(&panicked, &err, func() error {
			return db.RollbackTo(spID).Error
		}, spID)

		// 执行事务逻辑, 复用当前事务连接
		if err = txFunc(db); err != nil {
			return fmt.Errorf("嵌套事务执行失败: %w", err)
		}

		// 嵌套事务不需要 Commit, SavePoint 会随外层事务一起提交
		panicked = false

		return nil
	}

	// 非嵌套场景: 开启新事务
	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("开启事务失败: %w", tx.Error)
	}

	// 统一的 defer 回滚逻辑
	// 触发回滚的情况: panic(panicked 仍为 true)、业务错误、commit 失败
	defer rollbackOnFailure(&panicked, &err, func() error {
		return tx.Rollback().Error
	}, "")

	// 执行 txFunc 具体的事务处理逻辑
	if err = txFunc(tx); err != nil {
		return fmt.Errorf("执行事务逻辑失败: %w", err)
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	// 只有完全成功才设为 false, 确保 defer 中不会回滚
	panicked = false

	return nil
}

// rollbackOnFailure 统一的回滚处理函数
//   - panicked: panic 标志指针
//   - err: 错误指针
//   - rollbackFunc: 执行回滚的函数(返回回滚时的错误)
//   - savePoint: SavePoint 名称, 为空表示普通事务回滚
func rollbackOnFailure(panicked *bool, err *error, rollbackFunc func() error, savePoint string) {
	// 边界条件: 未发生 panic 且无错误时不回滚
	if !*panicked && *err == nil {
		return
	}

	// 执行回滚
	if rbErr := rollbackFunc(); rbErr != nil {
		// 回滚失败日志
		if savePoint != "" {
			zap.L().Error("回滚到 SavePoint 失败", zap.String("SavePoint", savePoint), zap.Error(rbErr))
			return
		}

		zap.L().Error("回滚事务失败", zap.Error(rbErr))

		return
	}

	// 回滚成功
	if savePoint != "" {
		zap.L().Warn("已回滚到 SavePoint", zap.String("SavePoint", savePoint), zap.Bool("panicked", *panicked), zap.Error(*err))
		return
	}

	zap.L().Warn("事务已回滚", zap.Bool("panicked", *panicked), zap.Error(*err))
}

// InsertFromDetails 从结构体中插入数据, processField 用于处理字段的函数.
func InsertFromDetails(db *gorm.DB, fields reflect.Type, values reflect.Value, processField func(i int, field reflect.StructField, value reflect.Value) error) error {
	return Transaction(db, func(tx *gorm.DB) error {
		// 遍历字段
		for i := range fields.NumField() {
			field := fields.Field(i)
			value := values.Field(i)

			// 处理字段
			if err := processField(i, field, value); err != nil {
				return err
			}
		}

		return nil
	})
}

// GenDateSQLConditionYM 生成字段 field 关于年月形式日期 sql 查询条件为字符串.
// - field: 字段名
// - ym: 年月指针
// 示例:
//
//	ym := &YearMonth{Year: 2021, Month: 1}
//	condition := GenDateSQLConditionYM("created_at", ym)
//	// condition 结果: "created_at BETWEEN '2021-01-01 00:00:00' AND '2021-01-31 23:59:59'"
func GenDateSQLConditionYM(field string, ym *YearMonth) string {
	startDate, endDate := YearMonthSelect(ym)
	start := startDate.Format(time.DateTime)
	end := endDate.Add(-1 * time.Second).Format(time.DateTime)

	return fmt.Sprintf("%s BETWEEN '%s' AND '%s'", field, start, end)
}
