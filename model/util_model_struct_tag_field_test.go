//
// FilePath    : go-utils\model\util_model_struct_tag_field_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 单测
//

package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// BaseModelNoPrimarykeyTest 基础模型没有主键(需要自己定义主键)
type BaseModelNoPrimarykeyTest struct {
	CreatedAt time.Time      `gorm:"column:created_at_gorm;type:timestamp(6) with time zone;comment:创建时间" json:"created_at_json"`
	UpdatedAt time.Time      `gorm:"column:updated_at_gorm;type:timestamp(6) with time zone;comment:更新时间" json:"updated_at_json"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at_gorm;type:timestamp(6) with time zone;index;comment:删除时间" json:"deleted_at_json"`
}

func (BaseModelNoPrimarykeyTest) TableName() string {
	return "base_model_no_primarykey_test"
}

type BaseModelTest struct {
	// ID 自增，增加索引
	ID uint64 `gorm:"column:id_gorm;type:bigint;primarykey;index;autoIncrement:true;not null;comment:自增ID" json:"id_json,string"`
	BaseModelNoPrimarykeyTest
}

type TestModel struct {
	BaseModelTest
	Name string `gorm:"column:name_gorm;type:varchar(100)" json:"name_json"`
}

func (t TestModel) TableName() string {
	return "test_models"
}

func TestFindField(t *testing.T) {
	model := &TestModel{Name: "Test Name"}
	fieldPtr := &model.Name
	field, err := findField(model, fieldPtr)
	assert.NoError(t, err)
	assert.NotNil(t, field)
}

func TestFindFieldName(t *testing.T) {
	model := &TestModel{Name: "Test Name"}
	fieldPtr := &model.Name

	fieldName, err := findFieldName(model, fieldPtr)
	assert.NoError(t, err)
	assert.Equal(t, "Name", fieldName)
}

func TestGetFieldNameFromPtr(t *testing.T) {
	model := &TestModel{Name: "Test Name"}
	fieldPtr := &model.Name

	fieldName, err := GetFieldNameFromPtr(model, fieldPtr)
	assert.NoError(t, err)
	assert.Equal(t, "Name", fieldName)
}

func TestGetTagContent(t *testing.T) {
	model := &TestModel{BaseModelTest: BaseModelTest{ID: 1}, Name: "Test Name"}

	content, err := GetTagContent(model, &model.Name, "gorm", "column:", ";")
	assert.NoError(t, err)
	assert.Equal(t, "name_gorm", content)

	content, err = GetTagContent(model, &model.ID, "json", "", ",")
	assert.NoError(t, err)
	assert.Equal(t, "id_json", content)

	content, err = GetTagContent(model, &model.Name, "json", "", ",")
	assert.NoError(t, err)
	assert.Equal(t, "name_json", content)
}

func TestGetExportedFieldPtrs(t *testing.T) {
	model := &TestModel{Name: "Test Name"}

	fieldPtrs, err := getExportedFieldPtrs(model)
	assert.NoError(t, err)
	assert.Len(t, fieldPtrs, 5)
}

func TestGetColumnName(t *testing.T) {
	model := &TestModel{BaseModelTest: BaseModelTest{ID: 1}, Name: "Test Name"}

	columnName, err := GetColumnName(model, &model.Name, WithTableName(true))
	assert.NoError(t, err)
	assert.Equal(t, "test_models.name_gorm", columnName)

	columnName, err = GetColumnName(model, &model.Name, WithTableName(false))
	assert.NoError(t, err)
	assert.Equal(t, "name_gorm", columnName)

	columnName, err = GetColumnName(model, &model.Name)
	assert.NoError(t, err)
	assert.Equal(t, "name_gorm", columnName)

	columnName, err = GetColumnName(model, &model.Name, WithPrefix("my_custom_prefix"))
	assert.NoError(t, err)
	assert.Equal(t, "my_custom_prefix.name_gorm", columnName)

	// 测试 json tag
	columnName, err = GetColumnName(model, &model.ID, WithTag("json"))
	assert.NoError(t, err)
	assert.Equal(t, "id_json", columnName)

	columnName, err = GetColumnName(model, &model.Name, WithTag("json"))
	assert.NoError(t, err)
	assert.Equal(t, "name_json", columnName)

	columnName, err = GetColumnName(model, &model.CreatedAt, WithTag("json"))
	assert.NoError(t, err)
	assert.Equal(t, "created_at_json", columnName)
}

func TestGetColumnNames(t *testing.T) {
	model := &TestModel{Name: "Test Name"}
	fieldPtr1 := &model.Name
	fieldPtr2 := &model.DeletedAt

	columnNames, err := GetColumnNames(model, []any{fieldPtr1, fieldPtr2}, WithTableName(true))
	assert.NoError(t, err)
	assert.Equal(t, []string{"test_models.name_gorm", "test_models.deleted_at_gorm"}, columnNames)

	columnNames, err = GetColumnNames(model, []any{fieldPtr1, fieldPtr2}, WithTableName(false))
	assert.NoError(t, err)
	assert.Equal(t, []string{"name_gorm", "deleted_at_gorm"}, columnNames)

	columnNames, err = GetColumnNames(model, []any{fieldPtr1, fieldPtr2})
	assert.NoError(t, err)
	assert.Equal(t, []string{"name_gorm", "deleted_at_gorm"}, columnNames)

	columnNames, err = GetColumnNames(model, []any{fieldPtr1, fieldPtr2}, WithPrefix("my_custom_prefix"))
	assert.NoError(t, err)
	assert.Equal(t, []string{"my_custom_prefix.name_gorm", "my_custom_prefix.deleted_at_gorm"}, columnNames)
}

func TestGetAllColumnNames(t *testing.T) {
	model := &TestModel{Name: "Test Name"}

	columnNames, err := GetAllColumnNames(model, WithTableName(true))
	assert.NoError(t, err)
	assert.Equal(t, []string{"test_models.id_gorm", "test_models.created_at_gorm", "test_models.updated_at_gorm", "test_models.deleted_at_gorm", "test_models.name_gorm"}, columnNames)

	columnNames, err = GetAllColumnNames(model, WithTableName(false))
	assert.NoError(t, err)
	assert.Equal(t, []string{"id_gorm", "created_at_gorm", "updated_at_gorm", "deleted_at_gorm", "name_gorm"}, columnNames)

	columnNames, err = GetAllColumnNames(model)
	assert.NoError(t, err)
	assert.Equal(t, []string{"id_gorm", "created_at_gorm", "updated_at_gorm", "deleted_at_gorm", "name_gorm"}, columnNames)

	columnNames, err = GetAllColumnNames(model, WithPrefix("my_custom_prefix"))
	assert.NoError(t, err)
	assert.Equal(t, []string{"my_custom_prefix.id_gorm", "my_custom_prefix.created_at_gorm", "my_custom_prefix.updated_at_gorm", "my_custom_prefix.deleted_at_gorm", "my_custom_prefix.name_gorm"}, columnNames)

	// json
	columnNames, err = GetAllColumnNames(model, WithTag("json"))
	assert.NoError(t, err)
	assert.Equal(t, []string{"id_json", "created_at_json", "updated_at_json", "deleted_at_json", "name_json"}, columnNames)
}

func TestGetAllColumnNamesExcept(t *testing.T) {
	model := &TestModel{Name: "Test Name"}

	columnNames, err := GetAllColumnNamesExcept(model, []any{&model.ID, &model.CreatedAt, &model.UpdatedAt}, WithTableName(true))
	assert.NoError(t, err)
	assert.Equal(t, []string{"test_models.deleted_at_gorm", "test_models.name_gorm"}, columnNames)

	columnNames, err = GetAllColumnNamesExcept(model, []any{&model.ID, &model.CreatedAt, &model.UpdatedAt}, WithTableName(false))
	assert.NoError(t, err)
	assert.Equal(t, []string{"deleted_at_gorm", "name_gorm"}, columnNames)

	columnNames, err = GetAllColumnNamesExcept(model, []any{&model.ID, &model.CreatedAt, &model.UpdatedAt})
	assert.NoError(t, err)
	assert.Equal(t, []string{"deleted_at_gorm", "name_gorm"}, columnNames)

	columnNames, err = GetAllColumnNamesExcept(model, []any{&model.ID, &model.CreatedAt, &model.UpdatedAt}, WithPrefix("my_custom_prefix"))
	assert.NoError(t, err)
	assert.Equal(t, []string{"my_custom_prefix.deleted_at_gorm", "my_custom_prefix.name_gorm"}, columnNames)

	// json
	columnNames, err = GetAllColumnNamesExcept(model, []any{&model.ID, &model.CreatedAt, &model.UpdatedAt}, WithTag("json"))
	assert.NoError(t, err)
	assert.Equal(t, []string{"deleted_at_json", "name_json"}, columnNames)
}

func TestGetColumnNameType(t *testing.T) {
	model := &TestModel{Name: "Test Name"}
	fieldPtr := &model.Name

	tableField, err := GetColumnNameType(model, fieldPtr, WithTableName(true))
	assert.NoError(t, err)
	assert.Equal(t, TableField{Name: "test_models.name_gorm", Type: "varchar(100)"}, tableField)

	tableField, err = GetColumnNameType(model, fieldPtr, WithTableName(false))
	assert.NoError(t, err)
	assert.Equal(t, TableField{Name: "name_gorm", Type: "varchar(100)"}, tableField)

	tableField, err = GetColumnNameType(model, fieldPtr)
	assert.NoError(t, err)
	assert.Equal(t, TableField{Name: "name_gorm", Type: "varchar(100)"}, tableField)

	tableField, err = GetColumnNameType(model, fieldPtr, WithPrefix("my_custom_prefix"))
	assert.NoError(t, err)
	assert.Equal(t, TableField{Name: "my_custom_prefix.name_gorm", Type: "varchar(100)"}, tableField)

	// json不能用
}

func TestGetColumnNameTypes(t *testing.T) {
	model := &TestModel{Name: "Test Name"}
	fieldPtr1 := &model.Name
	fieldPtr2 := &model.DeletedAt

	tableFields, err := GetColumnNameTypes(model, []any{fieldPtr1, fieldPtr2}, WithTableName(true))
	assert.NoError(t, err)
	assert.Equal(t, []TableField{
		{Name: "test_models.name_gorm", Type: "varchar(100)"},
		{Name: "test_models.deleted_at_gorm", Type: "timestamp(6) with time zone"},
	}, tableFields)

	tableFields, err = GetColumnNameTypes(model, []any{fieldPtr1, fieldPtr2}, WithTableName(false))
	assert.NoError(t, err)
	assert.Equal(t, []TableField{
		{Name: "name_gorm", Type: "varchar(100)"},
		{Name: "deleted_at_gorm", Type: "timestamp(6) with time zone"},
	}, tableFields)

	tableFields, err = GetColumnNameTypes(model, []any{fieldPtr1, fieldPtr2})
	assert.NoError(t, err)
	assert.Equal(t, []TableField{
		{Name: "name_gorm", Type: "varchar(100)"},
		{Name: "deleted_at_gorm", Type: "timestamp(6) with time zone"},
	}, tableFields)

	tableFields, err = GetColumnNameTypes(model, []any{fieldPtr1, fieldPtr2}, WithPrefix("my_custom_prefix"))
	assert.NoError(t, err)
	assert.Equal(t, []TableField{
		{Name: "my_custom_prefix.name_gorm", Type: "varchar(100)"},
		{Name: "my_custom_prefix.deleted_at_gorm", Type: "timestamp(6) with time zone"},
	}, tableFields)

	// json不能用
}

func TestGetAllColumnNameTypes(t *testing.T) {
	model := &TestModel{Name: "Test Name"}

	tableFields, err := GetAllColumnNameTypes(model, WithTableName(true))
	assert.NoError(t, err)
	assert.Equal(t, []TableField{
		{Name: "test_models.id_gorm", Type: "bigint"},
		{Name: "test_models.created_at_gorm", Type: "timestamp(6) with time zone"},
		{Name: "test_models.updated_at_gorm", Type: "timestamp(6) with time zone"},
		{Name: "test_models.deleted_at_gorm", Type: "timestamp(6) with time zone"},
		{Name: "test_models.name_gorm", Type: "varchar(100)"},
	}, tableFields)

	tableFields, err = GetAllColumnNameTypes(model, WithTableName(false))
	assert.NoError(t, err)
	assert.Equal(t, []TableField{
		{Name: "id_gorm", Type: "bigint"},
		{Name: "created_at_gorm", Type: "timestamp(6) with time zone"},
		{Name: "updated_at_gorm", Type: "timestamp(6) with time zone"},
		{Name: "deleted_at_gorm", Type: "timestamp(6) with time zone"},
		{Name: "name_gorm", Type: "varchar(100)"},
	}, tableFields)

	tableFields, err = GetAllColumnNameTypes(model)
	assert.NoError(t, err)
	assert.Equal(t, []TableField{
		{Name: "id_gorm", Type: "bigint"},
		{Name: "created_at_gorm", Type: "timestamp(6) with time zone"},
		{Name: "updated_at_gorm", Type: "timestamp(6) with time zone"},
		{Name: "deleted_at_gorm", Type: "timestamp(6) with time zone"},
		{Name: "name_gorm", Type: "varchar(100)"},
	}, tableFields)

	tableFields, err = GetAllColumnNameTypes(model, WithPrefix("my_custom_prefix"))
	assert.NoError(t, err)
	assert.Equal(t, []TableField{
		{Name: "my_custom_prefix.id_gorm", Type: "bigint"},
		{Name: "my_custom_prefix.created_at_gorm", Type: "timestamp(6) with time zone"},
		{Name: "my_custom_prefix.updated_at_gorm", Type: "timestamp(6) with time zone"},
		{Name: "my_custom_prefix.deleted_at_gorm", Type: "timestamp(6) with time zone"},
		{Name: "my_custom_prefix.name_gorm", Type: "varchar(100)"},
	}, tableFields)
}

func TestGetAllColumnNameTypesExcept(t *testing.T) {

	model := &TestModel{Name: "Test Name"}

	tableFields, err := GetAllColumnNameTypesExcept(model, []any{&model.ID, &model.CreatedAt, &model.UpdatedAt}, WithTableName(true))
	assert.NoError(t, err)
	assert.Equal(t, []TableField{
		{Name: "test_models.deleted_at_gorm", Type: "timestamp(6) with time zone"},
		{Name: "test_models.name_gorm", Type: "varchar(100)"},
	}, tableFields)

	tableFields, err = GetAllColumnNameTypesExcept(model, []any{&model.ID, &model.CreatedAt, &model.UpdatedAt}, WithTableName(false))
	assert.NoError(t, err)
	assert.Equal(t, []TableField{
		{Name: "deleted_at_gorm", Type: "timestamp(6) with time zone"},
		{Name: "name_gorm", Type: "varchar(100)"},
	}, tableFields)

	tableFields, err = GetAllColumnNameTypesExcept(model, []any{&model.ID, &model.CreatedAt, &model.UpdatedAt})
	assert.NoError(t, err)
	assert.Equal(t, []TableField{
		{Name: "deleted_at_gorm", Type: "timestamp(6) with time zone"},
		{Name: "name_gorm", Type: "varchar(100)"},
	}, tableFields)

	tableFields, err = GetAllColumnNameTypesExcept(model, []any{&model.ID, &model.CreatedAt, &model.UpdatedAt}, WithPrefix("my_custom_prefix"))
	assert.NoError(t, err)
	assert.Equal(t, []TableField{
		{Name: "my_custom_prefix.deleted_at_gorm", Type: "timestamp(6) with time zone"},
		{Name: "my_custom_prefix.name_gorm", Type: "varchar(100)"},
	}, tableFields)

}

func TestDeleteAtIsNull(t *testing.T) {
	model := &TestModel{Name: "Test Name"}

	condition := DeleteAtIsNull(model, WithTableName(true))
	assert.Equal(t, "test_models.deleted_at IS NULL", condition)

	condition = DeleteAtIsNull(model, WithTableName(false))
	assert.Equal(t, "deleted_at IS NULL", condition)

	condition = DeleteAtIsNull(model)
	assert.Equal(t, "deleted_at IS NULL", condition)

	condition = DeleteAtIsNull(model, WithPrefix("my_custom_prefix"))
	assert.Equal(t, "my_custom_prefix.deleted_at IS NULL", condition)
}
