//
// FilePath    : go-utils\logger\utils_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 日志工具单测
//

package logger

import (
	"fmt"

	"reflect"
	"testing"

	"github.com/jiaopengzi/go-utils"
	"github.com/stretchr/testify/require"
)

// TestStruct 测试结构体
type TestStruct struct {
	Username     string
	Password     string
	Email        string
	AccessToken  string
	Secret       string
	Profile      *Profile
	OtherDetails OtherDetails
}

// ObfuscatedUserRequest 模拟 garble 后字段名被改写, 但 JSON tag 保持稳定的请求结构.
type ObfuscatedUserRequest struct {
	FieldA string `json:"user_name"`
	FieldB string `json:"password"`
	FieldC string `json:"re_password"`
	FieldD string `json:"email"`
}

// Profile 嵌套结构体
type Profile struct {
	FirstName string
	LastName  string
}

// OtherDetails 另一个嵌套结构体
type OtherDetails struct {
	Address     string
	Phone       string
	SecretOther string
}

// TestDeepCopyAndMaskSensitiveFields 测试深拷贝和掩码敏感字段
func TestDeepCopyAndMaskSensitiveFields(t *testing.T) {

	// 创建一个测试用例
	testCase := struct {
		name     string
		input    *TestStruct
		expected *TestStruct
	}{
		name: "具有嵌套字段的基本结构",
		input: &TestStruct{
			Username:    "user1",
			Password:    "123456",
			Email:       "user1@example.com",
			AccessToken: "token",
			Secret:      "s3cr3t",
			Profile: &Profile{
				FirstName: "John",
				LastName:  "Doe",
			},
			OtherDetails: OtherDetails{
				Address:     "123 Main St",
				Phone:       "123-456-7890",
				SecretOther: "SecretOther",
			},
		},
		expected: &TestStruct{
			Username:    "user1",
			Password:    "******",
			Email:       "user1@example.com",
			AccessToken: "******",
			Secret:      "******",
			Profile: &Profile{
				FirstName: "John",
				LastName:  "Doe",
			},
			OtherDetails: OtherDetails{
				Address:     "123 Main St",
				Phone:       "123-456-7890",
				SecretOther: "******",
			},
		},
	}

	t.Run(testCase.name, func(t *testing.T) {
		// 深拷贝测试数据
		copiedData, err := utils.DeepCopy(testCase.input)
		if err != nil {
			fmt.Println("==>DeepCopy failed")
			return
		}

		// 移除敏感字段
		MaskSensitiveFields(copiedData)

		// 比较实际输出和期望输出
		if !reflect.DeepEqual(copiedData, testCase.expected) {
			t.Errorf("expected %+v, but got %+v", testCase.expected, copiedData)
		}
	})
}

// RedisNode 模拟 redis 节点配置(含敏感字段 Password)
type RedisNode struct {
	Host     string
	Port     int
	User     string
	Password string
	Database int
}

// StructWithSlicePointers 包含指针切片字段的结构体, 模拟 SetupRequest
type StructWithSlicePointers struct {
	PgsqlPassword string
	Redis         []*RedisNode
	ESPassword    string
}

// TestMaskSensitiveFields_SliceOfPointerStruct 测试切片中指针元素的敏感字段也能被掩码
func TestMaskSensitiveFields_SliceOfPointerStruct(t *testing.T) {
	input := &StructWithSlicePointers{
		PgsqlPassword: "pgsql_pwd",
		Redis: []*RedisNode{
			{Host: "10.10.10.1", Port: 7001, User: "default", Password: "redis_pwd_1", Database: 0},
			{Host: "10.10.10.2", Port: 7002, User: "default", Password: "redis_pwd_2", Database: 1},
		},
		ESPassword: "es_pwd",
	}

	expected := &StructWithSlicePointers{
		PgsqlPassword: "******",
		Redis: []*RedisNode{
			{Host: "10.10.10.1", Port: 7001, User: "default", Password: "******", Database: 0},
			{Host: "10.10.10.2", Port: 7002, User: "default", Password: "******", Database: 1},
		},
		ESPassword: "******",
	}

	// 深拷贝
	copied, err := utils.DeepCopy(input)
	if err != nil {
		t.Fatalf("DeepCopy failed: %v", err)
	}

	// 执行掩码
	MaskSensitiveFields(copied)

	// 校验
	if !reflect.DeepEqual(copied, expected) {
		t.Errorf("expected %+v, but got %+v", expected, copied)
		for i, node := range copied.Redis {
			t.Errorf("  Redis[%d]: %+v", i, node)
		}
	}
}

// PayConfigLike 模拟支付/社交登录配置, 字段名为 PascalCase 而敏感关键字为 snake_case
type PayConfigLike struct {
	AppID         string
	AppKey        string // 对应敏感关键字 app_key
	MchPrivateKey string // 对应敏感关键字 mch_private_key
	APIv3Key      string // 对应敏感关键字 api_v3_key
	AppPrivateKey string // 对应敏感关键字 app_private_key
	EncryptKey    string // 对应敏感关键字 encrypt_key
	NotifyHost    string
}

// TestMaskSensitiveFields_SnakeCaseKeywords 测试 snake_case 敏感关键字能匹配 PascalCase 结构体字段
func TestMaskSensitiveFields_SnakeCaseKeywords(t *testing.T) {
	// 设置包含下划线的敏感关键字
	SetSensitiveFields([]string{"password", "token", "secret", "app_key", "mch_private_key", "api_v3_key", "app_private_key", "encrypt_key"})
	defer SetSensitiveFields([]string{"password", "token", "secret", "captcha"}) // 恢复默认

	input := &PayConfigLike{
		AppID:         "2021005168672084",
		AppKey:        "c7caa9522f",
		MchPrivateKey: "-----BEGIN PRIVATE KEY-----\nMIIEvg...\n-----END PRIVATE KEY-----",
		APIv3Key:      "n6l81E6AB6LdVbrqFoc1",
		AppPrivateKey: "MIIEvwIBADsv7dESy5SmGxVk0cI07qhhq55d",
		EncryptKey:    "jempIHnilqQ==",
		NotifyHost:    "http://home.jiaopengzi.com:5426",
	}

	expected := &PayConfigLike{
		AppID:         "2021005168672084",
		AppKey:        "******",
		MchPrivateKey: "******",
		APIv3Key:      "******",
		AppPrivateKey: "******",
		EncryptKey:    "******",
		NotifyHost:    "http://home.jiaopengzi.com:5426",
	}

	// 深拷贝
	copied, err := utils.DeepCopy(input)
	if err != nil {
		t.Fatalf("DeepCopy failed: %v", err)
	}

	// 执行掩码
	MaskSensitiveFields(copied)

	// 校验
	if !reflect.DeepEqual(copied, expected) {
		t.Errorf("expected %+v, but got %+v", expected, copied)
	}
}

// TestSetSensitiveFields_ReplaceFields 测试设置自定义敏感关键字时, 会替换原有关键字列表.
func TestSetSensitiveFields_ReplaceFields(t *testing.T) {
	defer SetSensitiveFields([]string{"password", "token", "secret", "captcha"})

	type ReplaceFieldsConfig struct {
		Password string
		Captcha  string
		AppKey   string
	}

	input := &ReplaceFieldsConfig{
		Password: "password_value",
		Captcha:  "captcha_value",
		AppKey:   "app_key_value",
	}

	SetSensitiveFields([]string{"app_key"})
	MaskSensitiveFields(input)

	if input.Password != "password_value" {
		t.Fatalf("expected Password to remain unchanged, got %q", input.Password)
	}

	if input.Captcha != "captcha_value" {
		t.Fatalf("expected Captcha to remain unchanged, got %q", input.Captcha)
	}

	if input.AppKey != "******" {
		t.Fatalf("expected AppKey to be masked, got %q", input.AppKey)
	}

	fields := GetSensitiveFields()
	if !reflect.DeepEqual(fields, []string{"app_key"}) {
		t.Fatalf("expected sensitive fields to be replaced, got %+v", fields)
	}
}

// TestMaskSensitiveFields_UseJSONTag 测试敏感字段判断优先兼容 JSON tag, 避免 garble 混淆字段名后失效.
func TestMaskSensitiveFields_UseJSONTag(t *testing.T) {
	input := &ObfuscatedUserRequest{
		FieldA: "jiaopengzi",
		FieldB: "123QWEasd@",
		FieldC: "123QWEasd@",
		FieldD: "jiaopengzi@qq.com",
	}

	expected := &ObfuscatedUserRequest{
		FieldA: "jiaopengzi",
		FieldB: "******",
		FieldC: "******",
		FieldD: "jiaopengzi@qq.com",
	}

	copied, err := utils.DeepCopy(input)
	if err != nil {
		t.Fatalf("DeepCopy failed: %v", err)
	}

	MaskSensitiveFields(copied)

	if !reflect.DeepEqual(copied, expected) {
		t.Fatalf("expected %+v, but got %+v", expected, copied)
	}
}

// TestMaskSensitiveFields_MapStringKey 测试 map 字符串 key 也能按稳定业务标识掩码.
func TestMaskSensitiveFields_MapStringKey(t *testing.T) {
	input := map[string]any{
		"user_name":   "jiaopengzi",
		"password":    "123QWEasd@",
		"re_password": "123QWEasd@",
		"profile": map[string]any{
			"access_token": "token_value",
		},
	}

	MaskSensitiveFields(input)

	if input["password"] != "******" {
		t.Fatalf("input[password] = %v, want %q", input["password"], "******")
	}

	if input["re_password"] != "******" {
		t.Fatalf("input[re_password] = %v, want %q", input["re_password"], "******")
	}

	profile, ok := input["profile"].(map[string]any)
	if !ok {
		t.Fatalf("input[profile] type = %T, want map[string]any", input["profile"])
	}

	if profile["access_token"] != "******" {
		t.Fatalf("profile[access_token] = %v, want %q", profile["access_token"], "******")
	}
}

// BoolSensitiveStruct 模拟命中敏感关键字但值类型为 bool 的结构体.
type BoolSensitiveStruct struct {
	DeleteToken bool   `json:"delete_token"`
	AccessToken string `json:"access_token"`
}

// TestMaskSensitiveFields_NonStringSensitiveFieldDoesNotPanic 测试敏感字段值不是字符串时直接跳过, 不触发 panic.
func TestMaskSensitiveFields_NonStringSensitiveFieldDoesNotPanic(t *testing.T) {
	input := &BoolSensitiveStruct{
		DeleteToken: true,
		AccessToken: "token_value",
	}

	require.NotPanics(t, func() {
		MaskSensitiveFields(input)
	})

	if !input.DeleteToken {
		t.Fatalf("expected DeleteToken to remain true")
	}

	if input.AccessToken != "******" {
		t.Fatalf("expected AccessToken to be masked, got %q", input.AccessToken)
	}
}
