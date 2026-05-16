//
// FilePath    : go-utils\logger\gorm.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : gorm zap 日志库，对 zapgorm2 做薄封装。支持通过 ErrorClassifier 自定义错误日志级别。
//

package logger

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	gormlogger "gorm.io/gorm/logger"
	"moul.io/zapgorm2"
)

// ErrorClassifier 判断错误应使用的日志级别。返回 0 表示使用默认级别（ERROR）。
type ErrorClassifier func(err error) gormlogger.LogLevel

// LoggerOption GORM logger 配置选项。
type LoggerOption func(*gormLoggerWrapper)

// WithErrorClassifier 注册错误分类器，用于调整特定错误的日志级别。
// 按注册顺序执行，首个返回非零值的分类器生效。
func WithErrorClassifier(fn ErrorClassifier) LoggerOption {
	return func(w *gormLoggerWrapper) {
		w.classifiers = append(w.classifiers, fn)
	}
}

// NewZapGormLogger 实例化 gorm logger，使用 zap 接收 gorm 日志。
// 可通过 opts 注入 ErrorClassifier，将预期内的错误降级（如唯一约束冲突 → WARN）。
func NewZapGormLogger(zapLogger *zap.Logger, logLevel gormlogger.LogLevel, slowThreshold time.Duration, opts ...LoggerOption) gormlogger.Interface {
	w := &gormLoggerWrapper{
		inner: zapgorm2.Logger{
			ZapLogger:                 zapLogger,
			LogLevel:                  logLevel,
			SlowThreshold:             slowThreshold * time.Millisecond,
			SkipCallerLookup:          false,
			IgnoreRecordNotFoundError: true,
			Context:                   nil,
		},
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

// gormLoggerWrapper 包装 zapgorm2.Logger，支持通过 ErrorClassifier 调整错误日志级别。
type gormLoggerWrapper struct {
	inner       zapgorm2.Logger
	classifiers []ErrorClassifier
}

func (w *gormLoggerWrapper) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	inner, ok := w.inner.LogMode(level).(zapgorm2.Logger)
	if !ok {
		inner = w.inner
	}
	return &gormLoggerWrapper{
		inner:       inner,
		classifiers: w.classifiers,
	}
}

func (w *gormLoggerWrapper) Info(ctx context.Context, msg string, data ...any) {
	w.inner.Info(ctx, msg, data...)
}

func (w *gormLoggerWrapper) Warn(ctx context.Context, msg string, data ...any) {
	w.inner.Warn(ctx, msg, data...)
}

func (w *gormLoggerWrapper) Error(ctx context.Context, msg string, data ...any) {
	level := w.classifyFromData(data...)
	w.logAt(ctx, level, msg, data...)
}

func (w *gormLoggerWrapper) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if err != nil {
		sql, rows := fc()
		level := w.classifyError(err)
		w.logAt(ctx, level, "trace", zap.Error(err), zap.Duration("elapsed", time.Since(begin)), zap.Int64("rows", rows), zap.String("sql", sql))
		return
	}
	w.inner.Trace(ctx, begin, fc, err)
}

// classifyError 对错误运行所有分类器，返回调整后的日志级别。若无匹配返回 0（默认 ERROR）。
func (w *gormLoggerWrapper) classifyError(err error) gormlogger.LogLevel {
	for _, c := range w.classifiers {
		if l := c(err); l != 0 {
			return l
		}
	}
	return 0
}

// classifyFromData 从 Error(...) 的 data 参数中提取 error 并分类。
// zapgorm2 传入 zap.Error(err)，即 zap.Field{Key: "error", Type: zapcore.ErrorType}。
func (w *gormLoggerWrapper) classifyFromData(data ...any) gormlogger.LogLevel {
	for _, d := range data {
		var err error
		switch v := d.(type) {
		case error:
			err = v
		case zap.Field:
			if v.Key == "error" {
				err = errors.New(v.String)
			}
		}
		if err != nil {
			return w.classifyError(err)
		}
	}
	return 0
}

// logAt 根据日志级别选择对应的 inner 方法。
func (w *gormLoggerWrapper) logAt(ctx context.Context, level gormlogger.LogLevel, msg string, data ...any) {
	switch level {
	case gormlogger.Warn:
		w.inner.Warn(ctx, msg, data...)
	case gormlogger.Info:
		w.inner.Info(ctx, msg, data...)
	case gormlogger.Silent:
		return
	default:
		w.inner.Error(ctx, msg, data...)
	}
}
