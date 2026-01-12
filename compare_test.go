//
// FilePath    : go-utils\compare_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 对比单元测试
//

package utils

import "testing"

type testStruct struct {
	Name     string
	Age      int
	Password string
}

func TestIsSlicesEqual(t *testing.T) {
	tests := []struct {
		name     string
		sliceSrc []testStruct
		sliceTar []testStruct
		want     bool
	}{
		{
			name:     "相等的切片",
			sliceSrc: []testStruct{{Name: "a", Age: 1, Password: "b"}, {Name: "c", Age: 2, Password: "d"}},
			sliceTar: []testStruct{{Name: "a", Age: 1, Password: "b"}, {Name: "c", Age: 2, Password: "d"}},
			want:     true,
		},
		{
			name:     "不同长度的切片",
			sliceSrc: []testStruct{{Name: "a", Age: 1, Password: "b"}, {Name: "c", Age: 2, Password: "d"}},
			sliceTar: []testStruct{{Name: "a", Age: 1, Password: "b"}},
			want:     false,
		},
		{
			name:     "不同元素的切片",
			sliceSrc: []testStruct{{Name: "a", Age: 1, Password: "b"}, {Name: "c", Age: 2, Password: "d"}},
			sliceTar: []testStruct{{Name: "a", Age: 1, Password: "b"}, {Name: "c", Age: 2, Password: "e"}},
			want:     false,
		},
		{
			name:     "空的切片",
			sliceSrc: []testStruct{},
			sliceTar: []testStruct{},
			want:     true,
		},
		{
			name:     "nil切片",
			sliceSrc: nil,
			sliceTar: nil,
			want:     true,
		},
		{
			name:     "一个nil切片",
			sliceSrc: nil,
			sliceTar: []testStruct{},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSlicesEqual(tt.sliceSrc, tt.sliceTar); got != tt.want {
				t.Errorf("IsSlicesEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSlicesEqualByField(t *testing.T) {
	tests := []struct {
		name       string
		sliceSrc   []testStruct
		sliceTar   []testStruct
		fieldNames []string
		want       bool
		wantErr    bool
	}{
		{
			name: "按Name字段比较相等的切片",
			sliceSrc: []testStruct{
				{Name: "a", Age: 1, Password: "b"},
				{Name: "c", Age: 2, Password: "d"},
			},
			sliceTar: []testStruct{
				{Name: "a", Age: 1, Password: "b"},
				{Name: "c", Age: 2, Password: "d"},
			},
			fieldNames: []string{"Name"},
			want:       true,
			wantErr:    false,
		},
		{
			name: "不同长度的切片",
			sliceSrc: []testStruct{
				{Name: "a", Age: 1, Password: "b"},
				{Name: "c", Age: 2, Password: "d"},
			},
			sliceTar: []testStruct{
				{Name: "a", Age: 1, Password: "b"},
			},
			fieldNames: []string{"Name"},
			want:       false,
			wantErr:    true,
		},
		{
			name: "按Name字段比较不同元素的切片",
			sliceSrc: []testStruct{
				{Name: "a", Age: 1, Password: "b"},
				{Name: "c", Age: 2, Password: "d"},
			},
			sliceTar: []testStruct{
				{Name: "a", Age: 1, Password: "b"},
				{Name: "x", Age: 2, Password: "d"},
			},
			fieldNames: []string{"Name"},
			want:       false,
			wantErr:    false,
		},
		{
			name:       "空的切片",
			sliceSrc:   []testStruct{},
			sliceTar:   []testStruct{},
			fieldNames: []string{"Name"},
			want:       true,
			wantErr:    false,
		},
		{
			name:       "nil切片",
			sliceSrc:   nil,
			sliceTar:   nil,
			fieldNames: []string{"Name"},
			want:       true,
			wantErr:    false,
		},
		{
			name:       "一个nil切片",
			sliceSrc:   nil,
			sliceTar:   []testStruct{},
			fieldNames: []string{"Name"},
			want:       false,
			wantErr:    true,
		},
		{
			name: "按Age字段比较不同元素的切片",
			sliceSrc: []testStruct{
				{Name: "a", Age: 1, Password: "b"},
				{Name: "c", Age: 2, Password: "d"},
			},
			sliceTar: []testStruct{
				{Name: "a", Age: 1, Password: "b"},
				{Name: "c", Age: 3, Password: "d"},
			},
			fieldNames: []string{"Age"},
			want:       false,
			wantErr:    false,
		},
		{
			name: "按Password字段比较不同元素的切片",
			sliceSrc: []testStruct{
				{Name: "a", Age: 1, Password: "b"},
				{Name: "c", Age: 2, Password: "d"},
			},
			sliceTar: []testStruct{
				{Name: "a", Age: 1, Password: "bb"},
				{Name: "c", Age: 2, Password: "dd"},
			},
			fieldNames: []string{"Password"},
			want:       false,
			wantErr:    false,
		},
		{
			name: "按多个字段比较相等的切片",
			sliceSrc: []testStruct{
				{Name: "a", Age: 1, Password: "b"},
				{Name: "c", Age: 2, Password: "d"},
			},
			sliceTar: []testStruct{
				{Name: "a", Age: 1, Password: "b"},
				{Name: "c", Age: 2, Password: "d"},
			},
			fieldNames: []string{"Name", "Age"},
			want:       true,
			wantErr:    false,
		},
		{
			name: "按多个字段比较不同元素的切片",
			sliceSrc: []testStruct{
				{Name: "a", Age: 1, Password: "b"},
				{Name: "c", Age: 2, Password: "d"},
			},
			sliceTar: []testStruct{
				{Name: "aa", Age: 1, Password: "b"},
				{Name: "c", Age: 3, Password: "d"},
			},
			fieldNames: []string{"Name", "Age"},
			want:       false,
			wantErr:    false,
		},

		{
			name: "按多个字段比较不同元素的切片",
			sliceSrc: []testStruct{
				{Name: "a", Age: 1, Password: "b"},
				{Name: "c", Age: 2, Password: "d"},
			},
			sliceTar: []testStruct{
				{Name: "aa", Age: 1, Password: "bb"},
				{Name: "c", Age: 3, Password: "dd"},
			},
			fieldNames: []string{"Name", "Age"},
			want:       false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsSlicesEqualByField(tt.sliceSrc, tt.sliceTar, tt.fieldNames)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsSlicesEqualByField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsSlicesEqualByField() = %v, want %v", got, tt.want)
			}
		})
	}
}
