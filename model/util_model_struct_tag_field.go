//
// FilePath    : go-utils\model\util_model_struct_tag_field.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 模型结构体标签字段工具, 获取结构体标签字段.
//

package model

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/jiaopengzi/go-utils"
	"gorm.io/gorm"
)

// skipFieldValidation 默认为 false, 即开发模式, 执行验证
var skipFieldValidation bool

// SetSkipFieldValidation 设置是否跳过字段验证, 生产环境建议设置为 true 以提升性能
// 因为在开发环境为了代码编译和书写不出错才验证, 所以在生产环境不需要再次验证, 减少不必要的性能浪费。
func SetSkipFieldValidation(skip bool) {
	skipFieldValidation = skip
}

// 支持的标签名
const (
	gormTag = "gorm"
	jsonTag = "json"
)

// columnCache 使用 sync.Map 作为缓存列名称的缓存, 保证并发安全.
var columnCache sync.Map

// TableField 表格字段
type TableField struct {
	Name string // 字段名
	Type string // 字段类型
}

// Config 是否增加表名
type Config struct {
	TableName bool   // 是否增加默认表名
	Prefix    string // 自定义前缀
	Tag       string // 标签名
}

// Option 包含 Apply 方法，用来配置 Config
type Option interface {
	Apply(*Config)
}

// Validate 验证配置是否正确
func (cfg *Config) Validate() error {
	// 表名前缀和默认表名是互斥的
	if cfg.Prefix != "" && cfg.TableName {
		return fmt.Errorf("WithTableName and WithPrefix are mutually exclusive")
	}

	return nil
}

// TableNameOption 是否增加表名
type TableNameOption struct {
	TableName bool
}

// Apply 方法实现了 Option 接口，用来设置 Config 的 TableName
func (o TableNameOption) Apply(cfg *Config) {
	cfg.TableName = o.TableName
}

// WithTableName 增加表名可选参数
func WithTableName(flag bool) Option {
	return TableNameOption{TableName: flag}
}

// PrefixOption 是否增加自定义前缀
type PrefixOption struct {
	Prefix string
}

// Apply 方法实现了 Option 接口，用来设置 Config 的 Prefix
func (o PrefixOption) Apply(cfg *Config) {
	cfg.Prefix = o.Prefix
}

// WithPrefix 增加自定义前缀可选参数
func WithPrefix(s string) Option {
	return PrefixOption{Prefix: s}
}

// TagOption 是否指定标签名, 默认为 gorm
type TagOption struct {
	Tag string
}

// Apply 方法实现了 Option 接口，用来设置 Config 的 Tag
func (o TagOption) Apply(cfg *Config) {
	cfg.Tag = o.Tag
}

// WithTag 增加标签名可选参数
func WithTag(tag string) Option {
	return TagOption{Tag: tag}
}

// isFieldOf 判断 fieldPtr 是否是 structPtr 的字段, 如果是返回 true, 否则返回 false,注意 structPtr 和 fieldPtr 必须是指针
func isFieldOf(structPtr any, fieldPtr any) bool {
	// 获取 modelTar 的值
	modelValue := reflect.ValueOf(structPtr).Elem()

	// 确保 modelValue 是一个结构体
	if modelValue.Kind() != reflect.Struct {
		return false
	}

	for i := range modelValue.NumField() {
		field := modelValue.Field(i)
		fieldType := modelValue.Type().Field(i)

		// 跳过不可导出的字段
		if fieldType.PkgPath != "" || !field.CanInterface() {
			continue
		}

		// 当前字段的指针
		iFieldPtr := field.Addr().Interface()

		if iFieldPtr == fieldPtr {
			return true
		}

		// 递归检查,处理嵌套结构体
		if field.Kind() == reflect.Struct {
			if isFieldOf(iFieldPtr, fieldPtr) {
				return true
			}
		}
	}

	return false
}

// isFieldInModelAndIsPtr 检查 modelTar 和 fieldPtr 是否是指针，并且 fieldPtr 是否是 modelTar 的字段
func isFieldInModelAndIsPtr(modelTar any, fieldPtr any) (bool, error) {
	// 如果是生产环境，直接返回 true
	// 因为在开发环境为了代码编译和书写不出错才验证, 所以在生产环境不需要再次验证, 减少不必要的性能浪费。
	if skipFieldValidation {
		return true, nil
	}

	// 检查 modelTar 是否是指针
	if !utils.IsPointer(modelTar) {
		return false, fmt.Errorf("modelTar %T must be a pointer, with value %v", modelTar, modelTar)
	}

	// 检查 fieldPtr 是否是指针
	if !utils.IsPointer(fieldPtr) {
		return false, fmt.Errorf("fieldPtr %T must be a pointer, with value %v", fieldPtr, fieldPtr)
	}

	// 检查 fieldPtr 是否是 modelTar 的字段
	if !isFieldOf(modelTar, fieldPtr) {
		return false, fmt.Errorf("fieldPtr:%T is not a field of modelTar:%T", fieldPtr, modelTar)
	}

	return true, nil
}

// getExportedFieldPtrs 遍历结构体 modelTar 及其嵌套结构体，返回所有可导出字段的指针切片
func getExportedFieldPtrs(modelTar any) ([]any, error) {
	modelValue := reflect.ValueOf(modelTar).Elem() // 获取指针指向的值
	modelType := modelValue.Type()                 // 获取值的类型

	fieldPtrs := make([]any, 0, modelValue.NumField())

	// 定义需要排除递归处理的类型集合
	excludedTypes := map[reflect.Type]struct{}{
		reflect.TypeFor[time.Time]():      {},
		reflect.TypeFor[gorm.DeletedAt](): {},
		reflect.TypeFor[sql.NullTime]():   {},
	}

	// 遍历结构体的所有字段
	for i := range modelValue.NumField() {
		field := modelValue.Field(i)
		fieldType := modelType.Field(i)

		// 跳过不可导出的字段
		if fieldType.PkgPath != "" || !field.CanInterface() {
			continue
		}

		iFieldPtr := field.Addr().Interface()

		// 字段的种类是结构体并且字段的类型不在排除集合中才递归处理嵌套结构体
		if field.Kind() == reflect.Struct {
			if _, excluded := excludedTypes[field.Type()]; !excluded {
				nestedPtrs, err := getExportedFieldPtrs(iFieldPtr)
				if err != nil {
					return nil, err
				}

				fieldPtrs = append(fieldPtrs, nestedPtrs...)

				continue
			}
		}

		// 添加字段的指针到结果集合中
		fieldPtrs = append(fieldPtrs, iFieldPtr)
	}

	return fieldPtrs, nil
}

// findField 在 structPtr 中递归查找 fieldPtr，如果找到则返回字段信息，否则返回错误
func findField(structPtr any, fieldPtr any) (*reflect.StructField, error) {
	structVal := reflect.ValueOf(structPtr).Elem()
	for i := range structVal.NumField() {
		field := structVal.Field(i)
		fieldType := structVal.Type().Field(i)

		// 跳过不可导出的字段
		if fieldType.PkgPath != "" || !field.CanInterface() {
			continue
		}

		// 当前字段的指针
		iFieldPtr := field.Addr().Interface()

		if iFieldPtr == fieldPtr {
			return &fieldType, nil
		}

		// 递归处理嵌套结构体
		if field.Kind() == reflect.Struct {
			f, err := findField(iFieldPtr, fieldPtr)
			if err == nil {
				return f, nil
			}
		}
	}

	return nil, fmt.Errorf("无法找到字段")
}

// findFieldName 在 structPtr 中递归查找 fieldPtr，如果找到则返回字段名，否则返回错误
func findFieldName(structPtr any, fieldPtr any) (string, error) {
	f, err := findField(structPtr, fieldPtr)
	if err != nil {
		return "", err
	}

	return f.Name, nil
}

// GetFieldNameFromPtr 在 structPtr 中查找 fieldPtr，如果找到则返回字段名，否则返回错误, structPtr 和 fieldPtr 必须是指针
func GetFieldNameFromPtr(structPtr any, fieldPtr any) (string, error) {
	// 检查 modelTar 和 fieldPtr 是否是指针，并且 fieldPtr 是否是 modelTar 的字段
	if ok, err := isFieldInModelAndIsPtr(structPtr, fieldPtr); !ok {
		return "", err
	}

	return findFieldName(structPtr, fieldPtr)
}

// GetTagContent 函数用于解析结构体标签内容并返回,例如解析 gorm 标签的 column 键的内容
//   - structPtr: 需要解析的结构体指针
//   - fieldPtr: 需要解析的字段指针
//   - tag: 需要解析的标签名，例如 "gorm" 或 "json"
//   - key: 需要获取的键名，例如 "column"
//   - separator: 分隔符，例如 ";"
func GetTagContent(structPtr, fieldPtr any, tag, key, separator string) (string, error) {
	// 首先判断 tag 是否在指定的范围内
	if tag != gormTag && tag != jsonTag {
		return "", fmt.Errorf("不支持的标签 '%s', 只支持 %s 和 %s 的 tag", tag, gormTag, jsonTag)
	}

	// 从结构体指针中获取字段信息
	field, err := findField(structPtr, fieldPtr)
	if err != nil {
		return "", err
	}

	// 从字段信息中获取标签内容
	contentStr := field.Tag.Get(tag)

	// 处理 json 标签
	if tag == jsonTag {
		// 使用逗号分割标签内容
		parts := strings.Split(contentStr, separator)
		if len(parts) > 0 {
			return parts[0], nil
		}

		return "", fmt.Errorf("无法找到 '%s' 标签的内容", tag)
	}

	// 使用分隔符将标签内容分割成多个部分
	parts := strings.SplitSeq(contentStr, separator)

	// 遍历每个部分
	for part := range parts {
		// 如果开头是我们需要的键名
		if strings.HasPrefix(part, key) {
			// 使用键名将部分分割成两个部分
			keyValue := strings.SplitN(part, key, 2)

			// 如果分割后的部分长度为 2 ，说明找到了需要的内容
			if len(keyValue) == 2 {
				return strings.TrimSpace(keyValue[1]), nil // 去除空格后返回
			}
		}
	}

	// 如果函数没有提前返回，说明没有找到需要的内容，返回错误
	return "", fmt.Errorf("无法找到 '%s' 标签的 '%s' 键", tag, key)
}

// GetColumnName 获取结构体 modelTar 中字段 fieldPtr 的 tag column 内容。
//   - modelTar: 表模型指针
//   - fieldPtr: 字段指针
//   - opts:可选参数默认为空,表示不添加前缀
//   - 可选参数 WithTableName(true) 使用 tabler.TableName() 作为前缀
//   - 可选参数 WithTableName(false) 不添加前缀
//   - 可选参数 WithPrefix("my_custom_prefix") 使用自定义前缀
func GetColumnName(modelTar Tabler, fieldPtr any, opts ...Option) (string, error) {
	cfg := Config{
		Tag: "gorm", // 默认为 gorm 标签
	}

	// 应用选项
	for _, opt := range opts {
		opt.Apply(&cfg)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return "", err
	}

	// 获取模型类型和字段名称
	fieldName, err := GetFieldNameFromPtr(modelTar, fieldPtr)
	if err != nil {
		return "", err
	}

	// 生成缓存键
	cacheKey := fmt.Sprintf("%s.%s.%s.%t.%s", modelTar.TableName(), fieldName, cfg.Prefix, cfg.TableName, cfg.Tag)

	// 尝试从缓存中获取列名称
	if cachedColumnName, ok := columnCache.Load(cacheKey); ok {
		colName, ok := cachedColumnName.(string)
		if ok {
			return colName, nil
		}

		return "", fmt.Errorf("无法从缓存中获取列名称")
	}

	var ColumnName string

	switch cfg.Tag {
	// 处理 gorm 标签
	case gormTag:
		ColumnName, err = GetTagContent(modelTar, fieldPtr, gormTag, "column:", ";")
		if err != nil {
			return "", err
		}

	// 处理 json 标签
	case jsonTag:
		ColumnName, err = GetTagContent(modelTar, fieldPtr, jsonTag, "", ",")
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("不支持的标签 '%s', 只支持 %s 和 %s 的 tag", cfg.Tag, gormTag, jsonTag)
	}

	if cfg.Prefix != "" {
		ColumnName = fmt.Sprintf("%s.%s", cfg.Prefix, ColumnName)
	} else if cfg.TableName {
		ColumnName = fmt.Sprintf("%s.%s", modelTar.TableName(), ColumnName)
	}

	// 存入缓存
	columnCache.Store(cacheKey, ColumnName)

	return ColumnName, nil
}

// GetColumnNames 获取结构体 modelTar 中多个字段 fieldPtrs 的 tag column 内容。
//   - modelTar: 表模型指针
//   - fieldPtrs: 字段指针切片
//   - opts:可选参数默认为空,表示不添加前缀
//   - 可选参数 WithTableName(true) 使用 tabler.TableName() 作为前缀
//   - 可选参数 WithTableName(false) 不添加前缀
//   - 可选参数 WithPrefix("my_custom_prefix") 使用自定义前缀
func GetColumnNames(modelTar Tabler, fieldPtrs []any, opts ...Option) ([]string, error) {
	cfg := Config{
		Tag: "gorm", // 默认为 gorm 标签
	}

	// 应用选项
	for _, opt := range opts {
		opt.Apply(&cfg)
	}
	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	var columnNames []string

	var columnName string

	var err error
	for _, fieldPtr := range fieldPtrs {
		columnName, err = GetColumnName(modelTar, fieldPtr, WithTableName(cfg.TableName), WithPrefix(cfg.Prefix), WithTag(cfg.Tag))
		if err != nil {
			return nil, err
		}

		columnNames = append(columnNames, columnName)
	}

	return columnNames, nil
}

// GetAllColumnNames 获取结构体 modelTar 中所有字段的 tag column 内容
//   - modelTar: 表模型指针
//   - opts:可选参数默认为空,表示不添加前缀
//   - 可选参数 WithTableName(true) 使用 tabler.TableName() 作为前缀
//   - 可选参数 WithTableName(false) 不添加前缀
//   - 可选参数 WithPrefix("my_custom_prefix") 使用自定义前缀
func GetAllColumnNames(modelTar Tabler, opts ...Option) ([]string, error) {
	cfg := Config{
		Tag: "gorm", // 默认为 gorm 标签
	}

	// 应用选项
	for _, opt := range opts {
		opt.Apply(&cfg)
	}
	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// 获取所有字段的指针
	fieldPtrs, err := getExportedFieldPtrs(modelTar)
	if err != nil {
		return nil, err
	}

	// 获取所有字段的列名
	return GetColumnNames(modelTar, fieldPtrs, WithTableName(cfg.TableName), WithPrefix(cfg.Prefix), WithTag(cfg.Tag))
}

// GetAllColumnNamesExcept 获取结构体中 modelTar 除了指定字段 exceptFieldPtrs 之外的所有字段的 tag column 内容
//   - modelTar: 表模型指针
//   - exceptFieldPtrs: 需要排除的字段指针列表
//   - opts:可选参数默认为空,表示不添加前缀
//   - 可选参数 WithTableName(true) 使用 tabler.TableName() 作为前缀
//   - 可选参数 WithTableName(false) 不添加前缀
//   - 可选参数 WithPrefix("my_custom_prefix") 使用自定义前缀
func GetAllColumnNamesExcept(modelTar Tabler, exceptFieldPtrs []any, opts ...Option) ([]string, error) {
	cfg := Config{
		Tag: "gorm", // 默认为 gorm 标签
	}

	// 应用选项
	for _, opt := range opts {
		opt.Apply(&cfg)
	}
	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// 获取所有字段的指针
	fieldPtrs, err := getExportedFieldPtrs(modelTar)
	if err != nil {
		return nil, err
	}

	// 获取需要排除的字段名
	exceptColMap := make(map[string]struct{}, len(exceptFieldPtrs))

	for _, exceptFieldPtr := range exceptFieldPtrs {
		columnName, err := GetColumnName(modelTar, exceptFieldPtr, WithTableName(cfg.TableName), WithPrefix(cfg.Prefix), WithTag(cfg.Tag))
		if err != nil {
			return nil, err
		}

		exceptColMap[columnName] = struct{}{}
	}

	// 获取所有字段的列名
	var columnNames []string

	for _, fieldPtr := range fieldPtrs {
		columnName, err := GetColumnName(modelTar, fieldPtr, WithTableName(cfg.TableName), WithPrefix(cfg.Prefix), WithTag(cfg.Tag))
		if err != nil {
			return nil, err
		}

		// 如果字段名不在排除列表中，则添加到结果中
		if _, ok := exceptColMap[columnName]; !ok {
			columnNames = append(columnNames, columnName)
		}
	}

	return columnNames, nil
}

// GetColumnNameType 获取结构体 modelTar 中字段 fieldPtr 的 tag column 和 type 内容
//   - modelTar: 表模型指针
//   - fieldPtr: 字段指针
//   - opts:可选参数默认为空,表示不添加前缀
//   - 可选参数 WithTableName(true) 使用 tabler.TableName() 作为前缀
//   - 可选参数 WithTableName(false) 不添加前缀
func GetColumnNameType(modelTar Tabler, fieldPtr any, opts ...Option) (TableField, error) {
	cfg := Config{
		Tag: "gorm", // 默认为 gorm 标签
	}

	// 应用选项
	for _, opt := range opts {
		opt.Apply(&cfg)
	}
	// 验证配置
	if err := cfg.Validate(); err != nil {
		return TableField{}, err
	}

	var tableField TableField

	// 获取模型类型和字段名称
	fieldName, err := GetFieldNameFromPtr(modelTar, fieldPtr)
	if err != nil {
		return tableField, err
	}

	// 生成缓存键
	cacheKey := fmt.Sprintf("%s.%s.%s.%t.type.%s", modelTar.TableName(), fieldName, cfg.Prefix, cfg.TableName, cfg.Tag)

	// 尝试从缓存中获取 TableField
	if cachedTableField, ok := columnCache.Load(cacheKey); ok {
		tableField, ok = cachedTableField.(TableField)
		if ok {
			return tableField, nil
		}

		return tableField, fmt.Errorf("无法从缓存中获取 TableField")
	}

	// 获取 tag column type 内容
	ColumnName, err := GetTagContent(modelTar, fieldPtr, "gorm", "column:", ";")
	if err != nil {
		return tableField, err
	}

	ColumnType, err := GetTagContent(modelTar, fieldPtr, "gorm", "type:", ";")
	if err != nil {
		return tableField, err
	}

	if cfg.Prefix != "" {
		ColumnName = fmt.Sprintf("%s.%s", cfg.Prefix, ColumnName)
	} else if cfg.TableName {
		ColumnName = fmt.Sprintf("%s.%s", modelTar.TableName(), ColumnName)
	}

	// 赋值字段名和字段类型
	tableField.Name = ColumnName
	tableField.Type = ColumnType

	// 存入缓存
	columnCache.Store(cacheKey, tableField)

	return tableField, nil
}

// GetColumnNameTypes 获取结构体中 modelTar 多个字段 fieldPtrs 的 tag column 和 type 内容
//   - modelTar: 表模型指针
//   - fieldPtrs: 字段指针切片
//   - opts:可选参数默认为空,表示不添加前缀
//   - 可选参数 WithTableName(true) 使用 tabler.TableName() 作为前缀
//   - 可选参数 WithTableName(false) 不添加前缀
//   - 可选参数 WithPrefix("my_custom_prefix") 使用自定义前缀
func GetColumnNameTypes(modelTar Tabler, fieldPtrs []any, opts ...Option) ([]TableField, error) {
	cfg := Config{
		Tag: "gorm", // 默认为 gorm 标签
	}

	// 应用选项
	for _, opt := range opts {
		opt.Apply(&cfg)
	}
	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	var tableFields []TableField

	for _, fieldPtr := range fieldPtrs {
		tableField, err := GetColumnNameType(modelTar, fieldPtr, WithTableName(cfg.TableName), WithPrefix(cfg.Prefix))
		if err != nil {
			return nil, err
		}

		tableFields = append(tableFields, tableField)
	}

	return tableFields, nil
}

// GetAllColumnNameTypes 获取结构体 modelTar 中所有字段的 tag column 和 type 内容
//   - modelTar: 表模型指针
//   - opts:可选参数默认为空,表示不添加前缀
//   - 可选参数 WithTableName(true) 使用 tabler.TableName() 作为前缀
//   - 可选参数 WithTableName(false) 不添加前缀
//   - 可选参数 WithPrefix("my_custom_prefix") 使用自定义前缀
func GetAllColumnNameTypes(modelTar Tabler, opts ...Option) ([]TableField, error) {
	cfg := Config{
		Tag: "gorm", // 默认为 gorm 标签
	}

	// 应用选项
	for _, opt := range opts {
		opt.Apply(&cfg)
	}
	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// 获取所有字段的指针
	fieldPtrs, err := getExportedFieldPtrs(modelTar)
	if err != nil {
		return nil, err
	}

	// 获取所有字段的列名和类型
	return GetColumnNameTypes(modelTar, fieldPtrs, WithTableName(cfg.TableName), WithPrefix(cfg.Prefix))
}

// GetAllColumnNameTypesExcept 获取结构体中 modelTar 除了指定字段 exceptFieldPtrs 之外的所有字段的 tag column 和 type 内容
//   - modelTar: 表模型指针
//   - exceptFieldPtrs: 需要排除的字段指针列表
//   - opts:可选参数默认为空,表示不添加前缀
//   - 可选参数 WithTableName(true) 使用 tabler.TableName() 作为前缀
//   - 可选参数 WithTableName(false) 不添加前缀
//   - 可选参数 WithPrefix("my_custom_prefix") 使用自定义前缀
func GetAllColumnNameTypesExcept(modelTar Tabler, exceptFieldPtrs []any, opts ...Option) ([]TableField, error) {
	cfg := Config{
		Tag: "gorm", // 默认为 gorm 标签
	}

	// 应用选项
	for _, opt := range opts {
		opt.Apply(&cfg)
	}
	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// 获取所有字段的指针
	fieldPtrs, err := getExportedFieldPtrs(modelTar)
	if err != nil {
		return nil, err
	}

	// 获取需要排除的字段名
	exceptColMap := make(map[string]string, len(exceptFieldPtrs))

	for _, exceptFieldPtr := range exceptFieldPtrs {
		columnNameType, err := GetColumnNameType(modelTar, exceptFieldPtr, WithTableName(cfg.TableName), WithPrefix(cfg.Prefix))
		if err != nil {
			return nil, err
		}

		exceptColMap[columnNameType.Name] = columnNameType.Type
	}

	// 获取所有字段的列名和类型
	var tableFields []TableField

	for _, fieldPtr := range fieldPtrs {
		columnNameType, err := GetColumnNameType(modelTar, fieldPtr, WithTableName(cfg.TableName), WithPrefix(cfg.Prefix))
		if err != nil {
			return nil, err
		}

		// 如果字段名不在排除列表中，则添加到结果中
		if _, ok := exceptColMap[columnNameType.Name]; !ok {
			tableFields = append(tableFields, columnNameType)
		}
	}

	return tableFields, nil
}

// DeleteAtIsNull 生成删除时间字段为空的查询条件, 用于软删除,即查询未删除的数据
//   - modelTar: 表模型指针
//   - opts:可选参数默认为空,表示不添加前缀
//   - 可选参数 WithTableName(true) 使用 tabler.TableName() 作为前缀
//   - 可选参数 WithTableName(false) 不添加前缀
//   - 可选参数 WithPrefix("my_custom_prefix") 使用自定义前缀
func DeleteAtIsNull(modelTar Tabler, opts ...Option) string {
	cfg := Config{
		Tag: "gorm", // 默认为 gorm 标签
	}

	// 应用选项
	for _, opt := range opts {
		opt.Apply(&cfg)
	}
	// 验证配置
	if err := cfg.Validate(); err != nil {
		return ""
	}

	// 生成缓存键
	tableName := modelTar.TableName()
	cacheKey := fmt.Sprintf("%s.DeleteAtIsNull.%s.%t", tableName, cfg.Prefix, cfg.TableName)

	// 尝试从缓存中获取条件语句
	if cachedCondition, ok := columnCache.Load(cacheKey); ok {
		condition, ok := cachedCondition.(string)
		if ok {
			return condition
		}

		return ""
	}

	// 基础模型
	var baseModel BaseModelNoPrimarykey

	// 获取删除时间字段
	deleteAtColName, err := GetColumnName(&baseModel, &baseModel.DeletedAt)
	if err != nil {
		return ""
	}

	if cfg.Prefix != "" {
		deleteAtColName = fmt.Sprintf("%s.%s", cfg.Prefix, deleteAtColName)
	} else if cfg.TableName {
		deleteAtColName = fmt.Sprintf("%s.%s", modelTar.TableName(), deleteAtColName)
	}

	condition := fmt.Sprintf("%s IS NULL", deleteAtColName)

	// 存入缓存
	columnCache.Store(cacheKey, condition)

	return condition
}
