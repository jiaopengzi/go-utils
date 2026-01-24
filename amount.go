//
// FilePath    : go-utils\amount.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 金额相关
//

package utils

import (
	"errors"
	"fmt"
)

// Config 配置结构体
type Config struct {
	Numerator   int64 // 分子
	Denominator int64 // 分母
	Min         int64 // 最小金额(分)
}

// Option 包含 Apply 方法, 用来配置 Config
type Option interface {
	Apply(*Config)
}

// WithNumerator 设置比例的分子
type WithNumerator struct {
	Numerator int64
}

// Apply 实现 Option 接口的 Apply 方法
func (w WithNumerator) Apply(cfg *Config) {
	cfg.Numerator = w.Numerator
}

func WithNumeratorOption(numerator int64) Option {
	return WithNumerator{Numerator: numerator}
}

// WithDenominator 设置比例的分母
type WithDenominator struct {
	Denominator int64
}

// Apply 实现 Option 接口的 Apply 方法
func (w WithDenominator) Apply(cfg *Config) {
	cfg.Denominator = w.Denominator
}

func WithDenominatorOption(denominator int64) Option {
	return WithDenominator{Denominator: denominator}
}

// WithMinAmount 设置最小金额
type WithMinAmount struct {
	Min int64
}

// Apply 实现 Option 接口的 Apply 方法
func (w WithMinAmount) Apply(cfg *Config) {
	cfg.Min = w.Min
}

// WithMinAmountOption
func WithMinAmountOption(min int64) Option {
	return WithMinAmount{Min: min}
}

// NewConfig 创建配置对象
func NewConfig(opts ...Option) *Config {
	cfg := &Config{
		Denominator: 100, // 默认使用百分制
		Numerator:   1,   // 默认比例因子为 1, 即百分之 1 (1/100)
		Min:         1,   // 默认最小金额为 1 分
	}
	for _, opt := range opts {
		opt.Apply(cfg)
	}

	return cfg
}

// CalcMinAmountWithRatioAndFloor 计算最小金额, 基于比例和底线金额
//   - amount: 原始金额(分)
//   - opts: 可选参数, 用于配置计算的比例精度、分子因子和最小金额
//
// 返回计算后的金额(分), 如果计算结果小于底线金额则返回底线金额(分), 否则返回计算结果(分)
func CalcMinAmountWithRatioAndFloor(amount int64, opts ...Option) (int64, error) {
	cfg := NewConfig(opts...)

	// 配置检查
	if cfg.Denominator <= 0 || cfg.Numerator < 0 || cfg.Min < 0 { // 修改了字段名检查
		return 0, fmt.Errorf("invalid configuration: Denominator=%d, Numerator=%d, Min=%d", cfg.Denominator, cfg.Numerator, cfg.Min)
	}

	// 边界条件检查
	if amount < 0 {
		return 0, errors.New("amount cannot be negative")
	}

	// 计算金额: amount * (numerator / divisor) = (amount * numerator) / divisor
	calculatedAmount := amount * cfg.Numerator / cfg.Denominator

	// 当计算结果小于最小金额时, 返回最小金额
	if calculatedAmount < cfg.Min {
		return cfg.Min, nil
	}

	// 返回计算结果
	return calculatedAmount, nil
}
