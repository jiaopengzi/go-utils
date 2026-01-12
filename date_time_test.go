//
// FilePath    : go-utils\date_time_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : date time 单测
//

package utils

import (
	"testing"
	"time"
)

func TestYearMonthSelect(t *testing.T) {
	tests := []struct {
		name     string
		params   *YearMonth
		expected struct {
			startDate time.Time
			endDate   time.Time
		}
	}{
		{
			name:   "年月参数为空",
			params: &YearMonth{},
			expected: struct {
				startDate time.Time
				endDate   time.Time
			}{
				startDate: time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC),
				endDate:   time.Date(time.Now().Year()+1, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "只有年份参数",
			params: &YearMonth{
				Year: 2023,
			},
			expected: struct {
				startDate time.Time
				endDate   time.Time
			}{
				startDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				endDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "年月参数都有",
			params: &YearMonth{
				Year:  2023,
				Month: 5,
			},
			expected: struct {
				startDate time.Time
				endDate   time.Time
			}{
				startDate: time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
				endDate:   time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "年份参数都有且月份为12月",
			params: &YearMonth{
				Year:  2023,
				Month: 12,
			},
			expected: struct {
				startDate time.Time
				endDate   time.Time
			}{
				startDate: time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
				endDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startDate, endDate := YearMonthSelect(tt.params)
			if !startDate.Equal(tt.expected.startDate) {
				t.Errorf("expected startDate %v, got %v", tt.expected.startDate, startDate)
			}
			if !endDate.Equal(tt.expected.endDate) {
				t.Errorf("expected endDate %v, got %v", tt.expected.endDate, endDate)
			}
		})
	}
}
