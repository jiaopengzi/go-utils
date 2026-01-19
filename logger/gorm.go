//
// FilePath    : go-utils\logger\gorm.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : gorm zap 日志库 zapgorm2 1.3.0 中方法 New() 参数化修改
//

package logger

import (
	"time"

	"go.uber.org/zap"
	gormlogger "gorm.io/gorm/logger"
	"moul.io/zapgorm2"
)

// NewZapGormLogger 实例化 zapgorm2.Logger,使用 zap 接收 gorm 日志
//   - zapLogger: zap logger
//   - logLevel: gorm 日志级别
//   - slowThreshold: 慢日志阈值 毫秒
func NewZapGormLogger(zapLogger *zap.Logger, logLevel gormlogger.LogLevel, slowThreshold time.Duration) zapgorm2.Logger {
	return zapgorm2.Logger{
		ZapLogger:                 zapLogger,                        // zap 日志实例
		LogLevel:                  logLevel,                         // 日志级别
		SlowThreshold:             slowThreshold * time.Millisecond, // 慢日志阈值
		SkipCallerLookup:          false,                            // 跳过调用者查找
		IgnoreRecordNotFoundError: true,                             // 忽略记录未找到错误 `record not found`
		Context:                   nil,
	}
}
