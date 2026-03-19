//
// FilePath    : go-utils\dtovalidator\common_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 通用 DTO 校验器测试
//

package dtovalidator

import (
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/jiaopengzi/go-utils/types"
)

type mockFieldLevel struct {
	field any
}

func (m mockFieldLevel) Top() reflect.Value      { return reflect.ValueOf(struct{}{}) }
func (m mockFieldLevel) Parent() reflect.Value   { return reflect.ValueOf(struct{}{}) }
func (m mockFieldLevel) Field() reflect.Value    { return reflect.ValueOf(m.field) }
func (m mockFieldLevel) Param() string           { return "" }
func (m mockFieldLevel) GetTag() string          { return "" }
func (m mockFieldLevel) StructFieldName() string { return "" }
func (m mockFieldLevel) FieldName() string       { return "" }
func (m mockFieldLevel) GetStructFieldOK() (reflect.Value, reflect.Kind, bool) {
	return reflect.Value{}, reflect.Invalid, false
}
func (m mockFieldLevel) GetStructFieldOK2() (reflect.Value, reflect.Kind, bool, bool) {
	return reflect.Value{}, reflect.Invalid, false, false
}
func (m mockFieldLevel) ExtractType(field reflect.Value) (reflect.Value, reflect.Kind, bool) {
	return field, field.Kind(), false
}
func (m mockFieldLevel) GetStructFieldOKAdvanced(val reflect.Value, namespace string) (reflect.Value, reflect.Kind, bool) {
	return reflect.Value{}, reflect.Invalid, false
}
func (m mockFieldLevel) GetStructFieldOKAdvanced2(val reflect.Value, namespace string) (reflect.Value, reflect.Kind, bool, bool) {
	return reflect.Value{}, reflect.Invalid, false, false
}

var _ validator.FieldLevel = (*mockFieldLevel)(nil)

// TestValidateJSONInt64 校验 JSONInt64 指针兼容且仍要求正整数.
func TestValidateJSONInt64(t *testing.T) {
	value := types.JSONInt64(1)
	if !ValidateJSONInt64(mockFieldLevel{field: &value}) {
		t.Fatalf("ValidateJSONInt64() = false, want true")
	}

	zero := types.JSONInt64(0)
	if ValidateJSONInt64(mockFieldLevel{field: &zero}) {
		t.Fatalf("ValidateJSONInt64() = true, want false")
	}
}

// TestValidateJSONUint64 校验 JSONUint64 指针兼容且仍要求正整数.
func TestValidateJSONUint64(t *testing.T) {
	value := types.JSONUint64(1)
	if !ValidateJSONUint64(mockFieldLevel{field: &value}) {
		t.Fatalf("ValidateJSONUint64() = false, want true")
	}

	zero := types.JSONUint64(0)
	if ValidateJSONUint64(mockFieldLevel{field: &zero}) {
		t.Fatalf("ValidateJSONUint64() = true, want false")
	}
}

// TestValidateAndGetJSONUint64Slice 校验 JSONUint64Slice 指针支持.
func TestValidateAndGetJSONUint64Slice(t *testing.T) {
	values := types.JSONUint64Slice{1, 2, 3}
	actual, ok := ValidateAndGetJSONUint64Slice(mockFieldLevel{field: &values})

	if !ok {
		t.Fatalf("ValidateAndGetJSONUint64Slice() ok = false, want true")
	}

	if len(actual) != len(values) {
		t.Fatalf("ValidateAndGetJSONUint64Slice() len = %d, want %d", len(actual), len(values))
	}

	for i, value := range values {
		if actual[i] != value {
			t.Fatalf("ValidateAndGetJSONUint64Slice()[%d] = %v, want %d", i, actual[i], value)
		}
	}
}

// TestValidateAndGetJSONUint64SliceNilPointer 校验 nil 指针切片无效.
func TestValidateAndGetJSONUint64SliceNilPointer(t *testing.T) {
	var values *types.JSONUint64Slice
	if _, ok := ValidateAndGetJSONUint64Slice(mockFieldLevel{field: values}); ok {
		t.Fatalf("ValidateAndGetJSONUint64Slice() ok = true, want false")
	}
}

// TestValidateEnumString 校验 string 指针枚举兼容.
func TestValidateEnumString(t *testing.T) {
	value := "post"
	if !ValidateEnumString(mockFieldLevel{field: &value}, "post", "page") {
		t.Fatalf("ValidateEnumString() = false, want true")
	}
}

// TestValidateInt 校验 int 指针兼容且仍要求正整数.
func TestValidateInt(t *testing.T) {
	value := 1
	if !ValidateInt(mockFieldLevel{field: &value}) {
		t.Fatalf("ValidateInt() = false, want true")
	}

	zero := 0
	if ValidateInt(mockFieldLevel{field: &zero}) {
		t.Fatalf("ValidateInt() = true, want false")
	}
}

func TestStringHasPrefixIgnoreCase(t *testing.T) {
	cases := []struct {
		s, prefix string
		want      bool
	}{
		{"HelloWorld", "hello", true},
		{"HelloWorld", "HELLO", true},
		{"abc", "abcd", false},
		{"", "", true},
	}

	for _, c := range cases {
		if got := stringHasPrefixIgnoreCase(c.s, c.prefix); got != c.want {
			t.Fatalf("stringHasPrefixIgnoreCase(%q, %q) = %v; want %v", c.s, c.prefix, got, c.want)
		}
	}
}

func TestStringHasSuffixIgnoreCase(t *testing.T) {
	cases := []struct {
		s, suffix string
		want      bool
	}{
		{"HelloWorld", "world", true},
		{"HelloWorld", "WORLD", true},
		{"abc", "z", false},
		{"", "", true},
	}

	for _, c := range cases {
		if got := stringHasSuffixIgnoreCase(c.s, c.suffix); got != c.want {
			t.Fatalf("stringHasSuffixIgnoreCase(%q, %q) = %v; want %v", c.s, c.suffix, got, c.want)
		}
	}
}

func TestValidateCSR(t *testing.T) {
	v := validator.New()
	if err := v.RegisterValidation("ValidateCSR", ValidateCSR); err != nil {
		t.Fatalf("register validation failed: %v", err)
	}

	type S struct {
		CSR string `validate:"ValidateCSR"`
	}

	validCSR := `-----BEGIN CERTIFICATE REQUEST-----
MIG6MG4CAQAwFDESMBAGA1UEAxMJbG9jYWxob3N0MCowBQYDK2VwAyEAr2h/kLhK
6e0FsbWcOjyBYr6dewt95bS9TBZ95Dm9jTWgJzAlBgkqhkiG9w0BCQ4xGDAWMBQG
A1UdEQQNMAuCCWxvY2FsaG9zdDAFBgMrZXADQQDi/X6l3MkbAWkeYPSBjJGR/zxH
b0ywI0a+em51y5dgH/o6Ud052pGWysNVx7FRaLQoG/DQfT4ofSasRwqeIRoN
-----END CERTIFICATE REQUEST-----`

	s := S{CSR: validCSR}
	if err := v.Struct(s); err != nil {
		t.Fatalf("valid CSR flagged invalid: %v", err)
	}

	// missing header
	s2 := S{CSR: "some body\n-----END CERTIFICATE REQUEST-----"}
	if err := v.Struct(s2); err == nil {
		t.Fatalf("invalid CSR (missing header) was accepted")
	}

	// missing footer
	s3 := S{CSR: "-----BEGIN CERTIFICATE REQUEST-----\nsome body"}
	if err := v.Struct(s3); err == nil {
		t.Fatalf("invalid CSR (missing footer) was accepted")
	}
}
