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
		MaskSensitiveFields(copiedData, SensitiveFields)

		// 比较实际输出和期望输出
		if !reflect.DeepEqual(copiedData, testCase.expected) {
			t.Errorf("expected %+v, but got %+v", testCase.expected, copiedData)
		}
	})
}
