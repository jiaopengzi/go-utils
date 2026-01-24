//
// FilePath    : go-utils\amount_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 金额相关测试
//

package utils

import (
	"testing"
)

func TestNewConfig_DefaultValues(t *testing.T) {
	cfg := NewConfig()

	if cfg.Denominator != 100 {
		t.Errorf("期望 分母 为 %d, 实际为 %d", 100, cfg.Denominator)
	}
	if cfg.Numerator != 1 {
		t.Errorf("期望 分子 为 1, 实际为 %d", cfg.Numerator)
	}
	if cfg.Min != 1 {
		t.Errorf("期望 Min 为 1, 实际为 %d", cfg.Min)
	}
}

func TestNewConfig_WithDenominatorOption(t *testing.T) {
	cfg := NewConfig(WithDenominatorOption(1000))

	if cfg.Denominator != 1000 {
		t.Errorf("期望 分母 为 %d, 实际为 %d", 1000, cfg.Denominator)
	}
}

func TestNewConfig_WithNumeratorOption(t *testing.T) {
	cfg := NewConfig(WithNumeratorOption(5))

	if cfg.Numerator != 5 {
		t.Errorf("期望 分子 为 5, 实际为 %d", cfg.Numerator)
	}
}

func TestNewConfig_WithMinAmountOption(t *testing.T) {
	cfg := NewConfig(WithMinAmountOption(10))

	if cfg.Min != 10 {
		t.Errorf("期望 Min 为 10, 实际为 %d", cfg.Min)
	}
}

func TestNewConfig_WithMultipleOptions(t *testing.T) {
	cfg := NewConfig(
		WithDenominatorOption(10000),
		WithNumeratorOption(50),
		WithMinAmountOption(100),
	)

	if cfg.Denominator != 10000 {
		t.Errorf("期望 分母 为 %d, 实际为 %d", 10000, cfg.Denominator)
	}
	if cfg.Numerator != 50 {
		t.Errorf("期望 分子 为 50, 实际为 %d", cfg.Numerator)
	}
	if cfg.Min != 100 {
		t.Errorf("期望 Min 为 100, 实际为 %d", cfg.Min)
	}
}

func TestCalcMinAmountWithRatioAndFloor(t *testing.T) {
	tests := []struct {
		name    string
		amount  int64
		opts    []Option
		want    int64
		wantErr bool
	}{
		{
			name:    "默认配置, amount=1000",
			amount:  1000,
			opts:    nil,
			want:    10,
			wantErr: false,
		},
		{
			name:    "计算结果小于最小值时返回最小值",
			amount:  50,
			opts:    nil,
			want:    1,
			wantErr: false,
		},
		{
			name:    "使用千为基数",
			amount:  1000,
			opts:    []Option{WithDenominatorOption(1000)},
			want:    1,
			wantErr: false,
		},
		{
			name:    "自定义比率",
			amount:  1000,
			opts:    []Option{WithNumeratorOption(10)},
			want:    100,
			wantErr: false,
		},
		{
			name:    "自定义最小金额",
			amount:  100,
			opts:    []Option{WithMinAmountOption(50)},
			want:    50,
			wantErr: false,
		},
		{
			name:    "金额为零时返回最小值",
			amount:  0,
			opts:    nil,
			want:    1,
			wantErr: false,
		},
		{
			name:    "负数金额返回错误",
			amount:  -100,
			opts:    nil,
			want:    0,
			wantErr: true,
		},
		{
			name:    "无效配置：分母 为 0",
			amount:  100,
			opts:    []Option{WithDenominatorOption(0)},
			want:    0,
			wantErr: true,
		},
		{
			name:    "无效配置：分子 为负数",
			amount:  100,
			opts:    []Option{WithNumeratorOption(-1)},
			want:    0,
			wantErr: true,
		},
		{
			name:    "无效配置：Min 为负数",
			amount:  100,
			opts:    []Option{WithMinAmountOption(-1)},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalcMinAmountWithRatioAndFloor(tt.amount, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalcMinAmountWithRatioAndFloor() 返回错误 = %v, 期望错误: %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CalcMinAmountWithRatioAndFloor() = %v, 期望 = %v", got, tt.want)
			}
		})
	}
}
