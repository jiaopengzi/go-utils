//
// FilePath    : go-utils\copy_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 深拷贝单元测试
//

package utils

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"
)

type Person struct {
	Name    string
	Age     int
	Address Address
}

type Address struct {
	City    string
	ZipCode string
	Details AddressDetails
}

type AddressDetails struct {
	Street string
	Number int
	Date   *sql.NullTime
}

func TestDeepCopy(t *testing.T) {
	// 固定的时间
	fixedTime := time.Date(2024, time.November, 30, 12, 0, 0, 0, time.UTC)

	original := &Person{
		Name: "jiaopengzi",
		Age:  18,
		Address: Address{
			City:    "成都",
			ZipCode: "610000",
			Details: AddressDetails{
				Street: "怡心湖",
				Number: 42,
				Date: &sql.NullTime{
					Time:  fixedTime,
					Valid: true,
				},
			},
		},
	}

	copied, err := DeepCopy(original)
	if err != nil {
		fmt.Println("==>DeepCopy failed")
		return
	}

	// 检查拷贝的数据是否与原始数据相等
	if !reflect.DeepEqual(original, copied) {
		t.Errorf("期望值 %v，实际值 %v", original, copied)
	}

	// 修改拷贝的数据，确保原始数据不受影响
	copied.Name = "鲍勃"
	copied.Address.City = "梦幻岛"
	copied.Address.Details.Street = "二街"

	if original.Name == copied.Name {
		t.Errorf("期望原始名字为 '焦棚子'，实际值 %v", original.Name)
	}

	if original.Address.City == copied.Address.City {
		t.Errorf("期望原始城市为 '成都'，实际值 %v", original.Address.City)
	}

	if original.Address.Details.Street == copied.Address.Details.Street {
		t.Errorf("期望原始街道为 '怡心湖'，实际值 %v", original.Address.Details.Street)
	}
}
