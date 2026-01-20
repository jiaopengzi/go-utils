//
// FilePath    : go-utils\model\field_exists.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 字段存在性检查.
//

package model

import (
	"fmt"

	"gorm.io/gorm"
)

// CreateIfNotExists 如果记录不存在则创建
func CreateIfNotExists(db *gorm.DB, t Tabler, filedPointer any) error {
	name, err := GetColumnName(t, filedPointer) // 获取字段名
	if err != nil {
		return err
	}

	query := fmt.Sprintf("%s = ?", name)

	var count int64

	db.Table(t.TableName()).Where(query, filedPointer).Find(&t).Count(&count)

	if count == 0 {
		if createErr := db.Create(t).Error; createErr != nil {
			return fmt.Errorf("创建 %v 时发生错误：%v", filedPointer, createErr)
		}
	}

	return nil
}

// CheckFieldExist 检查字段数据是否存在
func CheckFieldExist(db *gorm.DB, table Tabler, fieldName string, fieldValue any) bool {
	query := fmt.Sprintf("%s = ?", fieldName)

	// 判断 fieldValue 是否为 nil
	if fieldValue == nil {
		query = fmt.Sprintf("%s IS NULL", fieldName)
	}

	var count int64

	// 使用 Limit
	db.Table(table.TableName()).
		Where(DeleteAtIsNull(table)).
		Where(query, fieldValue).Limit(1).Count(&count)

	return count > 0
}

// CheckFieldsExist 检查多字段数据是否存在
func CheckFieldsExist(db *gorm.DB, table Tabler, fields map[string]any) (bool, error) {
	query := db.Table(table.TableName()).Where(DeleteAtIsNull(table))

	for field, value := range fields {
		// 判断 value 是否为 nil
		if value == nil {
			query = query.Where(fmt.Sprintf("%s IS NULL", field))
			continue
		}

		query = query.Where(fmt.Sprintf("%s = ?", field), value)
	}

	var count int64

	// 使用 Limit
	err := query.Limit(1).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CheckFieldExistExcludingField 检查字段数据是否存在，排除指定字段
func CheckFieldExistExcludingField(db *gorm.DB, table Tabler, fieldName, excludeFieldName string, fieldValue, excludeFieldValue any) bool {
	// 查询条件
	query := fmt.Sprintf("%s = ?", fieldName)
	if fieldValue == nil {
		query = fmt.Sprintf("%s IS NULL", fieldName)
	}

	// 排除条件
	queryExclude := fmt.Sprintf("%s != ?", excludeFieldName)
	if fieldValue == nil {
		queryExclude = fmt.Sprintf("%s IS NOT NULL", excludeFieldName)
	}

	var count int64

	db.Table(table.TableName()).
		Where(DeleteAtIsNull(table)).
		Where(query, fieldValue).
		Where(queryExclude, excludeFieldValue).
		Limit(1).Count(&count)

	return count > 0
}

// CheckFiledMultiValueValidity 检查字段多个值的有效性
func CheckFiledMultiValueValidity(db *gorm.DB, table Tabler, fieldName string, fieldValues []any) (bool, error) {
	// 获取字段值的数量
	countSrc := len(fieldValues)

	query := db.Table(table.TableName()).
		Where(DeleteAtIsNull(table)).
		Where(fmt.Sprintf("%s IN (?)", fieldName), fieldValues)

	var count int64

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}

	// 判断字段值的数量是否相等
	return count == int64(countSrc), nil
}
