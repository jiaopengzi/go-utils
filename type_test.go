//
// FilePath    : go-utils\type_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 单测
//

package utils

import "testing"

func TestIsPointer(t *testing.T) {
	intVal := 42
	strVal := "hello"
	type myStruct struct{ A int }
	sVal := myStruct{A: 1}

	cases := []struct {
		name string
		v    any
		want bool
	}{
		{"int value", intVal, false},
		{"int pointer", &intVal, true},
		{"string value", strVal, false},
		{"string pointer", &strVal, true},
		{"struct value", sVal, false},
		{"struct pointer", &sVal, true},
		{"interface holding pointer", any(&sVal), true},
		{"interface holding value", any(sVal), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsPointer(tc.v)
			if got != tc.want {
				t.Fatalf("IsPointer(%T) = %v; want %v", tc.v, got, tc.want)
			}
		})
	}
}
