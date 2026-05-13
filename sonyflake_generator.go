//
// FilePath    : go-utils\sonyflake_generator.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 通用雪花 ID 生成器封装
//

package utils

import (
	"errors"
	"sync"

	"github.com/sony/sonyflake"
)

// SonyflakeSettingsProvider 定义雪花生成器初始化时使用的配置提供函数.
// 返回值中的 MachineID 可以继续使用项目侧的配置或运行时环境计算.
type SonyflakeSettingsProvider func() (sonyflake.Settings, error)

// SonyflakeGenerator 封装延迟初始化且可复用的雪花 ID 生成器.
// 它会在首次生成 ID 时初始化一次底层实例, 后续调用复用同一个序列状态.
type SonyflakeGenerator struct {
	once             sync.Once
	settingsProvider SonyflakeSettingsProvider
	instance         *sonyflake.Sonyflake
	initErr          error
}

// NewSonyflakeGenerator 创建一个延迟初始化的雪花 ID 生成器.
// settingsProvider 会在首次调用 NextID 时执行, 用于读取配置并构造 sonyflake.Settings.
func NewSonyflakeGenerator(settingsProvider SonyflakeSettingsProvider) *SonyflakeGenerator {
	return &SonyflakeGenerator{settingsProvider: settingsProvider}
}

// NextID 生成一个全局唯一的雪花 ID.
// 当首次初始化失败时, 会返回初始化错误并保持失败状态, 避免后续继续使用无效实例.
func (generator *SonyflakeGenerator) NextID() (uint64, error) {
	if generator == nil {
		return 0, errors.New("sonyflake generator is nil")
	}

	generator.once.Do(generator.init)
	if generator.initErr != nil {
		return 0, generator.initErr
	}

	return generator.instance.NextID()
}

// init 初始化底层 sonyflake 实例.
func (generator *SonyflakeGenerator) init() {
	if generator.settingsProvider == nil {
		generator.initErr = errors.New("sonyflake settings provider is nil")
		return
	}

	settings, err := generator.settingsProvider()
	if err != nil {
		generator.initErr = err
		return
	}

	generator.instance = sonyflake.NewSonyflake(settings)
	if generator.instance == nil {
		generator.initErr = errors.New("sonyflake init failed")
	}
}
